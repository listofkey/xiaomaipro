from __future__ import annotations

import asyncio
import json
import time
from collections.abc import AsyncIterator
from typing import Any

from openai import AsyncOpenAI

from agent.config import Settings
from agent.memory import ConversationStore
from agent.prompts import build_system_prompt
from agent.rag import TicketingKnowledgeBase
from agent.safety import TicketingSafetyGuard
from agent.schemas import (
    CardAction,
    ChatRequest,
    ChatResponse,
    Citation,
    ConversationState,
    RecentEvent,
    SlotExtraction,
    TicketCard,
    ToolTrace,
)
from agent.tools import ToolClient, summarize_search_result
from agent.utils import (
    extract_rank_reference,
    extract_slots_from_text,
    format_datetime_text,
    json_dumps,
    now_in_timezone,
    resolve_date_range,
    sse_event,
    truncate,
)


DETAIL_SALE_KEYWORDS = ("开票", "开售", "开抢", "几点开", "什么时候开")
DETAIL_REAL_NAME_KEYWORDS = ("实名", "实名制")
DETAIL_STOCK_KEYWORDS = ("余票", "票档", "票价", "库存", "还有什么票", "还有哪些票")
DETAIL_VENUE_KEYWORDS = ("场馆", "地址", "在哪", "地点")
DETAIL_LIMIT_KEYWORDS = ("限购", "最多买", "最多能买", "最多几张")
DETAIL_TICKET_TYPE_KEYWORDS = ("电子票", "纸质票", "票类型")
DETAIL_GENERAL_KEYWORDS = ("详情", "详细", "这场", "那场", "这个活动", "这个演出")
RULE_KNOWLEDGE_KEYWORDS = ("实名", "退票", "入场", "规则", "须知")
DEFAULT_EVENT_REFERENCE_KEYWORDS = ("这场", "那场", "这个", "刚才", "上面", "上一场")


def build_tool_schemas() -> list[dict[str, Any]]:
    return [
        {
            "type": "function",
            "function": {
                "name": "query_ticket_stock",
                "description": "查询演出活动、票档、余票、城市和时间信息。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "keyword": {"type": "string", "description": "艺人或演出关键词"},
                        "city": {"type": "string", "description": "查询城市"},
                        "date_hint": {"type": "string", "description": "相对时间，如下周六"},
                        "start_date": {"type": "string", "description": "开始日期，YYYY-MM-DD"},
                        "end_date": {"type": "string", "description": "结束日期，YYYY-MM-DD"},
                        "quantity": {"type": "integer", "description": "张数"},
                        "budget_total": {"type": "number", "description": "总预算"},
                        "budget_per_ticket": {"type": "number", "description": "单张预算"},
                        "area_preference": {"type": "string", "description": "区域偏好，如内场、看台前排"},
                        "page_size": {"type": "integer", "description": "最多返回场次数量"},
                    },
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "query_event_detail",
                "description": "查询具体活动详情、售票规则和票档。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "event_id": {"type": "string", "description": "活动 ID"},
                    },
                    "required": ["event_id"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "search_ticket_policy",
                "description": "查询实名、退票、入场、开票等规则知识库。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "query": {"type": "string", "description": "规则查询问题"},
                        "top_k": {"type": "integer", "description": "返回片段数"},
                    },
                    "required": ["query"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "recommend_hot_events",
                "description": "当无结果或用户想看推荐时，查询热门演出。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "city": {"type": "string", "description": "城市"},
                        "limit": {"type": "integer", "description": "数量"},
                    },
                },
            },
        },
    ]


class TicketingAgentService:
    def __init__(
        self,
        settings: Settings,
        store: ConversationStore,
        tool_client: ToolClient,
        knowledge_base: TicketingKnowledgeBase,
        safety_guard: TicketingSafetyGuard,
    ) -> None:
        self._settings = settings
        self._store = store
        self._tool_client = tool_client
        self._knowledge_base = knowledge_base
        self._safety_guard = safety_guard
        self._tools = build_tool_schemas()
        self._openai: AsyncOpenAI | None = None
        if settings.llm_enabled:
            self._openai = AsyncOpenAI(
                api_key=settings.openai_api_key,
                base_url=settings.openai_base_url,
                timeout=settings.openai_timeout_seconds,
            )

    async def bootstrap(self) -> int:
        return await self._knowledge_base.bootstrap()

    async def chat(self, request: ChatRequest) -> ChatResponse:
        return await self._run_chat(request)

    async def stream(self, request: ChatRequest) -> AsyncIterator[bytes]:
        yield sse_event("start", {"sessionId": request.session_id})
        response = await self._run_chat(request)

        yield sse_event(
            "meta",
            {
                "sessionId": response.session_id,
                "fallbackMode": response.fallback_mode,
                "slots": response.slots.model_dump(mode="json"),
            },
        )
        for tool in response.tools:
            yield sse_event("tool", tool.model_dump(mode="json"))
        if response.cards:
            yield sse_event("cards", {"cards": [card.model_dump(mode="json") for card in response.cards]})
        for citation in response.citations:
            yield sse_event("citation", citation.model_dump(mode="json"))

        for chunk in self._chunk_answer(response.answer):
            yield sse_event("token", {"content": chunk})
            await asyncio.sleep(0.01)

        yield sse_event("done", response.model_dump(mode="json"))

    async def _run_chat(self, request: ChatRequest) -> ChatResponse:
        state = await self._store.load(request.session_id)
        decision = self._safety_guard.inspect(request.message, has_context=bool(state.messages))
        slots = state.slots.merge(extract_slots_from_text(request.message))

        if not decision.allowed:
            response = ChatResponse(
                session_id=request.session_id,
                answer=decision.response or "我可以帮你查询票务相关问题。",
                slots=slots,
                fallback_mode=True,
                recent_events=state.recent_events,
            )
            await self._store.append_exchange(
                request.session_id,
                request.message,
                response.answer,
                slots,
                response.recent_events,
            )
            return response

        if self._openai is None:
            response = await self._rule_based_answer(request, state, slots, fallback_mode=True)
            await self._store.append_exchange(
                request.session_id,
                request.message,
                response.answer,
                response.slots,
                response.recent_events,
            )
            return response

        try:
            response = await self._llm_answer(request, state, slots)
        except Exception:
            if not self._settings.allow_mock_fallback:
                raise
            response = await self._rule_based_answer(request, state, slots, fallback_mode=True)

        await self._store.append_exchange(
            request.session_id,
            request.message,
            response.answer,
            response.slots,
            response.recent_events,
        )
        return response

    async def _llm_answer(
        self,
        request: ChatRequest,
        state: ConversationState,
        slots: SlotExtraction,
    ) -> ChatResponse:
        assert self._openai is not None
        messages = self._build_messages(state, request.message, slots)
        tool_traces: list[ToolTrace] = []
        citations: list[Citation] = []
        cards: list[TicketCard] = []
        resolved_slots = slots
        resolved_recent_events = list(state.recent_events)

        for _ in range(self._settings.answer_max_tool_rounds):
            completion = await self._openai.chat.completions.create(
                model=self._settings.openai_chat_model,
                messages=messages,
                tools=self._tools,
                tool_choice="auto",
                parallel_tool_calls=False,
                temperature=0.2,
                max_completion_tokens=self._settings.answer_max_tokens,
                user=request.user_id or request.session_id,
            )
            choice = completion.choices[0].message
            tool_calls = getattr(choice, "tool_calls", None) or []

            if not tool_calls:
                answer = (choice.content or "").strip()
                if not answer:
                    answer = self._compose_search_answer(None, citations)
                return ChatResponse(
                    session_id=request.session_id,
                    answer=answer,
                    cards=cards,
                    citations=citations,
                    tools=tool_traces,
                    slots=resolved_slots,
                    recent_events=resolved_recent_events,
                )

            messages.append(
                {
                    "role": "assistant",
                    "content": choice.content or "",
                    "tool_calls": [
                        {
                            "id": tool_call.id,
                            "type": "function",
                            "function": {
                                "name": tool_call.function.name,
                                "arguments": tool_call.function.arguments,
                            },
                        }
                        for tool_call in tool_calls
                    ],
                }
            )
            for tool_call in tool_calls:
                result, trace, new_citations, new_cards, new_slots, new_recent_events = await self._execute_tool(
                    tool_call.function.name,
                    tool_call.function.arguments,
                    request,
                    resolved_slots,
                )
                tool_traces.append(trace)
                citations.extend(new_citations)
                cards = self._merge_cards(cards, new_cards)
                resolved_slots = resolved_slots.merge(new_slots)
                resolved_recent_events = self._merge_recent_events(resolved_recent_events, new_recent_events)
                messages.append(
                    {
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": json_dumps(result),
                    }
                )

        fallback = await self._rule_based_answer(request, state, resolved_slots, fallback_mode=True)
        fallback.tools = tool_traces
        fallback.citations = citations or fallback.citations
        fallback.cards = self._merge_cards(cards, fallback.cards)
        fallback.slots = resolved_slots
        fallback.recent_events = self._merge_recent_events(resolved_recent_events, fallback.recent_events)
        return fallback

    async def _execute_tool(
        self,
        name: str,
        raw_arguments: str,
        request: ChatRequest,
        slots: SlotExtraction,
    ) -> tuple[dict[str, Any], ToolTrace, list[Citation], list[TicketCard], SlotExtraction, list[RecentEvent]]:
        started_at = time.perf_counter()
        try:
            arguments = json.loads(raw_arguments or "{}")
        except json.JSONDecodeError:
            arguments = {}

        citations: list[Citation] = []
        cards: list[TicketCard] = []
        slot_delta = SlotExtraction()
        recent_events: list[RecentEvent] = []

        if name == "query_ticket_stock":
            slot_delta = SlotExtraction(
                artist=arguments.get("keyword"),
                keyword=arguments.get("keyword"),
                city=arguments.get("city"),
                date_hint=arguments.get("date_hint"),
                quantity=arguments.get("quantity"),
                budget_total=arguments.get("budget_total"),
                budget_per_ticket=arguments.get("budget_per_ticket"),
                area_preference=arguments.get("area_preference"),
            )
            merged = slots.merge(slot_delta)
            result = await self._search_ticket_stock(
                request,
                merged,
                keyword=arguments.get("keyword"),
                city=arguments.get("city"),
                date_hint=arguments.get("date_hint"),
                start_date=arguments.get("start_date"),
                end_date=arguments.get("end_date"),
                quantity=arguments.get("quantity"),
                budget_total=arguments.get("budget_total"),
                budget_per_ticket=arguments.get("budget_per_ticket"),
                area_preference=arguments.get("area_preference"),
                page_size=min(int(arguments.get("page_size", 5) or 5), 8),
            )
            cards = self._build_event_cards(result.get("events", []))
            recent_events = self._recent_events_from_search_items(result.get("events", []))
            summary = summarize_search_result(result)
            trace = ToolTrace(
                name=name,
                arguments=arguments,
                summary=summary,
                latency_ms=int((time.perf_counter() - started_at) * 1000),
            )
            return result, trace, citations, cards, merged, recent_events

        if name == "query_event_detail":
            result = await self._tool_client.get_event_detail(
                str(arguments.get("event_id", "")),
                request.access_token,
            )
            cards = self._build_detail_cards([result])
            recent_events = self._recent_events_from_event_payload(result)
            trace = ToolTrace(
                name=name,
                arguments=arguments,
                summary=truncate(result.get("description", "") or result.get("title", ""), 120),
                latency_ms=int((time.perf_counter() - started_at) * 1000),
            )
            return result, trace, citations, cards, slot_delta, recent_events

        if name == "recommend_hot_events":
            result = await self._tool_client.get_hot_recommendations(
                city=arguments.get("city") or slots.city,
                limit=min(int(arguments.get("limit", 4) or 4), 6),
                access_token=request.access_token,
            )
            cards = self._build_recommend_cards(result.get("events", []))
            recent_events = self._recent_events_from_event_briefs(result.get("events", []))
            trace = ToolTrace(
                name=name,
                arguments=arguments,
                summary=f'推荐了 {result.get("total", 0)} 个热门场次',
                latency_ms=int((time.perf_counter() - started_at) * 1000),
            )
            return result, trace, citations, cards, slot_delta, recent_events

        if name == "search_ticket_policy":
            context, citations = await self._knowledge_base.search(
                arguments.get("query", request.message),
                int(arguments.get("top_k") or request.top_k or self._settings.knowledge_default_top_k),
            )
            result = {"context": context, "citations": [item.model_dump(mode="json") for item in citations]}
            trace = ToolTrace(
                name=name,
                arguments=arguments,
                summary=truncate(context or "知识库暂无结果", 120),
                latency_ms=int((time.perf_counter() - started_at) * 1000),
            )
            return result, trace, citations, cards, slot_delta, recent_events

        trace = ToolTrace(
            name=name,
            arguments=arguments,
            status="error",
            summary="未识别的工具",
            latency_ms=int((time.perf_counter() - started_at) * 1000),
        )
        return {"error": "unknown_tool"}, trace, citations, cards, slot_delta, recent_events

    async def _rule_based_answer(
        self,
        request: ChatRequest,
        state: ConversationState,
        slots: SlotExtraction,
        *,
        fallback_mode: bool,
    ) -> ChatResponse:
        merged_slots = state.slots.merge(slots)
        citations: list[Citation] = []
        cards: list[TicketCard] = []
        tool_traces: list[ToolTrace] = []
        recent_events = list(state.recent_events)
        message = request.message

        detail_focus = self._detect_event_detail_focus(message)
        referenced_event, referenced_rank, explicit_reference = self._resolve_recent_event(
            message,
            recent_events,
            prefer_first=bool(detail_focus),
        )

        if detail_focus and referenced_event:
            detail_started_at = time.perf_counter()
            result = await self._tool_client.get_event_detail(referenced_event.event_id, request.access_token)
            cards = self._build_detail_cards([result])
            recent_events = self._merge_recent_events(recent_events, self._recent_events_from_event_payload(result))
            tool_traces.append(
                ToolTrace(
                    name="query_event_detail",
                    arguments={"event_id": referenced_event.event_id},
                    summary=truncate(result.get("description", "") or result.get("title", ""), 120),
                    latency_ms=int((time.perf_counter() - detail_started_at) * 1000),
                )
            )
            answer = self._compose_event_detail_answer(
                result,
                focus=detail_focus,
                referenced_rank=referenced_rank,
                assumed=not explicit_reference and len(state.recent_events) > 1,
            )
            return ChatResponse(
                session_id=request.session_id,
                answer=answer,
                cards=cards,
                citations=citations,
                tools=tool_traces,
                slots=merged_slots,
                fallback_mode=fallback_mode,
                recent_events=recent_events,
            )

        if detail_focus and merged_slots.keyword:
            search_started_at = time.perf_counter()
            candidate_result = await self._search_ticket_stock(request, merged_slots, page_size=3)
            tool_traces.append(
                ToolTrace(
                    name="query_ticket_stock",
                    summary=summarize_search_result(candidate_result),
                    latency_ms=int((time.perf_counter() - search_started_at) * 1000),
                )
            )
            candidate_events = self._recent_events_from_search_items(candidate_result.get("events", []))
            recent_events = self._merge_recent_events(recent_events, candidate_events)
            if candidate_events:
                result = await self._tool_client.get_event_detail(candidate_events[0].event_id, request.access_token)
                cards = self._build_detail_cards([result])
                recent_events = self._merge_recent_events(recent_events, self._recent_events_from_event_payload(result))
                tool_traces.append(
                    ToolTrace(
                        name="query_event_detail",
                        arguments={"event_id": candidate_events[0].event_id},
                        summary=truncate(result.get("description", "") or result.get("title", ""), 120),
                    )
                )
                answer = self._compose_event_detail_answer(
                    result,
                    focus=detail_focus,
                    referenced_rank=1,
                    assumed=True,
                )
                return ChatResponse(
                    session_id=request.session_id,
                    answer=answer,
                    cards=cards,
                    citations=citations,
                    tools=tool_traces,
                    slots=merged_slots,
                    fallback_mode=fallback_mode,
                    recent_events=recent_events,
                )

        if any(keyword in message for keyword in RULE_KNOWLEDGE_KEYWORDS):
            context, citations = await self._knowledge_base.search(
                message,
                request.top_k or self._settings.knowledge_default_top_k,
            )
            answer = self._compose_rule_answer(message, citations, context)
            if not answer:
                answer = "我先给你结论：这类规则建议以下单页和活动页展示为准，当前知识库里还没有命中到更具体的说明。"
            return ChatResponse(
                session_id=request.session_id,
                answer=answer,
                cards=cards,
                citations=citations,
                tools=tool_traces,
                slots=merged_slots,
                fallback_mode=fallback_mode,
                recent_events=recent_events,
            )

        if any(keyword in message for keyword in ("推荐", "热门", "最近有什么")) and not merged_slots.keyword:
            result = await self._tool_client.get_hot_recommendations(
                city=merged_slots.city,
                limit=4,
                access_token=request.access_token,
            )
            cards = self._build_recommend_cards(result.get("events", []))
            recent_events = self._merge_recent_events(recent_events, self._recent_events_from_event_briefs(result.get("events", [])))
            tool_traces.append(
                ToolTrace(name="recommend_hot_events", summary=f'推荐了 {result.get("total", 0)} 个热门场次')
            )
            answer = self._compose_recommendation_answer(result)
            return ChatResponse(
                session_id=request.session_id,
                answer=answer,
                cards=cards,
                citations=citations,
                tools=tool_traces,
                slots=merged_slots,
                fallback_mode=fallback_mode,
                recent_events=recent_events,
            )

        search_started_at = time.perf_counter()
        result = await self._search_ticket_stock(request, merged_slots, page_size=4)
        cards = self._build_event_cards(result.get("events", []))
        recent_events = self._merge_recent_events(recent_events, self._recent_events_from_search_items(result.get("events", [])))
        tool_traces.append(
            ToolTrace(
                name="query_ticket_stock",
                summary=summarize_search_result(result),
                latency_ms=int((time.perf_counter() - search_started_at) * 1000),
            )
        )
        answer = self._compose_search_answer(result, citations)

        if not result.get("events"):
            hot_result = await self._tool_client.get_hot_recommendations(
                city=merged_slots.city,
                limit=3,
                access_token=request.access_token,
            )
            cards = self._merge_cards(cards, self._build_recommend_cards(hot_result.get("events", [])))
            recent_events = self._merge_recent_events(recent_events, self._recent_events_from_event_briefs(hot_result.get("events", [])))
            answer += "\n\n我顺手给你补了几个热门场次，可以继续看看是否要换日期、换城市或改看台区域。"

        return ChatResponse(
            session_id=request.session_id,
            answer=answer,
            cards=cards,
            citations=citations,
            tools=tool_traces,
            slots=merged_slots,
            fallback_mode=fallback_mode,
            recent_events=recent_events,
        )

    async def _search_ticket_stock(
        self,
        request: ChatRequest,
        slots: SlotExtraction,
        *,
        keyword: str | None = None,
        city: str | None = None,
        date_hint: str | None = None,
        start_date: str | None = None,
        end_date: str | None = None,
        quantity: int | None = None,
        budget_total: float | None = None,
        budget_per_ticket: float | None = None,
        area_preference: str | None = None,
        page_size: int = 4,
    ) -> dict[str, Any]:
        resolved_start_date = start_date
        resolved_end_date = end_date
        if not resolved_start_date and not resolved_end_date:
            resolved_start_date, resolved_end_date = resolve_date_range(
                date_hint or slots.date_hint,
                self._settings.app_timezone,
                self._settings.app_reference_date,
            )

        resolved_quantity = quantity or slots.quantity
        resolved_budget_total = budget_total or slots.budget_total
        resolved_budget_per_ticket = budget_per_ticket or slots.budget_per_ticket
        if resolved_budget_per_ticket is None and resolved_budget_total and resolved_quantity:
            resolved_budget_per_ticket = round(float(resolved_budget_total) / max(resolved_quantity, 1), 2)

        return await self._tool_client.search_ticket_stock(
            keyword=keyword or slots.keyword,
            city=city or slots.city,
            start_date=resolved_start_date,
            end_date=resolved_end_date,
            quantity=resolved_quantity,
            max_price_per_ticket=resolved_budget_per_ticket,
            area_preference=area_preference or slots.area_preference,
            page_size=page_size,
            access_token=request.access_token,
        )

    def _build_messages(
        self,
        state: ConversationState,
        user_message: str,
        slots: SlotExtraction,
    ) -> list[dict[str, Any]]:
        now = now_in_timezone(self._settings.app_timezone)
        messages: list[dict[str, Any]] = [
            {
                "role": "system",
                "content": build_system_prompt(now.strftime("%Y-%m-%d"), self._settings.app_timezone),
            }
        ]
        if slots.model_dump(exclude_none=True):
            messages.append(
                {
                    "role": "system",
                    "content": f"当前会话已知槽位: {json_dumps(slots.model_dump(exclude_none=True))}",
                }
            )
        if state.recent_events:
            recent_lines = ["最近命中的场次："]
            for index, event in enumerate(state.recent_events[:5], start=1):
                parts = [event.city or "", format_datetime_text(event.event_start_time)]
                meta = " | ".join(part for part in parts if part and part != "待定")
                suffix = f" | {meta}" if meta else ""
                recent_lines.append(f"{index}. [{event.event_id}] {event.title}{suffix}")
            messages.append({"role": "system", "content": "\n".join(recent_lines)})
        for item in state.messages[-max(self._settings.session_history_limit * 2, 2) :]:
            messages.append({"role": item.role, "content": item.content})
        messages.append({"role": "user", "content": user_message})
        return messages

    def _build_event_cards(self, items: list[dict[str, Any]]) -> list[TicketCard]:
        cards: list[TicketCard] = []
        for item in items[:3]:
            event = item.get("event", {})
            tiers = item.get("matched_tiers", [])
            body = item.get("match_reason", "")
            if tiers:
                tier_lines = [f'{tier["name"]}: {tier["price"]:.0f}元, 余票{tier["remain_stock"]}' for tier in tiers[:2]]
                body = f"{body}；" + "；".join(tier_lines)
            cards.append(
                TicketCard(
                    title=event.get("title", ""),
                    subtitle=f'{event.get("city", "")} | {event.get("venue_name", "")} | {format_datetime_text(event.get("event_start_time"))}',
                    body=body,
                    event_id=event.get("id"),
                    tags=[tag for tag in [event.get("artist"), event.get("city")] if tag],
                    actions=[
                        CardAction(
                            label="查看详情",
                            action_type="detail",
                            payload={"eventId": event.get("id")},
                        ),
                        CardAction(
                            label="去购票页",
                            action_type="buy",
                            payload={"eventId": event.get("id")},
                        ),
                    ],
                    metadata={"matchedTiers": tiers},
                )
            )
        return cards

    def _build_detail_cards(self, items: list[dict[str, Any]]) -> list[TicketCard]:
        cards: list[TicketCard] = []
        for event in items[:2]:
            tiers = event.get("ticket_tiers", [])
            body = "；".join(
                f'{tier["name"]} {tier["price"]:.0f}元 余票{tier["remain_stock"]}'
                for tier in tiers[:3]
            )
            cards.append(
                TicketCard(
                    title=event.get("title", ""),
                    subtitle=f'{event.get("city", "")} | {event.get("venue_name", "")} | {format_datetime_text(event.get("event_start_time"))}',
                    body=body or truncate(event.get("description", ""), 120),
                    event_id=event.get("id"),
                    tags=["活动详情"],
                    actions=[
                        CardAction(
                            label="查看详情",
                            action_type="detail",
                            payload={"eventId": event.get("id")},
                        )
                    ],
                )
            )
        return cards

    def _build_recommend_cards(self, items: list[dict[str, Any]]) -> list[TicketCard]:
        cards: list[TicketCard] = []
        for event in items[:3]:
            cards.append(
                TicketCard(
                    title=event.get("title", ""),
                    subtitle=f'{event.get("city", "")} | {event.get("venue_name", "")} | {format_datetime_text(event.get("event_start_time"))}',
                    body=f'最低票价 {event.get("min_price", 0):.0f} 元',
                    event_id=event.get("id"),
                    tags=["热门推荐"],
                    actions=[
                        CardAction(
                            label="查看详情",
                            action_type="detail",
                            payload={"eventId": event.get("id")},
                        )
                    ],
                )
            )
        return cards

    def _merge_cards(self, existing: list[TicketCard], incoming: list[TicketCard]) -> list[TicketCard]:
        seen = {card.event_id or card.title for card in existing}
        merged = list(existing)
        for card in incoming:
            key = card.event_id or card.title
            if key in seen:
                continue
            seen.add(key)
            merged.append(card)
        return merged

    def _compose_search_answer(self, result: dict[str, Any] | None, citations: list[Citation]) -> str:
        if not result or not result.get("events"):
            suffix = ""
            if citations:
                suffix = f" 另外我命中了 {len(citations)} 条规则知识，你也可以继续问我实名、退票或入场问题。"
            return "暂时没找到完全符合条件的场次。你可以换个日期、城市，或者把区域从内场放宽到看台前排试试。" + suffix

        lines = ["我先帮你看了一下，当前比较匹配的场次有："]
        for index, item in enumerate(result["events"][:3], start=1):
            event = item["event"]
            tiers = item.get("matched_tiers") or []
            reason = item.get("match_reason", "")
            if tiers:
                tier_text = "；".join(
                    f'{tier["name"]} {tier["price"]:.0f}元，余票{tier["remain_stock"]}'
                    for tier in tiers[:2]
                )
                if reason:
                    tier_text = f"{reason}；{tier_text}"
            else:
                tier_text = reason or "暂无完全命中的可售票档"
            lines.append(
                f'{index}. {event["title"]}，{event["city"]}，{format_datetime_text(event.get("event_start_time"))}，{tier_text}'
            )
        lines.append("如果你愿意，我可以继续按“更低预算 / 其他区域 / 同艺人其他城市”帮你缩小范围。")
        return "\n".join(lines)

    def _compose_rule_answer(self, question: str, citations: list[Citation], context: str) -> str:
        if not citations:
            return ""
        intro = "我根据当前知识库整理了一下："
        bullets = [f"{index}. {citation.snippet}" for index, citation in enumerate(citations[:3], start=1)]
        source_line = "来源：" + "、".join(citation.source for citation in citations[:3])
        if "退票" in question:
            intro = "关于退票，我先给你知识库里的结论："
        elif "实名" in question:
            intro = "关于实名制要求，我查到这些要点："
        elif "入场" in question:
            intro = "关于入场须知，我查到这些信息："
        return "\n".join([intro, *bullets, source_line])

    def _compose_event_detail_answer(
        self,
        event: dict[str, Any],
        *,
        focus: str,
        referenced_rank: int | None,
        assumed: bool,
    ) -> str:
        title = event.get("title", "这场演出")
        prefix = f"关于“{title}”，"
        if assumed and referenced_rank:
            prefix = f"如果你问的是刚才第{referenced_rank}场“{title}”，"

        sale_time = format_datetime_text(event.get("sale_start_time"))
        event_time = format_datetime_text(event.get("event_start_time"))
        venue_name = event.get("venue_name") or "场馆待定"
        venue_address = event.get("venue_address") or "地址待补充"
        purchase_limit = event.get("purchase_limit")
        real_name_required = event.get("need_real_name") == 1
        real_name_text = "需要实名制购票和入场" if real_name_required else "当前未标注需要实名制要求"
        purchase_limit_text = f"每人限购 {purchase_limit} 张" if purchase_limit else "暂未看到限购说明"
        ticket_type_text = self._ticket_type_text(event.get("ticket_type"))
        tier_summary = self._summarize_ticket_tiers(event.get("ticket_tiers", []))

        if focus == "sale_time":
            return "\n".join(
                [
                    f"{prefix}开售时间是 {sale_time}。",
                    f"演出时间是 {event_time}，{real_name_text}，{purchase_limit_text}。",
                ]
            )
        if focus == "real_name":
            return "\n".join(
                [
                    f"{prefix}{real_name_text}。",
                    f"演出时间是 {event_time}，开售时间 {sale_time}，{purchase_limit_text}。",
                ]
            )
        if focus == "venue":
            return "\n".join(
                [
                    f"{prefix}场馆是 {venue_name}，地址是 {venue_address}。",
                    f"演出时间 {event_time}，开售时间 {sale_time}。",
                ]
            )
        if focus == "purchase_limit":
            return "\n".join(
                [
                    f"{prefix}{purchase_limit_text}。",
                    f"{real_name_text}，开售时间 {sale_time}。",
                ]
            )
        if focus == "ticket_type":
            return "\n".join(
                [
                    f"{prefix}当前票务类型是 {ticket_type_text}。",
                    f"{real_name_text}，{purchase_limit_text}。",
                ]
            )
        if focus == "stock":
            return "\n".join(
                [
                    f"{prefix}当前主要票档是：{tier_summary}。",
                    f"演出时间 {event_time}，开售时间 {sale_time}。",
                ]
            )

        return "\n".join(
            [
                f"{prefix}演出时间是 {event_time}，开售时间 {sale_time}。",
                f"{real_name_text}，{purchase_limit_text}，票务类型 {ticket_type_text}。",
                f"当前主要票档：{tier_summary}。",
            ]
        )

    def _compose_recommendation_answer(self, result: dict[str, Any]) -> str:
        events = result.get("events", [])
        if not events:
            return "当前没有拿到热门推荐结果。你可以直接告诉我想看的城市、艺人或预算，我继续帮你查。"
        lines = ["你可以先看看这些热门场次："]
        for index, event in enumerate(events[:3], start=1):
            lines.append(
                f'{index}. {event["title"]}，{event["city"]}，{format_datetime_text(event.get("event_start_time"))}，最低 {event.get("min_price", 0):.0f} 元'
            )
        return "\n".join(lines)

    def _detect_event_detail_focus(self, text: str) -> str | None:
        if any(keyword in text for keyword in ("退票规则", "入场规则", "实名规则", "购票规则", "规则", "须知")):
            return None
        if any(keyword in text for keyword in DETAIL_SALE_KEYWORDS):
            return "sale_time"
        if any(keyword in text for keyword in DETAIL_REAL_NAME_KEYWORDS):
            return "real_name"
        if any(keyword in text for keyword in DETAIL_STOCK_KEYWORDS):
            return "stock"
        if any(keyword in text for keyword in DETAIL_VENUE_KEYWORDS):
            return "venue"
        if any(keyword in text for keyword in DETAIL_LIMIT_KEYWORDS):
            return "purchase_limit"
        if any(keyword in text for keyword in DETAIL_TICKET_TYPE_KEYWORDS):
            return "ticket_type"
        if any(keyword in text for keyword in DETAIL_GENERAL_KEYWORDS):
            return "general"
        return None

    def _resolve_recent_event(
        self,
        text: str,
        recent_events: list[RecentEvent],
        *,
        prefer_first: bool,
    ) -> tuple[RecentEvent | None, int | None, bool]:
        if not recent_events:
            return None, None, False

        rank = extract_rank_reference(text)
        if rank:
            index = len(recent_events) - 1 if rank < 0 else rank - 1
            if 0 <= index < len(recent_events):
                return recent_events[index], index + 1, True

        for index, event in enumerate(recent_events):
            searchable_tokens = [event.event_id, event.title, event.city, event.artist, event.venue_name]
            if any(token and token in text for token in searchable_tokens):
                return event, index + 1, True

        if len(recent_events) == 1:
            return recent_events[0], 1, False

        if any(keyword in text for keyword in DEFAULT_EVENT_REFERENCE_KEYWORDS):
            return recent_events[0], 1, False

        if prefer_first and len(text.strip()) <= 12:
            return recent_events[0], 1, False

        return None, None, False

    def _recent_events_from_search_items(self, items: list[dict[str, Any]]) -> list[RecentEvent]:
        return [self._to_recent_event(item.get("event", {})) for item in items[:5] if item.get("event")]

    def _recent_events_from_event_briefs(self, items: list[dict[str, Any]]) -> list[RecentEvent]:
        return [self._to_recent_event(item) for item in items[:5]]

    def _recent_events_from_event_payload(self, payload: dict[str, Any]) -> list[RecentEvent]:
        if not payload:
            return []
        return [self._to_recent_event(payload)]

    def _to_recent_event(self, payload: dict[str, Any]) -> RecentEvent:
        return RecentEvent(
            event_id=str(payload.get("id", "")),
            title=payload.get("title", ""),
            city=payload.get("city"),
            artist=payload.get("artist"),
            venue_name=payload.get("venue_name"),
            event_start_time=payload.get("event_start_time"),
        )

    def _merge_recent_events(
        self,
        existing: list[RecentEvent],
        incoming: list[RecentEvent],
    ) -> list[RecentEvent]:
        merged: list[RecentEvent] = []
        seen: set[str] = set()
        for event in [*incoming, *existing]:
            if not event.event_id or event.event_id in seen:
                continue
            seen.add(event.event_id)
            merged.append(event)
        return merged[:5]

    def _summarize_ticket_tiers(self, tiers: list[dict[str, Any]]) -> str:
        if not tiers:
            return "暂未拿到票档信息"
        return "；".join(
            f'{tier["name"]} {tier["price"]:.0f}元(余票{tier["remain_stock"]})'
            for tier in tiers[:4]
        )

    def _ticket_type_text(self, ticket_type: int | None) -> str:
        if ticket_type == 1:
            return "电子票"
        if ticket_type == 2:
            return "纸质票"
        return "待定"

    def _chunk_answer(self, answer: str) -> list[str]:
        if len(answer) <= 32:
            return [answer]
        chunks: list[str] = []
        start = 0
        while start < len(answer):
            end = min(start + 28, len(answer))
            chunks.append(answer[start:end])
            start = end
        return chunks

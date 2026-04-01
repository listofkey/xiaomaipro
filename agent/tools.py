from __future__ import annotations

import asyncio
from dataclasses import dataclass
from typing import Any, Protocol

import httpx
from pydantic import BaseModel, Field

from agent.config import Settings
from agent.utils import truncate


class TicketTier(BaseModel):
    id: str
    event_id: str
    name: str
    price: float
    remain_stock: int
    status: int = 0


class EventSummary(BaseModel):
    id: str
    title: str
    city: str
    artist: str
    venue_name: str
    event_start_time: str
    event_end_time: str | None = None
    min_price: float | None = None
    status: int = 0
    poster_url: str | None = None
    is_hot: bool = False


class EventDetail(EventSummary):
    description: str = ""
    sale_start_time: str | None = None
    sale_end_time: str | None = None
    venue_address: str | None = None
    purchase_limit: int | None = None
    need_real_name: int | None = None
    ticket_type: int | None = None
    ticket_tiers: list[TicketTier] = Field(default_factory=list)


class ToolClient(Protocol):
    async def search_ticket_stock(
        self,
        *,
        keyword: str | None,
        city: str | None,
        start_date: str | None,
        end_date: str | None,
        quantity: int | None,
        max_price_per_ticket: float | None,
        area_preference: str | None,
        page_size: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        ...

    async def get_event_detail(self, event_id: str, access_token: str | None = None) -> dict[str, Any]:
        ...

    async def get_hot_recommendations(
        self,
        *,
        city: str | None,
        limit: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        ...


def _match_tiers(
    detail: EventDetail,
    *,
    quantity: int | None,
    max_price_per_ticket: float | None,
    area_preference: str | None,
) -> tuple[list[TicketTier], str]:
    available = [tier for tier in detail.ticket_tiers if tier.remain_stock > 0]
    if not available:
        return [], "当前可售票档为空"

    quantity_filtered = available
    quantity_shortage = False
    if quantity:
        enough_quantity = [tier for tier in available if tier.remain_stock >= quantity]
        if enough_quantity:
            quantity_filtered = enough_quantity
        else:
            quantity_shortage = True

    area_filtered = quantity_filtered
    area_sold_out = False
    area_quantity_shortage = False
    area_not_found = False
    if area_preference:
        keyword = area_preference.lower()
        exact_area = [tier for tier in quantity_filtered if keyword in tier.name.lower()]
        if exact_area:
            area_filtered = exact_area
        else:
            area_exists = any(keyword in tier.name.lower() for tier in detail.ticket_tiers)
            area_available = any(keyword in tier.name.lower() for tier in available)
            if area_exists and area_available:
                area_quantity_shortage = quantity_shortage
            elif area_exists:
                area_sold_out = True
            else:
                area_not_found = True

    price_filtered = area_filtered
    if max_price_per_ticket is not None:
        exact_price = [tier for tier in area_filtered if tier.price <= max_price_per_ticket]
        if exact_price:
            price_filtered = exact_price

    if price_filtered:
        if quantity_shortage and price_filtered == available:
            return sorted(price_filtered, key=lambda item: item.price), f"暂时没有满足 {quantity} 张同档余票的票档，以下是当前可售备选"
        if area_quantity_shortage and price_filtered == available:
            return sorted(price_filtered, key=lambda item: item.price), f"{area_preference}区域暂时没有满足 {quantity} 张同档余票的票档，以下是其他可售区域备选"
        if area_sold_out and price_filtered == available:
            return sorted(price_filtered, key=lambda item: item.price), f"{area_preference}票已售罄，以下是其他可售区域备选"
        return sorted(price_filtered, key=lambda item: item.price), "命中可售票档"

    if area_preference and area_filtered != available:
        return sorted(area_filtered, key=lambda item: item.price), "目标区域无符合预算的票档，以下是同区域备选"

    if quantity_shortage:
        return sorted(available, key=lambda item: item.price)[:3], f"暂时没有满足 {quantity} 张同档余票的票档，以下是当前可售备选"

    if area_quantity_shortage:
        return sorted(available, key=lambda item: item.price)[:3], f"{area_preference}区域暂时没有满足 {quantity} 张同档余票的票档，以下是其他可售区域备选"

    if area_sold_out:
        return sorted(available, key=lambda item: item.price)[:3], f"{area_preference}票已售罄，以下是其他可售区域备选"

    if area_not_found:
        return sorted(available, key=lambda item: item.price)[:3], f"没有完全匹配“{area_preference}”的票档，以下是当前可售备选"

    if max_price_per_ticket is not None:
        return sorted(available, key=lambda item: item.price)[:3], "没有命中预算票档，以下是最接近的可售票档"

    return sorted(available, key=lambda item: item.price), "以下是当前可售票档"


def _build_ticket_search_payload(
    items: list[EventDetail],
    *,
    keyword: str | None,
    city: str | None,
    start_date: str | None,
    end_date: str | None,
    quantity: int | None,
    max_price_per_ticket: float | None,
    area_preference: str | None,
) -> dict[str, Any]:
    results: list[dict[str, Any]] = []
    for detail in items:
        matched_tiers, reason = _match_tiers(
            detail,
            quantity=quantity,
            max_price_per_ticket=max_price_per_ticket,
            area_preference=area_preference,
        )
        results.append(
            {
                "event": detail.model_dump(mode="json"),
                "matched_tiers": [tier.model_dump(mode="json") for tier in matched_tiers[:3]],
                "match_reason": reason,
            }
        )

    return {
        "query": {
            "keyword": keyword,
            "city": city,
            "start_date": start_date,
            "end_date": end_date,
            "quantity": quantity,
            "max_price_per_ticket": max_price_per_ticket,
            "area_preference": area_preference,
        },
        "events": results,
        "total": len(results),
    }


class GatewayTicketingClient:
    def __init__(self, settings: Settings) -> None:
        if not settings.gateway_base_url:
            raise ValueError("gateway_base_url is required in http mode")
        self._settings = settings
        self._http = httpx.AsyncClient(
            base_url=settings.gateway_base_url.rstrip("/"),
            timeout=settings.gateway_timeout_seconds,
        )

    async def close(self) -> None:
        await self._http.aclose()

    async def search_ticket_stock(
        self,
        *,
        keyword: str | None,
        city: str | None,
        start_date: str | None,
        end_date: str | None,
        quantity: int | None,
        max_price_per_ticket: float | None,
        area_preference: str | None,
        page_size: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        params = {
            "keyword": keyword or "",
            "city": city or "",
            "startDate": start_date or "",
            "endDate": end_date or "",
            "page": 1,
            "pageSize": page_size,
        }
        response = await self._http.get(
            "/program/events/search",
            params=params,
            headers=self._headers(access_token),
        )
        response.raise_for_status()
        payload = response.json()
        events = [EventSummary.model_validate(self._map_event_summary(item)) for item in payload.get("events", [])]
        details = await asyncio.gather(
            *[self._fetch_detail(event.id, access_token) for event in events],
            return_exceptions=True,
        )
        valid_details = [item for item in details if isinstance(item, EventDetail)]
        return _build_ticket_search_payload(
            valid_details,
            keyword=keyword,
            city=city,
            start_date=start_date,
            end_date=end_date,
            quantity=quantity,
            max_price_per_ticket=max_price_per_ticket,
            area_preference=area_preference,
        )

    async def get_event_detail(self, event_id: str, access_token: str | None = None) -> dict[str, Any]:
        detail = await self._fetch_detail(event_id, access_token)
        return detail.model_dump(mode="json")

    async def get_hot_recommendations(
        self,
        *,
        city: str | None,
        limit: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        response = await self._http.get(
            "/program/hot-recommend",
            params={"city": city or "", "limit": limit},
            headers=self._headers(access_token),
        )
        response.raise_for_status()
        payload = response.json()
        events = [self._map_event_summary(item) for item in payload.get("events", [])]
        return {"events": events, "total": len(events)}

    async def _fetch_detail(self, event_id: str, access_token: str | None) -> EventDetail:
        response = await self._http.get(
            "/program/events/detail",
            params={"eventId": event_id},
            headers=self._headers(access_token),
        )
        response.raise_for_status()
        payload = response.json().get("event", {})
        return EventDetail.model_validate(self._map_event_detail(payload))

    def _headers(self, access_token: str | None) -> dict[str, str]:
        token = access_token or self._settings.gateway_bearer_token
        if not token:
            return {}
        return {"Authorization": f"Bearer {token}"}

    @staticmethod
    def _map_event_summary(payload: dict[str, Any]) -> dict[str, Any]:
        return {
            "id": str(payload.get("id", "")),
            "title": payload.get("title", ""),
            "city": payload.get("city", ""),
            "artist": payload.get("artist", ""),
            "venue_name": payload.get("venueName", ""),
            "event_start_time": payload.get("eventStartTime", ""),
            "event_end_time": payload.get("eventEndTime"),
            "min_price": payload.get("minPrice"),
            "status": payload.get("status", 0),
            "poster_url": payload.get("posterUrl"),
            "is_hot": payload.get("isHot", False),
        }

    @staticmethod
    def _map_event_detail(payload: dict[str, Any]) -> dict[str, Any]:
        tiers = [
            TicketTier(
                id=str(item.get("id", "")),
                event_id=str(item.get("eventId", "")),
                name=item.get("name", ""),
                price=float(item.get("price", 0)),
                remain_stock=int(item.get("remainStock", 0)),
                status=int(item.get("status", 0)),
            )
            for item in payload.get("ticketTiers", [])
        ]
        venue = payload.get("venue", {}) or {}
        mapped = GatewayTicketingClient._map_event_summary(payload)
        mapped.update(
            {
                "venue_name": venue.get("name") or mapped.get("venue_name"),
                "min_price": min((tier.price for tier in tiers), default=mapped.get("min_price") or 0),
                "description": payload.get("description", ""),
                "sale_start_time": payload.get("saleStartTime"),
                "sale_end_time": payload.get("saleEndTime"),
                "venue_address": venue.get("address"),
                "purchase_limit": payload.get("purchaseLimit"),
                "need_real_name": payload.get("needRealName"),
                "ticket_type": payload.get("ticketType"),
                "ticket_tiers": tiers,
            }
        )
        return mapped


@dataclass(slots=True)
class _MockEventRecord:
    summary: EventSummary
    detail: EventDetail


class MockTicketingClient:
    def __init__(self) -> None:
        self._events = self._build_events()

    async def search_ticket_stock(
        self,
        *,
        keyword: str | None,
        city: str | None,
        start_date: str | None,
        end_date: str | None,
        quantity: int | None,
        max_price_per_ticket: float | None,
        area_preference: str | None,
        page_size: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        _ = access_token
        records = []
        for record in self._events:
            if keyword and keyword not in record.summary.title and keyword not in record.summary.artist:
                continue
            if city and city != record.summary.city:
                continue
            if start_date and record.summary.event_start_time[:10] < start_date:
                continue
            if end_date and record.summary.event_start_time[:10] > end_date:
                continue
            records.append(record.detail)

        return _build_ticket_search_payload(
            records[:page_size],
            keyword=keyword,
            city=city,
            start_date=start_date,
            end_date=end_date,
            quantity=quantity,
            max_price_per_ticket=max_price_per_ticket,
            area_preference=area_preference,
        )

    async def get_event_detail(self, event_id: str, access_token: str | None = None) -> dict[str, Any]:
        _ = access_token
        for record in self._events:
            if record.detail.id == event_id:
                return record.detail.model_dump(mode="json")
        raise ValueError(f"event not found: {event_id}")

    async def get_hot_recommendations(
        self,
        *,
        city: str | None,
        limit: int,
        access_token: str | None,
    ) -> dict[str, Any]:
        _ = access_token
        events = []
        for record in self._events:
            if city and record.summary.city != city:
                continue
            if record.summary.is_hot:
                events.append(record.summary.model_dump(mode="json"))
        if not events:
            events = [record.summary.model_dump(mode="json") for record in self._events]
        return {"events": events[:limit], "total": min(len(events), limit)}

    def _build_events(self) -> list[_MockEventRecord]:
        data = [
            self._make_event(
                event_id="1001",
                title="陈奕迅 FEAR and DREAMS 北京站",
                city="北京",
                artist="陈奕迅",
                venue_name="国家体育场",
                event_start_time="2026-04-04T19:30:00+08:00",
                is_hot=True,
                sale_start_time="2026-03-28T12:00:00+08:00",
                description="北京站返场，支持电子票实名入场。",
                venue_address="北京市朝阳区国家体育场南路1号",
                tiers=[
                    ("t10011", "内场A", 1680, 0),
                    ("t10012", "看台前排", 1280, 18),
                    ("t10013", "看台", 980, 36),
                ],
            ),
            self._make_event(
                event_id="1002",
                title="周杰伦 嘉年华 上海站",
                city="上海",
                artist="周杰伦",
                venue_name="上海体育场",
                event_start_time="2026-04-11T19:00:00+08:00",
                is_hot=True,
                sale_start_time="2026-03-30T11:00:00+08:00",
                description="热门演唱会场次，建议尽早关注开票提醒。",
                venue_address="上海市徐汇区天钥桥路666号",
                tiers=[
                    ("t10021", "VIP内场", 1880, 6),
                    ("t10022", "看台A", 1380, 22),
                    ("t10023", "看台B", 1080, 40),
                ],
            ),
            self._make_event(
                event_id="1003",
                title="陈奕迅 FEAR and DREAMS 杭州站",
                city="杭州",
                artist="陈奕迅",
                venue_name="杭州奥体中心",
                event_start_time="2026-04-05T19:30:00+08:00",
                is_hot=False,
                sale_start_time="2026-03-29T12:00:00+08:00",
                description="同巡演杭州站，票档相对更友好。",
                venue_address="杭州市滨江区飞虹路3号",
                tiers=[
                    ("t10031", "内场", 1480, 12),
                    ("t10032", "看台前排", 1180, 25),
                    ("t10033", "看台", 880, 54),
                ],
            ),
        ]
        return data

    def _make_event(
        self,
        *,
        event_id: str,
        title: str,
        city: str,
        artist: str,
        venue_name: str,
        event_start_time: str,
        is_hot: bool,
        sale_start_time: str,
        description: str,
        venue_address: str,
        tiers: list[tuple[str, str, float, int]],
    ) -> _MockEventRecord:
        ticket_tiers = [
            TicketTier(
                id=tier_id,
                event_id=event_id,
                name=name,
                price=price,
                remain_stock=stock,
                status=1 if stock > 0 else 2,
            )
            for tier_id, name, price, stock in tiers
        ]
        min_price = min(item.price for item in ticket_tiers)
        summary = EventSummary(
            id=event_id,
            title=title,
            city=city,
            artist=artist,
            venue_name=venue_name,
            event_start_time=event_start_time,
            event_end_time=event_start_time,
            min_price=min_price,
            status=1,
            is_hot=is_hot,
        )
        detail = EventDetail(
            **summary.model_dump(),
            description=description,
            sale_start_time=sale_start_time,
            sale_end_time=None,
            venue_address=venue_address,
            purchase_limit=4,
            need_real_name=1,
            ticket_type=1,
            ticket_tiers=ticket_tiers,
        )
        return _MockEventRecord(summary=summary, detail=detail)


def build_tool_client(settings: Settings) -> ToolClient:
    if settings.gateway_mode == "http" and settings.gateway_base_url:
        return GatewayTicketingClient(settings)
    return MockTicketingClient()


def summarize_search_result(result: dict[str, Any]) -> str:
    if not result.get("events"):
        return "没有找到符合条件的场次"

    lines = []
    for item in result["events"][:3]:
        event = item["event"]
        tiers = item.get("matched_tiers") or []
        if tiers:
            tier_text = "，".join(
                f'{tier["name"]} {tier["price"]:.0f}元 余票{tier["remain_stock"]}'
                for tier in tiers[:2]
            )
        else:
            tier_text = "暂无符合条件的可售票档"
        lines.append(
            f'{event["title"]} | {event["city"]} | {event["event_start_time"][:16]} | {truncate(tier_text, 56)}'
        )
    return "\n".join(lines)

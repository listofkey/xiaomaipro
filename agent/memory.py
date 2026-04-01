from __future__ import annotations

from typing import Protocol

import redis.asyncio as redis

from agent.config import Settings
from agent.schemas import ConversationMessage, ConversationState, RecentEvent, SlotExtraction
from agent.utils import json_dumps


class ConversationStore(Protocol):
    async def load(self, session_id: str) -> ConversationState:
        ...

    async def save(self, state: ConversationState) -> None:
        ...

    async def append_exchange(
        self,
        session_id: str,
        user_message: str,
        assistant_message: str,
        slots: SlotExtraction,
        recent_events: list[RecentEvent] | None = None,
    ) -> ConversationState:
        ...


class InMemoryConversationStore:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._states: dict[str, ConversationState] = {}

    async def load(self, session_id: str) -> ConversationState:
        return self._states.get(session_id, ConversationState(session_id=session_id))

    async def save(self, state: ConversationState) -> None:
        state.messages = self._trim_messages(state.messages)
        self._states[state.session_id] = state

    async def append_exchange(
        self,
        session_id: str,
        user_message: str,
        assistant_message: str,
        slots: SlotExtraction,
        recent_events: list[RecentEvent] | None = None,
    ) -> ConversationState:
        state = await self.load(session_id)
        state.messages.extend(
            [
                ConversationMessage(role="user", content=user_message),
                ConversationMessage(role="assistant", content=assistant_message),
            ]
        )
        state.messages = self._trim_messages(state.messages)
        state.slots = state.slots.merge(slots)
        if recent_events is not None:
            state.recent_events = recent_events
        await self.save(state)
        return state

    def _trim_messages(self, messages: list[ConversationMessage]) -> list[ConversationMessage]:
        keep = max(self._settings.session_history_limit * 2, 2)
        if len(messages) <= keep:
            return messages
        return messages[-keep:]


class RedisConversationStore:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._client = redis.from_url(settings.redis_url, decode_responses=True)

    def _key(self, session_id: str) -> str:
        return f"ticketing_copilot:session:{session_id}"

    async def load(self, session_id: str) -> ConversationState:
        payload = await self._client.get(self._key(session_id))
        if not payload:
            return ConversationState(session_id=session_id)
        return ConversationState.model_validate_json(payload)

    async def save(self, state: ConversationState) -> None:
        state.messages = state.messages[-max(self._settings.session_history_limit * 2, 2) :]
        await self._client.set(
            self._key(state.session_id),
            json_dumps(state.model_dump(mode="json")),
            ex=self._settings.session_ttl_seconds,
        )

    async def append_exchange(
        self,
        session_id: str,
        user_message: str,
        assistant_message: str,
        slots: SlotExtraction,
        recent_events: list[RecentEvent] | None = None,
    ) -> ConversationState:
        state = await self.load(session_id)
        state.messages.extend(
            [
                ConversationMessage(role="user", content=user_message),
                ConversationMessage(role="assistant", content=assistant_message),
            ]
        )
        state.slots = state.slots.merge(slots)
        if recent_events is not None:
            state.recent_events = recent_events
        await self.save(state)
        return state


def build_conversation_store(settings: Settings) -> ConversationStore:
    if settings.redis_url:
        return RedisConversationStore(settings)
    return InMemoryConversationStore(settings)

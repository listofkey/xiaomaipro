from __future__ import annotations

from datetime import datetime
from typing import Any, Literal

from pydantic import BaseModel, Field


class SlotExtraction(BaseModel):
    artist: str | None = None
    city: str | None = None
    date_hint: str | None = None
    quantity: int | None = None
    budget_total: float | None = None
    budget_per_ticket: float | None = None
    area_preference: str | None = None
    keyword: str | None = None

    def merge(self, other: "SlotExtraction | None") -> "SlotExtraction":
        if other is None:
            return self.model_copy(deep=True)
        payload: dict[str, Any] = self.model_dump()
        for key, value in other.model_dump().items():
            if value not in (None, "", []):
                payload[key] = value
        return SlotExtraction(**payload)


class ConversationMessage(BaseModel):
    role: Literal["system", "user", "assistant"] = "user"
    content: str
    created_at: datetime = Field(default_factory=datetime.utcnow)


class RecentEvent(BaseModel):
    event_id: str
    title: str
    city: str | None = None
    artist: str | None = None
    venue_name: str | None = None
    event_start_time: str | None = None


class ConversationState(BaseModel):
    session_id: str
    messages: list[ConversationMessage] = Field(default_factory=list)
    slots: SlotExtraction = Field(default_factory=SlotExtraction)
    recent_events: list[RecentEvent] = Field(default_factory=list)


class Citation(BaseModel):
    source: str
    snippet: str
    score: float | None = None
    updated_at: str | None = None


class CardAction(BaseModel):
    label: str
    action_type: Literal["detail", "buy", "search", "external"] = "detail"
    url: str | None = None
    payload: dict[str, Any] = Field(default_factory=dict)


class TicketCard(BaseModel):
    card_type: Literal["event", "knowledge"] = "event"
    title: str
    subtitle: str | None = None
    body: str | None = None
    tags: list[str] = Field(default_factory=list)
    event_id: str | None = None
    actions: list[CardAction] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)


class ToolTrace(BaseModel):
    name: str
    arguments: dict[str, Any] = Field(default_factory=dict)
    status: Literal["success", "error"] = "success"
    summary: str
    latency_ms: int | None = None


class ChatRequest(BaseModel):
    session_id: str
    message: str
    user_id: str | None = None
    access_token: str | None = None
    top_k: int = 4


class ChatResponse(BaseModel):
    session_id: str
    answer: str
    cards: list[TicketCard] = Field(default_factory=list)
    citations: list[Citation] = Field(default_factory=list)
    tools: list[ToolTrace] = Field(default_factory=list)
    slots: SlotExtraction = Field(default_factory=SlotExtraction)
    fallback_mode: bool = False
    recent_events: list[RecentEvent] = Field(default_factory=list, exclude=True)


class KnowledgeDocumentIn(BaseModel):
    source: str
    category: str = "rule"
    content: str
    updated_at: str | None = None


class KnowledgeIngestRequest(BaseModel):
    documents: list[KnowledgeDocumentIn] = Field(default_factory=list)
    bootstrap_from_directory: bool = True


class KnowledgeIngestResponse(BaseModel):
    ingested_count: int
    collection_name: str


class HealthResponse(BaseModel):
    status: Literal["ok"] = "ok"
    app_name: str
    llm_enabled: bool
    gateway_mode: str
    redis_enabled: bool
    rag_enabled: bool

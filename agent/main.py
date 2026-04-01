from __future__ import annotations

import argparse
import asyncio
from contextlib import asynccontextmanager
from typing import Any

import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse

from agent.config import Settings, get_settings
from agent.memory import build_conversation_store
from agent.rag import TicketingKnowledgeBase
from agent.safety import TicketingSafetyGuard
from agent.schemas import (
    ChatRequest,
    ChatResponse,
    HealthResponse,
    KnowledgeIngestRequest,
    KnowledgeIngestResponse,
)
from agent.service import TicketingAgentService
from agent.tools import build_tool_client


def create_app(settings: Settings | None = None) -> FastAPI:
    app_settings = settings or get_settings()
    conversation_store = build_conversation_store(app_settings)
    tool_client = build_tool_client(app_settings)
    knowledge_base = TicketingKnowledgeBase(app_settings)
    safety_guard = TicketingSafetyGuard()
    service = TicketingAgentService(
        app_settings,
        conversation_store,
        tool_client,
        knowledge_base,
        safety_guard,
    )

    @asynccontextmanager
    async def lifespan(_: FastAPI):
        await service.bootstrap()
        yield
        if hasattr(tool_client, "close"):
            await tool_client.close()  # type: ignore[attr-defined]

    app = FastAPI(title=app_settings.app_name, lifespan=lifespan)
    app.state.settings = app_settings
    app.state.service = service
    app.state.knowledge_base = knowledge_base

    @app.get("/healthz", response_model=HealthResponse)
    async def healthz() -> HealthResponse:
        return HealthResponse(
            app_name=app_settings.app_name,
            llm_enabled=app_settings.llm_enabled,
            gateway_mode=app_settings.gateway_mode,
            redis_enabled=bool(app_settings.redis_url),
            rag_enabled=knowledge_base.enabled,
        )

    @app.post("/api/v1/chat", response_model=ChatResponse)
    async def chat(payload: ChatRequest, request: Request) -> ChatResponse:
        return await request.app.state.service.chat(payload)

    @app.post("/api/v1/chat/stream")
    async def chat_stream(payload: ChatRequest, request: Request) -> StreamingResponse:
        stream = request.app.state.service.stream(payload)
        return StreamingResponse(
            stream,
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no",
            },
        )

    @app.post("/api/v1/knowledge/ingest", response_model=KnowledgeIngestResponse)
    async def ingest_knowledge(payload: KnowledgeIngestRequest, request: Request) -> KnowledgeIngestResponse:
        total = 0
        if payload.documents:
            total += await request.app.state.knowledge_base.ingest_documents(payload.documents)
        if payload.bootstrap_from_directory:
            total += await request.app.state.knowledge_base.ingest_from_directory(app_settings.knowledge_directory)
        return KnowledgeIngestResponse(
            ingested_count=total,
            collection_name=app_settings.knowledge_collection_name,
        )

    return app


app = create_app()


async def ingest_once() -> int:
    settings = get_settings()
    knowledge_base = TicketingKnowledgeBase(settings)
    total = await knowledge_base.ingest_from_directory(settings.knowledge_directory)
    print(
        f"Ingested {total} knowledge chunks into {settings.knowledge_collection_name}"
    )
    return total


def main(argv: list[str] | None = None) -> Any:
    parser = argparse.ArgumentParser(description="Ticketing Copilot agent service")
    parser.add_argument("command", nargs="?", default="serve", choices=["serve", "ingest"])
    args = parser.parse_args(argv)

    settings = get_settings()
    if args.command == "ingest":
        return asyncio.run(ingest_once())

    if settings.app_debug:
        uvicorn.run(
            "agent.main:app",
            host=settings.app_host,
            port=settings.app_port,
            reload=True,
        )
        return None

    uvicorn.run(
        app,
        host=settings.app_host,
        port=settings.app_port,
    )
    return None


if __name__ == "__main__":
    main()

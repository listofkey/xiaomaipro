from __future__ import annotations

from functools import lru_cache
from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


BASE_DIR = Path(__file__).resolve().parent


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=BASE_DIR / ".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    app_name: str = "Ticketing Copilot Agent"
    app_env: str = "development"
    app_host: str = "0.0.0.0"
    app_port: int = 9000
    app_debug: bool = False
    app_timezone: str = "Asia/Shanghai"
    app_reference_date: str | None = None

    openai_api_key: str | None = None
    openai_base_url: str | None = None
    openai_chat_model: str = "gpt-5.4"
    openai_timeout_seconds: float = 30.0

    embedding_api_key: str | None = None
    embedding_base_url: str | None = None
    embedding_model: str = "text-embedding-v4"
    embedding_check_ctx_length: bool = False

    gateway_mode: str = "mock"
    gateway_base_url: str | None = None
    gateway_bearer_token: str | None = None
    gateway_timeout_seconds: float = 8.0

    chroma_persist_directory: Path = Field(default_factory=lambda: BASE_DIR / ".chroma")
    knowledge_directory: Path = Field(default_factory=lambda: BASE_DIR / "knowledge")
    knowledge_collection_name: str = "ticketing_knowledge"
    knowledge_chunk_size: int = 500
    knowledge_chunk_overlap: int = 80
    knowledge_default_top_k: int = 4
    bootstrap_knowledge_on_startup: bool = True

    redis_url: str | None = None
    session_ttl_seconds: int = 1800
    session_history_limit: int = 8

    allow_mock_fallback: bool = True
    answer_max_tool_rounds: int = 3
    answer_max_tokens: int = 700

    @property
    def llm_enabled(self) -> bool:
        return bool(self.openai_api_key)

    @property
    def embeddings_enabled(self) -> bool:
        return bool(self.embedding_api_key or self.openai_api_key)


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()

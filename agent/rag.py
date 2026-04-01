from __future__ import annotations

from pathlib import Path

from langchain_chroma import Chroma
from langchain_core.documents import Document
from langchain_openai import OpenAIEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter

from agent.config import Settings
from agent.schemas import Citation, KnowledgeDocumentIn
from agent.utils import truncate


class TicketingKnowledgeBase:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._store: Chroma | None = None
        if settings.embeddings_enabled:
            embeddings = OpenAIEmbeddings(
                model=settings.embedding_model,
                openai_api_key=settings.embedding_api_key or settings.openai_api_key,
                openai_api_base=settings.embedding_base_url or settings.openai_base_url,
                request_timeout=settings.openai_timeout_seconds,
                check_embedding_ctx_length=settings.embedding_check_ctx_length,
            )
            self._store = Chroma(
                collection_name=settings.knowledge_collection_name,
                embedding_function=embeddings,
                persist_directory=str(settings.chroma_persist_directory),
            )
        self._splitter = RecursiveCharacterTextSplitter(
            chunk_size=settings.knowledge_chunk_size,
            chunk_overlap=settings.knowledge_chunk_overlap,
        )

    @property
    def enabled(self) -> bool:
        return self._store is not None

    async def bootstrap(self) -> int:
        if not self.enabled or not self._settings.bootstrap_knowledge_on_startup:
            return 0
        return await self.ingest_from_directory(self._settings.knowledge_directory)

    async def ingest_from_directory(self, directory: Path) -> int:
        if not self.enabled or not directory.exists():
            return 0

        documents: list[KnowledgeDocumentIn] = []
        for path in sorted(directory.rglob("*")):
            if path.suffix.lower() not in {".md", ".txt"}:
                continue
            content = path.read_text(encoding="utf-8").strip()
            if not content:
                continue
            documents.append(
                KnowledgeDocumentIn(
                    source=str(path.relative_to(directory)),
                    content=content,
                    category="rule",
                )
            )
        return await self.ingest_documents(documents)

    async def ingest_documents(self, documents: list[KnowledgeDocumentIn]) -> int:
        if not self.enabled or not documents:
            return 0

        split_docs: list[Document] = []
        ids: list[str] = []
        for item in documents:
            base_doc = Document(
                page_content=item.content,
                metadata={
                    "source": item.source,
                    "category": item.category,
                    "updated_at": item.updated_at or "",
                },
            )
            chunks = self._splitter.split_documents([base_doc])
            for index, chunk in enumerate(chunks):
                chunk_id = f"{item.source}:{index}"
                split_docs.append(chunk)
                ids.append(chunk_id)

        if not split_docs or self._store is None:
            return 0

        try:
            self._store.delete(ids=ids)
        except Exception:
            pass
        self._store.add_documents(split_docs, ids=ids)
        return len(split_docs)

    async def search(self, query: str, top_k: int) -> tuple[str, list[Citation]]:
        if not self.enabled or self._store is None:
            return "", []

        results = self._store.similarity_search_with_relevance_scores(query, k=top_k)
        citations: list[Citation] = []
        contexts: list[str] = []
        for doc, score in results:
            source = str(doc.metadata.get("source", "knowledge"))
            contexts.append(f"[{source}] {truncate(doc.page_content, 320)}")
            citations.append(
                Citation(
                    source=source,
                    snippet=truncate(doc.page_content, 160),
                    score=round(float(score), 4) if score is not None else None,
                    updated_at=doc.metadata.get("updated_at") or None,
                )
            )
        return "\n".join(contexts), citations

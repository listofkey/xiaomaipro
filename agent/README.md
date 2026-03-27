## OpenAI RAG

`embed.py` now uses OpenAI vector stores for document ingestion and retrieval.

Before running `ingest`, put your knowledge content into `data.md`.

### Environment variables

PowerShell example:

```powershell
$env:OPENAI_API_KEY="sk-..."
```

Optional variables:

```powershell
$env:OPENAI_BASE_URL="https://your-compatible-endpoint/v1"
$env:OPENAI_CHAT_MODEL="gpt-4o-mini"
$env:OPENAI_VECTOR_STORE_NAME="xiaomaipro-rag"
```

If you use a compatible gateway, it must support the `vector_stores` API. Otherwise `ingest` and `search` will fail.

### Commands

From the `agent` directory:

```powershell
uv run .\embed.py ingest
```

Search:

```powershell
uv run .\embed.py search "这里输入你的问题"
```

Ask with RAG:

```powershell
uv run .\embed.py ask "这里输入你的问题"
```

`agent.py` also reads the API key from the environment now.

From the repository root, the equivalent command is:

```powershell
uv run --project agent python agent/embed.py ingest
```

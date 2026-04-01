# Ticketing Copilot Agent

`agent` 模块现在提供一个可运行的票务智能问答服务，覆盖：

- 自然语言票务查询
- 实名 / 退票 / 入场规则解释
- 多轮会话记忆
- RAG 检索
- SSE 流式输出
- `mock` / `http gateway` 两种工具调用模式

## 运行方式

在仓库根目录执行：

```powershell
uv run --project agent python agent/main.py
```

默认监听 `0.0.0.0:9000`。

## 常用环境变量

```powershell
$env:OPENAI_API_KEY="sk-..."
$env:OPENAI_BASE_URL="https://your-openai-compatible-endpoint/v1"
$env:OPENAI_CHAT_MODEL="gpt-5.4"

$env:EMBEDDING_API_KEY="sk-..."
$env:EMBEDDING_BASE_URL="https://your-embedding-endpoint/v1"
$env:EMBEDDING_MODEL="text-embedding-v4"

$env:GATEWAY_MODE="mock"   # mock 或 http
$env:GATEWAY_BASE_URL="http://localhost:8080"
$env:GATEWAY_BEARER_TOKEN="your-token"

$env:REDIS_URL="redis://localhost:6379/0"
```

如果没有配置 `OPENAI_API_KEY`，服务会自动退回规则式回答 + mock/http 工具查询，不会阻塞启动。

## 知识库导入

将规则文档放入 `agent/knowledge/` 后执行：

```powershell
uv run --project agent python agent/main.py ingest
```

启动时也会尝试自动导入该目录下的 `.md` / `.txt` 文件。

## API

### 健康检查

```http
GET /healthz
```

### 普通问答

```http
POST /api/v1/chat
Content-Type: application/json

{
  "session_id": "demo-session",
  "message": "帮我找两张下周六陈奕迅在北京的内场票，预算一共3000元"
}
```

### SSE 流式问答

```http
POST /api/v1/chat/stream
Content-Type: application/json
Accept: text/event-stream

{
  "session_id": "demo-session",
  "message": "那杭州的呢？"
}
```

### 手动导入知识

```http
POST /api/v1/knowledge/ingest
Content-Type: application/json

{
  "bootstrap_from_directory": true,
  "documents": []
}
```

## 当前实现说明

- 工具调用优先走 `query_ticket_stock / query_event_detail / search_ticket_policy / recommend_hot_events`
- 票务边界被固定为“查 + 解释 + 跳转”，不会执行写操作
- `http` 模式当前调用的是 Go gateway 的 REST 接口，后续切 gRPC 时只需要替换 tool client

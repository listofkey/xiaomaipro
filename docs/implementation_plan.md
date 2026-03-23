# 🤖 票务抢票系统 — AI & Agent 功能融合方案

> [!IMPORTANT]
> 本方案在原有票务系统设计基础上，融入 **AI 能力 + Agent 智能体** 概念，覆盖用户端、运营端和系统自治三个层面。

---

## 一、AI 功能全景图

将 AI 能力按使用场景分为 **用户侧 AI**、**运营侧 AI**、**系统自治 Agent** 三大类：

```mermaid
graph LR
    subgraph 用户侧AI["🧑 用户侧 AI"]
        A1["智能客服 Agent"]
        A2["个性化推荐引擎"]
        A3["智能选座助手"]
        A4["自然语言搜索"]
        A5["购票意图理解"]
    end

    subgraph 运营侧AI["👔 运营侧 AI"]
        B1["智能定价 Agent"]
        B2["销售预测 Agent"]
        B3["内容生成 Agent"]
        B4["数据洞察 Agent"]
        B5["智能排期助手"]
    end

    subgraph 系统自治Agent["⚙️ 系统自治 Agent"]
        C1["风控反欺诈 Agent"]
        C2["流量调度 Agent"]
        C3["异常自愈 Agent"]
        C4["舆情监控 Agent"]
    end

    A1 -.-> LLM["LLM 大模型"]
    A4 -.-> LLM
    B3 -.-> LLM
    B4 -.-> LLM
    C4 -.-> LLM

    A2 -.-> ML["推荐/ML 模型"]
    B1 -.-> ML
    B2 -.-> ML
    C1 -.-> ML
```

---

## 二、用户侧 AI 功能详细设计

### 1. 🤖 智能客服 Agent（核心功能）

> 基于 LLM 的多轮对话客服，具备 **Tool Use（工具调用）** 能力，真正的 Agent 架构。

**Agent 架构设计：**

```mermaid
graph TB
    USER["用户提问"] --> AGENT["客服 Agent（LLM）"]
    AGENT --> PLAN["思考 & 规划"]
    PLAN --> TOOLS{"选择工具"}
    
    TOOLS --> T1["🔍 订单查询工具"]
    TOOLS --> T2["📋 活动查询工具"]
    TOOLS --> T3["💰 退款操作工具"]
    TOOLS --> T4["📖 知识库检索工具(RAG)"]
    TOOLS --> T5["👨‍💼 转人工工具"]
    TOOLS --> T6["📝 工单创建工具"]
    
    T1 --> OBS["观察结果"]
    T2 --> OBS
    T3 --> OBS
    T4 --> OBS
    T5 --> OBS
    T6 --> OBS
    
    OBS --> AGENT
    AGENT --> REPLY["生成回复"]
```

**具备的 Tool（工具）列表：**

| 工具名 | 功能 | 示例触发 |
|--------|------|---------|
| `query_order` | 查询用户订单列表/详情 | "我的订单什么状态了？" |
| `query_event` | 查询活动信息/余票/票价 | "周杰伦演唱会还有票吗？" |
| `apply_refund` | 发起退款申请 | "我要退票" |
| `search_knowledge` | RAG 检索常见问题知识库 | "电子票怎么入场？" |
| `check_ticket_status` | 查询电子票验票状态 | "我的票能用吗？" |
| `transfer_human` | 转接人工客服 | "我要找人工" |
| `create_ticket` | 创建客诉工单 | 复杂问题升级 |

**对话示例 — Agent 多步推理：**

```
用户: 我上周买的那场演唱会，临时去不了，能退吗？

Agent 思考: 用户要退票，需要先查到具体订单 → 检查退款政策 → 执行退款
  ├─ Step 1: 调用 query_order(user_id, status=已支付, 时间=最近一周)
  │   → 找到订单 #2026022800123, 周杰伦演唱会 3/15, ￥880
  ├─ Step 2: 调用 search_knowledge("退款政策")
  │   → 活动开始前72小时可全额退款，24-72小时退80%
  ├─ Step 3: 计算距离活动还有15天，符合全额退款
  └─ 回复: 找到您的订单 #2026022800123，是3月15日周杰伦演唱会的内场票 ￥880。
          距离演出还有15天，满足全额退款条件。是否确认退款？

用户: 好的，退了吧

Agent 思考: 用户确认退款
  ├─ Step 1: 调用 apply_refund(order_id=2026022800123, reason="用户主动申请")
  │   → 退款申请已提交，预计1-3个工作日到账
  └─ 回复: 退款已提交！￥880 将在1-3个工作日原路退回。退款编号 R20260228001。
```

**技术实现要点：**

- 后端新增 **AI 服务（ai-svc）** 微服务，封装 LLM 调用和 Tool 编排
- 使用 **Function Calling** 能力（兼容 OpenAI API 格式的国产大模型）
- 会话上下文存储在 Redis（TTL 30 分钟）
- 前端使用 **SSE 流式输出**（你之前已实现过）
- 知识库使用文档切片 + Embedding + 向量数据库（Milvus/pgvector）检索

---

### 2. 🎯 个性化推荐引擎

**推荐策略矩阵：**

| 推荐场景 | 策略 | 数据来源 |
|---------|------|---------|
| 首页推荐 | 协同过滤 + 热度加权 | 用户浏览/购票/收藏记录 |
| 活动详情页 "看了又看" | Item-based 相似推荐 | 用户共现行为矩阵 |
| "猜你喜欢" | User Embedding + 内容标签 | 用户画像 + 活动标签 |
| 搜索后推荐 | 搜索意图 + 补充推荐 | 搜索词 + 无结果时降级 |
| 开票提醒智能排序 | 用户偏好评分 | 收藏 + 浏览时长 + 历史偏好 |

**用户画像标签体系：**

```
用户画像
├── 基础属性: 城市、年龄段、性别
├── 兴趣偏好: 类型偏好(演唱会/话剧/体育)、艺人偏好、价格敏感度
├── 行为特征: 活跃度、抢票成功率、退票率
└── 消费能力: 平均客单价、购票频率、VIP偏好度
```

---

### 3. 💬 自然语言搜索 & 意图理解

> 用户不再需要手动选筛选条件，直接用自然语言描述需求。

**输入 → 意图解析 → 结构化查询：**

```
用户输入: "下个月北京有什么摇滚演出，500块以内的"

LLM 意图解析 →
{
  "intent": "search_event",
  "params": {
    "city": "北京",
    "time_range": {"start": "2026-03-01", "end": "2026-03-31"},
    "category": "演唱会",
    "tags": ["摇滚"],
    "max_price": 500
  }
}

→ 转换为后端查询条件 → 返回结果
```

**能力扩展 — 多轮对话选票：**

```
用户: 帮我找个周末带孩子去的演出
Agent: 好的！请问您在哪个城市？
用户: 上海
Agent: 为您找到3场适合亲子的周末活动：
       1. 🎭 《小王子》儿童剧 - 3/8 周六 - ￥180起
       2. 🎪 太阳马戏团 - 3/9 周日 - ￥280起
       3. 🎵 迪士尼音乐会 - 3/15 周六 - ￥220起
       您对哪场感兴趣？
用户: 第2个，帮我看看有没有3张连座
Agent: [调用 check_seats(event_id, count=3, adjacent=true)]
       太阳马戏团 3/9 场次，A区有3张连座（A排12-14号），￥380/张。
       要帮您锁定吗？15分钟内需完成支付。
```

---

### 4. 🪑 智能选座助手

> 用户描述偏好，AI 自动推荐最优座位。

```
用户: 帮我选两张靠前排中间的好位置
AI分析:
  ├── 约束条件: 数量=2, 相邻=true
  ├── 偏好权重: 前排(0.4) + 中央(0.4) + 价格适中(0.2)
  ├── 可用座位评分计算
  └── 推荐: B区3排15-16号 (评分92/100)
           理由: 前排5排内+正中央位置+距离舞台约15米
```

---

## 三、运营侧 AI 功能详细设计

### 1. 💰 智能定价 Agent

> 基于市场供需、历史数据、竞品分析的动态定价建议 Agent。

**Agent 工作流程：**

```mermaid
graph LR
    INPUT["定价请求"] --> COLLECT["数据收集 Agent"]
    COLLECT --> |历史同类活动| ANALYZE["分析 Agent"]
    COLLECT --> |当前市场热度| ANALYZE
    COLLECT --> |艺人热搜指数| ANALYZE
    ANALYZE --> PREDICT["预测模型"]
    PREDICT --> SUGGEST["定价建议"]
    SUGGEST --> HUMAN["运营审核"]
    HUMAN --> |确认| APPLY["应用价格"]
    HUMAN --> |调整| SUGGEST
```

**输出示例：**

```
🎯 定价建议报告 — 周杰伦2026巡回演唱会·北京站

📊 参考数据:
  - 同类演唱会平均票价: VIP ￥1580, 内场 ￥980, 看台 ￥480
  - 周杰伦2024北京站: VIP ￥1280, 内场 ￥880, 看台 ￥380 (30秒售罄)
  - 当前微博热搜指数: 9800万 (极高)
  - 鸟巢容量: 80000, 预估供需比: 1:50

💡 建议定价:
  VIP:   ￥1880 (上浮 19%, 信心度 85%)
  内场:  ￥1080 (上浮 10%, 信心度 90%)
  看台A: ￥580  (上浮 5%,  信心度 92%)
  看台B: ￥380  (持平,     信心度 95%)

⚠️ 风险提示: VIP定价上浮较大，建议关注舆情反馈
```

---

### 2. 📈 销售预测 Agent

| 预测能力 | 说明 |
|---------|------|
| 开售销量预测 | 预测开售后 1分钟/5分钟/30分钟 的销量曲线 |
| 售罄时间预测 | 预测每个票档的售罄时间点 |
| 库存调整建议 | 根据预售/关注度推荐加场或调整票档库存 |
| 服务器扩容建议 | 基于预测流量推荐开售时的服务器扩容倍数 |

---

### 3. ✍️ 内容生成 Agent

> 自动生成运营所需的各类文案内容。

| 生成能力 | 示例 |
|---------|------|
| 活动描述文案 | 输入艺人+场地+日期，自动生成吸引人的活动详情 |
| 营销推广文案 | 生成社交媒体推广文案（小红书/微博/朋友圈风格） |
| 短信通知模板 | 根据场景自动生成开票提醒/抢票成功等短信文案 |
| SEO 优化标题 | 自动生成搜索引擎友好的活动标题和描述 |
| 客服话术库 | 基于常见问题自动生成标准化客服回复模板 |

---

### 4. 📊 数据洞察 Agent（对话式 BI）

> 运营人员用自然语言提问，Agent 自动查询数据并生成可视化报告。

```
运营: 上周演唱会类活动的销售情况怎么样？

Agent:
  ├─ 调用 query_sales_data(category="演唱会", period="last_week")
  ├─ 数据分析 & 可视化
  └─ 回复:
     📊 上周演唱会类销售报告 (2/17 - 2/23):
     ─────────────────────
     总销售额: ￥2,380,000  (环比 +15%)
     总出票量: 4,200 张    (环比 +8%)
     热销 TOP3:
       1. 薛之谦巡回·上海站 - ￥980,000 (售罄)
       2. 五月天·北京站     - ￥680,000 (余票 12%)
       3. 新裤子·广州站     - ￥320,000 (余票 35%)
     
     💡 洞察: 上海市场演唱会需求旺盛，建议增加场次。
              低价票档占比过高(65%)，可考虑优化票档结构。
```

---

## 四、系统自治 Agent 详细设计

### 1. 🛡 风控反欺诈 Agent（核心 Agent）

> 实时检测并阻断黄牛、机器人、异常行为，具备自主决策和学习能力。

**Agent 多层检测架构：**

```mermaid
graph TB
    REQ["用户请求"] --> L1["第一层: 规则引擎(毫秒级)"]
    L1 --> |通过| L2["第二层: ML实时评分(百毫秒级)"]
    L1 --> |拦截| BLOCK["阻断 + 记录"]
    L2 --> |通过| L3["第三层: Agent深度分析(秒级)"]
    L2 --> |拦截| BLOCK
    L3 --> |可疑| REVIEW["人工审核队列"]
    L3 --> |正常| PASS["放行"]
    
    BLOCK --> LEARN["反馈学习"]
    REVIEW --> LEARN
    LEARN --> L1
    LEARN --> L2
```

**检测维度：**

| 维度 | 检测点 | Agent 决策 |
|------|--------|-----------|
| 设备指纹 | 同一设备多账号、模拟器、改机工具 | 自动标记 + 加人机验证 |
| 行为分析 | 请求频率异常、操作路径过短（跳过浏览直接下单） | 降低优先级 / 加验证码难度 |
| IP 分析 | 代理 IP、机房 IP、同 IP 大量请求 | 动态限流 / 封禁 IP 段 |
| 关联分析 | 多个账号关联同一手机/地址/支付方式 | 标记为疑似黄牛团伙 |
| 时序分析 | 购票时间过于精确（毫秒级卡点） | 延迟处理 + 深度审查 |

---

### 2. 🌊 流量调度 Agent

> 自主感知系统负载，动态调整限流策略和资源分配。

```mermaid
graph LR
    MONITOR["监控指标"] --> AGENT["流量调度 Agent"]
    AGENT --> |CPU > 80%| SCALE["自动扩容"]
    AGENT --> |QPS突增| LIMIT["动态限流"]
    AGENT --> |Redis延迟高| CACHE["缓存策略调整"]
    AGENT --> |数据库慢查询| DB["读写分离优化"]
    
    MONITOR -.-> |实时数据| PROMETHEUS["Prometheus"]
    PROMETHEUS -.-> GRAFANA["Grafana告警"]
```

**自治能力：**

| 场景 | Agent 自主行为 |
|------|---------------|
| 开票前 5 分钟流量骤增 | 自动预扩容服务实例、预热 Redis 缓存 |
| 某票档快速售罄 | 自动将流量引导至其他票档页面 |
| 数据库 QPS 接近阈值 | 自动提升缓存 TTL、降级非核心查询 |
| 单节点异常 | 自动摘除故障节点、通知运维 |

---

### 3. 🔧 异常自愈 Agent

| 异常场景 | Agent 自愈动作 |
|---------|---------------|
| Redis-MySQL 库存不一致 | 自动触发对账、修正 Redis 数据 |
| 消息积压超阈值 | 自动增加 Kafka Consumer 实例 |
| 支付回调丢失 | 主动轮询第三方支付查询接口补偿 |
| 订单超时未处理 | 自动触发延迟取消 + 库存归还 |

---

## 五、AI 模块新增数据表设计

在原有 ER 图基础上，新增以下 AI 相关表：

```mermaid
erDiagram
    AI_CHAT_SESSION {
        bigint id PK "会话ID"
        bigint user_id FK "用户ID"
        tinyint scene "场景: 1客服 2选座 3搜索推荐"
        tinyint status "状态: 0进行中 1已结束 2转人工"
        text summary "AI总结的会话摘要"
        int message_count "消息数量"
        int tool_call_count "工具调用次数"
        tinyint satisfaction "满意度: 1-5星"
        datetime created_at "创建时间"
        datetime ended_at "结束时间"
    }

    AI_CHAT_MESSAGE {
        bigint id PK "消息ID"
        bigint session_id FK "会话ID"
        tinyint role "角色: 1用户 2AI 3系统 4工具"
        text content "消息内容"
        varchar tool_name "工具名（role=4时）"
        text tool_args "工具参数JSON"
        text tool_result "工具返回结果JSON"
        int token_count "消耗Token数"
        datetime created_at "创建时间"
    }

    USER_PROFILE_TAG {
        bigint id PK "标签ID"
        bigint user_id FK "用户ID"
        varchar tag_key "标签键（如 genre_pref）"
        varchar tag_value "标签值（如 rock）"
        decimal score "权重评分 0-1"
        datetime updated_at "更新时间"
    }

    AI_RECOMMENDATION_LOG {
        bigint id PK "推荐日志ID"
        bigint user_id FK "用户ID"
        varchar scene "推荐场景"
        text event_ids "推荐的活动ID列表JSON"
        bigint clicked_event_id "用户点击的活动ID"
        tinyint is_converted "是否转化下单: 0否 1是"
        text model_version "模型版本"
        datetime created_at "创建时间"
    }

    RISK_EVENT {
        bigint id PK "风控事件ID"
        bigint user_id FK "关联用户"
        varchar event_type "事件类型"
        tinyint risk_level "风险等级: 1低 2中 3高"
        text detail "详情JSON (IP/设备/行为数据)"
        tinyint action_taken "处置: 0无 1加验证 2限流 3封禁"
        tinyint is_false_positive "是否误判: 0否 1是"
        datetime created_at "触发时间"
    }

    KNOWLEDGE_DOC {
        bigint id PK "知识文档ID"
        varchar title "标题"
        text content "原始内容"
        varchar category "分类(退票/入场/支付...)"
        tinyint status "状态: 0草稿 1启用 2禁用"
        datetime created_at "创建时间"
        datetime updated_at "更新时间"
    }

    KNOWLEDGE_CHUNK {
        bigint id PK "分片ID"
        bigint doc_id FK "文档ID"
        text chunk_text "分片文本"
        int chunk_index "分片序号"
        varchar embedding_model "Embedding模型版本"
        datetime created_at "创建时间"
    }

    USER ||--o{ AI_CHAT_SESSION : "发起会话"
    AI_CHAT_SESSION ||--o{ AI_CHAT_MESSAGE : "包含消息"
    USER ||--o{ USER_PROFILE_TAG : "用户画像标签"
    USER ||--o{ AI_RECOMMENDATION_LOG : "推荐记录"
    USER ||--o{ RISK_EVENT : "风控事件"
    KNOWLEDGE_DOC ||--o{ KNOWLEDGE_CHUNK : "文档分片"
```

---

## 六、新增 AI 服务架构

在原有微服务架构基础上，新增一个 **AI 服务（ai-svc）**：

```mermaid
graph TB
    subgraph AI_SVC["🤖 AI 服务 (ai-svc)"]
        ROUTER["Agent 路由"]
        CS_AGENT["客服 Agent"]
        SEARCH_AGENT["搜索意图 Agent"]
        RISK_AGENT["风控 Agent"]
        RECOMMEND["推荐引擎"]
        RAG["RAG 检索"]
    end

    subgraph 外部AI["🧠 AI 基础设施"]
        LLM["LLM API<br/>(DeepSeek/Qwen/GPT)"]
        VECTOR["向量数据库<br/>(Milvus/pgvector)"]
        EMBEDDING["Embedding 服务"]
    end

    GATEWAY["API Gateway"] --> ROUTER
    ROUTER --> CS_AGENT
    ROUTER --> SEARCH_AGENT
    ROUTER --> RISK_AGENT
    ROUTER --> RECOMMEND

    CS_AGENT --> LLM
    CS_AGENT --> RAG
    SEARCH_AGENT --> LLM
    RISK_AGENT --> LLM

    RAG --> VECTOR
    RAG --> EMBEDDING
    RECOMMEND --> REDIS["Redis<br/>用户画像缓存"]
    RECOMMEND --> MYSQL["MySQL<br/>行为日志"]

    CS_AGENT -.->|Tool Call| ORDER_SVC["订单服务"]
    CS_AGENT -.->|Tool Call| EVENT_SVC["活动服务"]
    CS_AGENT -.->|Tool Call| REFUND_SVC["退款申请"]
```

**技术选型建议：**

| 组件 | 推荐方案 | 理由 |
|------|---------|------|
| LLM | DeepSeek-V3 / 通义千问 / Moonshot | 国产大模型，支持 Function Calling，性价比高 |
| Embedding | text-embedding-3-small / BGE-M3 | BGE 中文效果好，部署成本低 |
| 向量数据库 | Milvus (大规模) / pgvector (轻量) | pgvector 初期足够，Milvus 可扩展 |
| Agent 框架 | 自研 (Go) / LangChain (Python微服务) | 核心链路用 Go，推荐/RAG 可用 Python |

---

## 七、前端 AI 组件新增

```
src/components/ai/
├── AiChatPanel.vue        # AI 客服悬浮聊天面板
├── AiChatBubble.vue       # 聊天气泡（支持Markdown渲染）
├── AiSearchBar.vue        # 自然语言搜索输入框
├── AiSeatRecommend.vue    # 智能选座推荐面板
├── AiEventRecommend.vue   # "为你推荐" 组件
├── AiTypingIndicator.vue  # AI 打字中动画
└── AiSatisfaction.vue     # 满意度评价组件
```

---

## 用户审核事项

> [!WARNING]
> **以下设计决策需要用户确认：**
> 1. **LLM 选型**：推荐使用 DeepSeek-V3 或通义千问（成本低+中文好），还是 OpenAI GPT-4o（效果好但贵且受限）？
> 2. **AI 服务语言**：推荐核心 Agent 用 Go 编写（统一技术栈），推荐/RAG 部分用 Python 微服务。还是全部用 Go？
> 3. **初期 AI 功能范围**：建议优先实现 ① 智能客服 Agent ② 自然语言搜索 ③ 风控 Agent。其余功能后续迭代。是否同意此优先级？
> 4. **向量数据库**：初期用 pgvector（依赖 PostgreSQL），还是直接用 Milvus（独立部署更灵活）？

## 验证方式

由于本方案为架构设计文档，验证方式主要为：

### 人工审查
- 请用户审阅此设计文档，确认 AI 功能范围、技术选型和优先级
- 审阅新增数据表是否满足业务需求
- 确认 Agent 工具列表是否覆盖核心场景

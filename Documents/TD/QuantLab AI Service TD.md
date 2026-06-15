# QuantLab AI Service TD

Version: 1\.0

Module: AI Service

Priority: P1（Supporting Domain → Future Core Capability）

Status: Draft

Owner: AI Platform Team

Dependencies:

```Plain Text
User Service
Strategy Service
Formula Engine
Portfolio Service
Backtest Engine
Market Data Service
Ranking Service
Community Service
Notification Service
Event Bus(Kafka)
LLM Provider(OpenAI/Claude/Qwen/DeepSeek)
Vector Database
```

---

# 服务定位

AI Service 是 QuantLab 的智能研究中枢（Quant Copilot）。

负责：

```Plain Text
自然语言生成策略
策略解释
策略优化建议
组合构建
组合优化
回测分析
风险分析
排行榜分析
社区问答
智能投研助手
```

不负责：

```Plain Text
策略执行
回测计算
行情计算
排名计算
消息发送
```

AI Service 本质定位：

```Plain Text
AI Research Layer
+
Quant Copilot
+
Knowledge Intelligence Platform
```

---

# 产品目标

V1：

```Plain Text
AI生成策略
AI解释策略
AI分析回测
AI生成组合
```

---

V2：

```Plain Text
策略优化
因子发现
策略评分
AI选股助手
```

---

V3：

```Plain Text
Agent化量化研究
自动研究员
自动策略工厂
```

---

# 核心职责

AI Service 提供：

```Plain Text
Strategy Copilot
Portfolio Copilot
Backtest Copilot
Research Copilot
Community Copilot
```

---

# DDD设计

AI Domain

```Plain Text
AITask
PromptTemplate
KnowledgeDocument
Embedding
AIReport
Conversation
```

---

# 聚合根

## AITask

```Plain Text
type AITask struct {
    ID

    UserID

    TaskType

    Status

    Input

    Output

    CreatedAt
}
```

---

# TaskType

```Plain Text
GENERATE_STRATEGY

EXPLAIN_STRATEGY

OPTIMIZE_STRATEGY

GENERATE_PORTFOLIO

OPTIMIZE_PORTFOLIO

ANALYZE_BACKTEST

ANALYZE_RANKING

ASK_QUANTLAB
```

---

# TaskStatus

```Plain Text
PENDING

RUNNING

COMPLETED

FAILED

CANCELLED
```

---

# PromptTemplate

```Plain Text
type PromptTemplate struct {
    ID

    Name

    Version

    Prompt

    Variables
}
```

---

# AIReport

统一分析报告

```Plain Text
type AIReport struct {
    ID

    ObjectType

    ObjectID

    Summary

    Strengths

    Risks

    Suggestions
}
```

---

# ObjectType

```Plain Text
STRATEGY

PORTFOLIO

BACKTEST

RANKING

AUTHOR
```

---

# Conversation

未来聊天能力

```Plain Text
type Conversation struct {
    ID

    UserID

    Title
}
```

---

# 服务目录结构

```Plain Text
ai-service
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── application
│   │
│   ├── domain
│   │   ├── task
│   │   ├── report
│   │   ├── prompt
│   │   ├── conversation
│   │   └── knowledge
│   │
│   ├── llm
│   │   ├── openai
│   │   ├── claude
│   │   ├── deepseek
│   │   └── router
│   │
│   ├── rag
│   │
│   ├── embedding
│   │
│   ├── consumer
│   │
│   ├── workflow
│   │
│   └── scheduler
│
├── configs
│
└── deploy
```

---

# AI能力架构

```Plain Text
User
 ↓
AI Gateway
 ↓
Prompt Builder
 ↓
Context Loader
 ↓
LLM Router
 ↓
Provider
 ↓
Result Processor
 ↓
AI Report
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_ai
```

---

# ai\_task

```Plain Text
CREATE TABLE ai_task
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT,

    task_type VARCHAR(64),

    status VARCHAR(32),

    input_json JSON,

    output_json JSON,

    created_at DATETIME,

    updated_at DATETIME,

    INDEX idx_user(user_id),
    INDEX idx_status(status)
);
```

---

# ai\_report

```Plain Text
CREATE TABLE ai_report
(
    id BIGINT PRIMARY KEY,

    object_type VARCHAR(32),

    object_id BIGINT,

    summary TEXT,

    strengths TEXT,

    risks TEXT,

    suggestions TEXT,

    created_at DATETIME,

    INDEX idx_object(
        object_type,
        object_id
    )
);
```

---

# prompt\_template

```Plain Text
CREATE TABLE prompt_template
(
    id BIGINT PRIMARY KEY,

    name VARCHAR(128),

    version VARCHAR(32),

    prompt TEXT,

    created_at DATETIME
);
```

---

# conversation

```Plain Text
CREATE TABLE conversation
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT,

    title VARCHAR(256),

    created_at DATETIME
);
```

---

# message

```Plain Text
CREATE TABLE message
(
    id BIGINT PRIMARY KEY,

    conversation_id BIGINT,

    role VARCHAR(32),

    content LONGTEXT,

    created_at DATETIME
);
```

---

# knowledge\_document

```Plain Text
CREATE TABLE knowledge_document
(
    id BIGINT PRIMARY KEY,

    source_type VARCHAR(64),

    source_id BIGINT,

    content LONGTEXT,

    embedding_id VARCHAR(128),

    created_at DATETIME
);
```

---

# 向量数据库

推荐：

```Plain Text
Qdrant
```

原因：

```Plain Text
开源
高性能
云原生
支持过滤查询
```

---

Collection：

```Plain Text
strategy_docs

portfolio_docs

backtest_docs

community_docs
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/ai
```

---

# 生成策略

```Plain Text
POST /strategies/generate
```

Request：

```Plain Text
{
  "prompt":"寻找高ROE低PE策略"
}
```

---

Response：

```Plain Text
{
  "task_id":1001
}
```

---

# 解释策略

```Plain Text
POST /strategies/explain
```

---

# 优化策略

```Plain Text
POST /strategies/optimize
```

---

# 生成组合

```Plain Text
POST /portfolios/generate
```

---

# 优化组合

```Plain Text
POST /portfolios/optimize
```

---

# 分析回测

```Plain Text
POST /backtests/analyze
```

---

# AI问答

```Plain Text
POST /chat
```

---

# 获取任务状态

```Plain Text
GET /tasks/{id}
```

---

# 获取报告

```Plain Text
GET /reports/{id}
```

---

# AI策略生成流程

```Plain Text
用户输入
 ↓
Prompt Builder
 ↓
Strategy DSL Generator
 ↓
DSL Validator
 ↓
Formula Engine
 ↓
返回DSL
```

---

# AI解释策略流程

输入：

```Plain Text
Strategy DSL
```

---

输出：

```Plain Text
策略目标

选股逻辑

适用市场

风险分析

历史表现解释
```

---

# AI组合生成流程

输入：

```Plain Text
稳健型投资组合
```

---

输出：

```Plain Text
价值策略 40%

成长策略 30%

红利策略 20%

现金策略 10%
```

---

# AI回测分析流程

输入：

```Plain Text
Backtest Report
```

---

输出：

```Plain Text
收益分析

风险分析

回撤原因

优化建议
```

---

# RAG知识库

知识来源：

```Plain Text
Strategy

Portfolio

Backtest

Ranking

Community Article
```

---

构建：

```Plain Text
Embedding
 ↓
Vector DB
```

---

# Context Builder

统一上下文组装器

```Plain Text
type ContextBuilder interface {
    Build()
}
```

---

支持：

```Plain Text
StrategyContext

PortfolioContext

BacktestContext

RankingContext
```

---

# LLM Router

统一模型路由

```Plain Text
type LLMRouter interface {
    Route()
}
```

---

策略生成：

```Plain Text
Claude
```

---

快速问答：

```Plain Text
GPT

Qwen
```

---

复杂分析：

```Plain Text
Claude
```

---

# Prompt管理

模板化：

```Plain Text
strategy_generate_v1

portfolio_generate_v1

backtest_analyze_v1
```

---

支持：

```Plain Text
A/B测试

版本管理
```

---

# 事件消费

监听：

```Plain Text
strategy-events
portfolio-events
backtest-events
ranking-events
community-events
```

---

# StrategyCreated

触发：

```Plain Text
Embedding构建
```

---

# StrategyPublished

触发：

```Plain Text
AI分析报告
```

---

# PortfolioPublished

触发：

```Plain Text
组合分析
```

---

# BacktestCompleted

触发：

```Plain Text
AI回测总结
```

---

# RankingUpdated

触发：

```Plain Text
AI排行榜洞察
```

---

# AI事件

Topic：

```Plain Text
ai-events
```

---

# AITaskCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "AITaskCreated",
  "aggregate_type": "AI_TASK",
  "producer": "ai-service",
  "payload": {
    "task_id": 1001,
    "task_type": "GENERATE_STRATEGY"
  }
}
```

---

# AITaskCompleted

```JSON
{
  "event_id": "uuid",
  "event_type": "AITaskCompleted",
  "aggregate_type": "AI_TASK",
  "aggregate_id": "1001",
  "payload": {
    "task_id": 1001,
    "status": "COMPLETED"
  }
}
```

---

# StrategyGenerated

```JSON
{
  "event_id": "uuid",
  "event_type": "StrategyGenerated",
  "aggregate_type": "AI_TASK",
  "payload": {
    "strategy_id": 2001
  }
}
```

---

# PortfolioGenerated

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioGenerated",
  "aggregate_type": "AI_TASK",
  "payload": {
    "portfolio_id": 3001
  }
}
```

---

# Redis缓存设计

任务状态：

```Plain Text
ai:task:{id}
```

TTL：

```Plain Text
1h
```

---

报告：

```Plain Text
ai:report:{type}:{id}
```

TTL：

```Plain Text
24h
```

---

Prompt：

```Plain Text
ai:prompt:{name}
```

TTL：

```Plain Text
6h
```

---

# 权限设计

Free：

```Plain Text
20次AI调用/天
```

---

Pro：

```Plain Text
200次AI调用/天
```

---

Master：

```Plain Text
无限
```

---

# 限流设计

聊天：

```Plain Text
5 req/min
```

---

策略生成：

```Plain Text
20 req/day
```

---

组合生成：

```Plain Text
20 req/day
```

---

# 部署架构

```Plain Text
AI Gateway
      ↓
AI Service
      ↓
Prompt Service
      ↓
Vector DB(Qdrant)
      ↓
LLM Router
      ↓
OpenAI/Claude/Qwen
```

---

# Kubernetes部署

```Plain Text
ai-api

3 Pods
```

---

```Plain Text
ai-worker

10 Pods
```

---

```Plain Text
embedding-worker

5 Pods
```

---

```Plain Text
rag-worker

3 Pods
```

---

# 性能指标

普通问答：

```Plain Text
P95 < 5s
```

---

策略生成：

```Plain Text
P95 < 15s
```

---

回测分析：

```Plain Text
P95 < 20s
```

---

Embedding生成：

```Plain Text
1000 docs/min
```

---

# 可观测性

Metrics：

```Plain Text
ai_task_total

ai_task_success_total

ai_task_failed_total

llm_token_usage_total

embedding_total
```

---

Trace：

```Plain Text
OpenTelemetry
```

---

日志：

```Plain Text
JSON Structured Log
```

---

# 成本控制设计

新增表：

```Plain Text
ai_usage
```

```Plain Text
CREATE TABLE ai_usage
(
    user_id BIGINT,

    provider VARCHAR(64),

    prompt_tokens BIGINT,

    completion_tokens BIGINT,

    cost DECIMAL(18,6),

    created_at DATETIME
);
```

---

统计：

```Plain Text
用户成本

模型成本

任务成本

日成本
```

---

# AI安全设计

输入检测：

```Plain Text
Prompt Injection

越权访问

敏感数据泄露
```

---

输出检测：

```Plain Text
违规内容

金融误导

敏感信息
```

---

# MVP范围

必须实现：

✅ AI生成策略

✅ AI解释策略

✅ AI生成组合

✅ AI分析回测

✅ AI聊天

✅ Prompt模板

✅ Vector DB

✅ Embedding

✅ 成本统计

---

暂缓：

❌ Multi\-Agent

❌ 自动策略研究员

❌ 自动因子挖掘

❌ AutoML

❌ 实盘AI交易

---

# 后续演进

V2 升级为 QuantLab Copilot Agent：

```Plain Text
AI Planner → Tool Calling → Strategy/Portfolio/Backtest/MarketData Services → LLM → Response
```

用户通过自然语言即可完成策略生成、回测、分析、优化的完整闭环。

---

# QuantLab 核心服务 TD 清单

```Plain Text
user-service               ✅
strategy-service           ✅
formula-engine             ✅
portfolio-service          ✅
market-data-service        ✅
backtest-engine            ✅
ranking-service            ✅
community-service          ✅
notification-service       ✅
ai-service                 ✅
billing-service            ✅
```




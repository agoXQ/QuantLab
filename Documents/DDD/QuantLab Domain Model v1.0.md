# QuantLab Domain Model v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Architecture Team

Purpose:

定义 QuantLab 平台统一领域模型。

所有：

```Plain Text
PRD
TD
API
事件
数据库
```

均必须遵循本模型。

---

# 平台领域架构

QuantLab整体划分为：

```Plain Text
Core Domain
Supporting Domain
Infrastructure Domain
```

---

# Core Domain

核心竞争力领域

```Plain Text
Strategy Domain

Portfolio Domain

Backtest Domain

Ranking Domain
```

---

这些决定：

```Plain Text
QuantLab的核心价值
```

---

# Supporting Domain

支撑领域

```Plain Text
Community Domain

AI Domain

Notification Domain

Billing Domain
```

---

# Infrastructure Domain

基础设施领域

```Plain Text
User Domain

Market Data Domain
```

---

# 总体领域图

```Plain Text
User
 │
 │ owns
 ▼

Strategy
 │
 │ compose
 ▼

Portfolio
 │
 │ execute
 ▼

Backtest
 │
 │ evaluate
 ▼

Ranking
 │
 │ expose
 ▼

Community
 │
 │ notify
 ▼

Notification

AI
 ↑
 │
 consume all
```

---

# 核心聚合根

平台统一聚合根：

```Plain Text
User

Strategy

Portfolio

Backtest

Ranking

Content

Notification

AI Task

Security

Billing

```

---

# User Domain

服务：

```Plain Text
User Service
```

---

聚合根：

```Plain Text
User
```

---

职责：

```Plain Text
身份

认证

授权

会员体系
```

---

实体：

```Plain Text
User

Role

Membership
```

---

# User

```Plain Text
type User struct {
    UserID
    Email
    Nickname
    Status
}
```

---

# Strategy Domain

服务：

```Plain Text
Strategy Service

Formula Engine
```

---

聚合根：

```Plain Text
Strategy
```

---

职责：

```Plain Text
策略管理

公式管理

版本管理
```

---

# Strategy

```Plain Text
type Strategy struct {
    ID
    OwnerID
    Name
    Formula
    Version
}
```

---

# Strategy Version

```Plain Text
type StrategyVersion struct {
    Version
    DSL
}
```

---

# Portfolio Domain

新增核心领域

服务：

```Plain Text
Portfolio Service
```

---

聚合根：

```Plain Text
Portfolio
```

---

职责：

```Plain Text
策略组合

资产配置

调仓管理
```

---

# Portfolio

```Plain Text
type Portfolio struct {
    ID
    OwnerID
    Name
    Visibility
}
```

---

# PortfolioItem

```Plain Text
type PortfolioItem struct {
    StrategyID
    Weight
}
```

---

# Portfolio Rebalance

```Plain Text
type RebalanceRule struct {
    Frequency
    Method
}
```

---

# Backtest Domain

服务：

```Plain Text
Backtest Engine
```

---

聚合根：

```Plain Text
Backtest
```

---

职责：

```Plain Text
策略回测

组合回测
```

---

# Backtest

```Plain Text
type Backtest struct {
    ID
    ObjectType
    ObjectID
}
```

---

ObjectType：

```Plain Text
STRATEGY

PORTFOLIO
```

---

# BacktestReport

```Plain Text
type BacktestReport struct {
    Return
    Sharpe
    Drawdown
}
```

---

# Ranking Domain

服务：

```Plain Text
Ranking Service
```

---

聚合根：

```Plain Text
Ranking
```

---

职责：

```Plain Text
评分

排名

可信度
```

---

# Ranking

```Plain Text
type Ranking struct {
    RankingType
    Period
}
```

---

# Score

```Plain Text
type Score struct {
    TotalScore
    TrustScore
}
```

---

# Ranking Object

统一对象：

```Plain Text
STRATEGY

PORTFOLIO

AUTHOR
```

---

# Community Domain

服务：

```Plain Text
Community Service
```

---

聚合根：

```Plain Text
Content
```

---

职责：

```Plain Text
内容分发

互动

关注
```

---

# Content

```Plain Text
type Content struct {
    ID
    Type
    ObjectID
}
```

---

# Content Type

```Plain Text
STRATEGY

PORTFOLIO

ARTICLE
```

---

# Comment

```Plain Text
type Comment struct {
    ID
    ContentID
}
```

---

# Follow

```Plain Text
type Follow struct {
    FollowerID
    FolloweeID
}
```

---

# Notification Domain

服务：

```Plain Text
Notification Service
```

---

聚合根：

```Plain Text
Notification
```

---

职责：

```Plain Text
站内信

邮件

Push
```

---

# Notification

```Plain Text
type Notification struct {
    UserID
    Type
    Payload
}
```

---

# Notification Type

```Plain Text
LIKE

COMMENT

FOLLOW

RANKING

SYSTEM
```

---

# AI Domain

服务：

```Plain Text
AI Service
```

---

聚合根：

```Plain Text
AITask
```

---

职责：

```Plain Text
公式生成

策略分析

组合生成

风险解释
```

---

# AITask

```Plain Text
type AITask struct {
    ID
    TaskType
}
```

---

# Task Type

```Plain Text
GENERATE_STRATEGY

EXPLAIN_STRATEGY

GENERATE_PORTFOLIO

OPTIMIZE_PORTFOLIO

ANALYZE_BACKTEST
```

---

# Billing Domain

服务：

```Plain Text
Billing Service
```

---

聚合根：

```Plain Text
Order
```

---

职责：

```Plain Text
会员订阅

策略订阅

支付管理

创作者收益

结算管理
```

---

实体：

```Plain Text
Order

Payment

Membership

Subscription

RevenueShare

Settlement
```

---

# Order

```Plain Text
type Order struct {
    ID
    UserID
    Amount
    Currency
    Status
}
```

---

# Payment

```Plain Text
type Payment struct {
    ID
    OrderID
    Channel
    Status
}
```

---

# Membership

```Plain Text
type Membership struct {
    UserID
    Tier
    Status
    ExpiredAt
}
```

---

# Subscription

```Plain Text
type Subscription struct {
    ID
    UserID
    ResourceType
    ResourceID
    Status
}
```

---

# RevenueShare

```Plain Text
type RevenueShare struct {
    OrderID
    CreatorID
    PlatformAmount
    CreatorAmount
}
```

---

# Settlement

```Plain Text
type Settlement struct {
    ID
    CreatorID
    Amount
    Status
}
```

---

# Market Data Domain

服务：

```Plain Text
Market Data Service
```

---

聚合根：

```Plain Text
Security
```

---

职责：

```Plain Text
行情

财务

因子
```

---

# Security

```Plain Text
type Security struct {
    Code
    Name
}
```

---

# MarketBar

```Plain Text
type MarketBar struct {
    TradeDate
    Close
}
```

---

# FinancialStatement

```Plain Text
type FinancialStatement struct {
    Revenue
    NetProfit
}
```

---

# 平台统一对象类型

整个系统统一：

```Plain Text
USER

STRATEGY

PORTFOLIO

BACKTEST

CONTENT

COMMENT

AUTHOR

ORDER

MEMBERSHIP
```

---

所有服务必须使用：

```Plain Text
ObjectType
```

统一枚举。

---

# 平台统一状态模型

统一状态：

```Plain Text
DRAFT

PUBLISHED

ARCHIVED

DELETED
```

---

适用于：

```Plain Text
Strategy

Portfolio

Content
```

---

# 平台统一可见性

统一：

```Plain Text
PRIVATE

PUBLIC

UNLISTED
```

---

适用于：

```Plain Text
Strategy

Portfolio

Content
```

---

# 平台统一事件模型

所有事件：

```Plain Text
type DomainEvent struct {
    EventID
    EventType
    AggregateType
    AggregateID
    OccurredAt
}
```

---

# AggregateType

统一：

```Plain Text
USER

STRATEGY

PORTFOLIO

BACKTEST

CONTENT

RANKING

BILLING
```

---

# 平台统一事件总线

Kafka Topic：

```Plain Text
user-events

strategy-events

portfolio-events

backtest-events

ranking-events

community-events

notification-events

ai-events

billing-events
```

---

# 统一ID规范

建议：

```Plain Text
Snowflake
```

---

ID格式：

```Plain Text
Long
```

统一。

---

# 统一审计字段

所有聚合：

```Plain Text
CreatedAt

UpdatedAt

CreatedBy

UpdatedBy
```

---

# 统一软删除

统一：

```Plain Text
DeletedAt
```

---

禁止：

```Plain Text
物理删除
```

---

# 平台统一标签系统

未来新增：

```Plain Text
Tag
```

---

支持：

```Plain Text
Strategy Tag

Portfolio Tag

Content Tag
```

---

# 平台统一收藏系统

未来统一：

```Plain Text
Favorite
```

支持：

```Plain Text
STRATEGY

PORTFOLIO

CONTENT
```

---

# 平台统一评分系统

未来统一：

```Plain Text
Rating
```

支持：

```Plain Text
Strategy

Portfolio
```

---

# QuantLab v1最终服务版图

```Plain Text
quantlab
│
├── user-service
│
├── strategy-service
│
├── portfolio-service
│
├── formula-engine
│
├── market-data-service
│
├── backtest-engine
│
├── ranking-service
│
├── community-service
│
├── notification-service
│
├── ai-service
│
└── billing-service
```

---

# v1领域依赖图

```Plain Text
User
 │
 ▼
Strategy
 │
 ▼
Portfolio
 │
 ▼
Backtest
 │
 ▼
Ranking
 │
 ▼
Community
 │
 ▼
Notification

AI
 ↑
 │
 └───────────────────
     Consume All

Billing
 ↑
 │
 ├── User (membership)
 │
 ├── Strategy (subscription)
 │
 └── Portfolio (subscription)
```

---

# 未来领域预留（V2+）

## Subscription Domain

V1 中订阅能力归属 Billing Domain。

当策略订阅市场、组合订阅市场、创作者经济达到一定规模后，可独立为 Subscription Domain。

负责：

```Plain Text
策略订阅
组合订阅
付费内容
作者收益
```

---

## Execution Domain

面向实盘与模拟盘场景。

负责：

```Plain Text
模拟盘
实盘
券商接口
自动跟投
```

---

## 长期领域版图

```Plain Text
User
Strategy
Portfolio
Backtest
Ranking
Community
Subscription
Execution
Notification
AI
MarketData
Billing
```

当前领域模型可支撑 QuantLab 从 MVP 演进至百万用户级平台，无需推翻核心领域边界。


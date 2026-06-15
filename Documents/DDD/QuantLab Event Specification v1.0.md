# QuantLab Event Specification v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Architecture Team

Scope:

适用于 QuantLab 全部微服务：

- User Service

- Strategy Service

- Portfolio Service

- Backtest Engine

- Ranking Service

- Community Service

- Notification Service

- AI Service

- Market Data Service

- Billing Service

---

# 设计目标

建立统一事件驱动架构（EDA）。

所有跨服务通信必须优先采用事件。

禁止：

同步RPC级联调用

例如：

Strategy Service

→ Ranking Service

→ Community Service

→ Notification Service

应改为：

Strategy Service

↓

StrategyPublished Event

↓

Ranking / Community / Notification

---

# 事件设计原则

## Principle 1

事件描述已经发生的事实

正确：

StrategyPublished

错误：

PublishStrategy

---

## Principle 2

事件不可修改

事件一旦发布：

禁止更新

禁止撤回

只允许追加新事件

---

## Principle 3

事件必须可重放

所有消费者：

必须支持历史事件重放

---

## Principle 4

事件必须幂等

同一个事件：

消费一次

消费十次

结果必须一致

---

## Principle 5

事件必须版本化

所有事件：

必须带版本号

---

# Kafka Topic规范

统一命名：

\-events

例如：

user\-events

strategy\-events

portfolio\-events

backtest\-events

ranking\-events

community\-events

notification\-events

ai\-events

---

# Topic Ownership

一个Topic只能有一个Owner

例如：

strategy\-events

Owner：

Strategy Service

其他服务：

只允许消费

禁止写入

---

# Event Envelope

所有事件统一包装：

\{

"event\_id": "",

"event\_type": "",

"event\_version": "1\.0",

"aggregate\_type": "",

"aggregate\_id": "",

"occurred\_at": "",

"producer": "",

"trace\_id": "",

"payload": \{\}

\}

---

# Envelope字段定义

event\_id

全局唯一ID

推荐：

UUID v7

---

event\_type

事件类型

例如：

StrategyPublished

---

event\_version

事件版本

例如：

1\.0

---

aggregate\_type

聚合根类型

例如：

STRATEGY

PORTFOLIO

CONTENT

---

aggregate\_id

聚合根ID

---

occurred\_at

事件发生时间

UTC

---

producer

生产者服务

例如：

strategy\-service

---

trace\_id

OpenTelemetry TraceID

用于全链路追踪

---

payload

业务数据

---

# Aggregate Type规范

统一枚举：

USER

STRATEGY

PORTFOLIO

BACKTEST

CONTENT

COMMENT

RANKING

NOTIFICATION

AI\_TASK

SECURITY

ORDER

MEMBERSHIP

SUBSCRIPTION

PAYMENT

REVENUE\_SHARE

SETTLEMENT

---

# Event Naming规范

格式：

例如：

StrategyCreated

StrategyPublished

PortfolioUpdated

CommentCreated

UserFollowed

禁止：

CreateStrategy

PublishPortfolio

FollowUser

---

# Version规范

事件Schema变化：

新增字段：

Minor Version

例如：

1\.0 → 1\.1

---

不兼容变更：

Major Version

例如：

1\.0 → 2\.0

---

# 幂等规范

消费者必须维护：

processed\_event

表

CREATE TABLE processed\_event

\(

event\_id VARCHAR\(128\) PRIMARY KEY,

processed\_at DATETIME

\);

消费前：

检查event\_id

已处理：

直接ACK

---

# 重试规范

失败：

自动重试

指数退避：

1s

5s

30s

300s

---

最大：

5次

---

# DLQ规范

重试失败：

进入：

Dead Letter Queue

例如：

strategy\-events\-dlq

portfolio\-events\-dlq

---

# 顺序性规范

同一Aggregate：

必须保证顺序

例如：

PortfolioUpdated

PortfolioPublished

不能乱序

---

Kafka Key：

aggregate\_id

---

# Event Schema Registry

统一维护：

schema\-registry

管理：

事件结构

版本

兼容性

---

# User Domain Events

UserCreated

UserUpdated

UserDeleted

MembershipUpgraded

MembershipExpired

UserBanned

---

# Strategy Domain Events

StrategyCreated

StrategyUpdated

StrategyVersionCreated

StrategyPublished

StrategyArchived

StrategyDeleted

---

# Portfolio Domain Events

PortfolioCreated

PortfolioUpdated

PortfolioDeleted

PortfolioVersionCreated

PortfolioPublished

PortfolioItemAdded

PortfolioItemRemoved

PortfolioWeightsUpdated

PortfolioRebalanced

---

# Backtest Domain Events

BacktestRequested

BacktestStarted

BacktestCompleted

BacktestFailed

---

# Ranking Domain Events

RankingCalculated

RankingUpdated

TopStrategyChanged

TopPortfolioChanged

---

# Community Domain Events

ContentCreated

ContentPublished

ContentLiked

ContentFavorited

CommentCreated

UserFollowed

---

# Notification Domain Events

NotificationCreated

NotificationSent

NotificationFailed

NotificationRead

---

# AI Domain Events

AITaskCreated

AITaskStarted

AITaskCompleted

AITaskFailed

StrategyGenerated

PortfolioGenerated

---

# Billing Domain Events

MembershipCreated

MembershipActivated

MembershipRenewed

MembershipExpired

MembershipCancelled

---

SubscriptionCreated

SubscriptionActivated

SubscriptionRenewed

SubscriptionExpired

SubscriptionCancelled

---

OrderCreated

OrderPaid

OrderFailed

OrderCancelled

OrderRefunded

---

PaymentInitiated

PaymentSucceeded

PaymentFailed

PaymentRefunded

---

RevenueCalculated

RevenueAdjusted

RevenueSettled

---

SettlementRequested

SettlementApproved

SettlementRejected

SettlementPaid

---

CouponCreated

CouponApplied

CouponExpired

---

InvoiceRequested

InvoiceIssued

---

# Saga规范

跨服务业务：

采用 Saga

---

案例：

Portfolio Publish Saga

PortfolioPublished

↓

BacktestRequested

↓

BacktestCompleted

↓

RankingUpdated

↓

ContentCreated

↓

NotificationCreated

---

# Event Flow

策略发布：

StrategyCreated

↓

StrategyPublished

↓

Community ContentCreated

↓

NotificationCreated

---

组合发布：

PortfolioPublished

↓

BacktestRequested

↓

BacktestCompleted

↓

RankingUpdated

↓

Community ContentCreated

---

会员购买 Saga：

OrderCreated

↓

PaymentSucceeded

↓

OrderPaid

↓

MembershipActivated

↓

NotificationCreated

---

策略订阅 Saga：

OrderCreated

↓

PaymentSucceeded

↓

OrderPaid

↓

SubscriptionActivated

↓

RevenueCalculated

↓

NotificationCreated

---

退款 Saga：

RefundRequested

↓

PaymentRefunded

↓

OrderRefunded

↓

SubscriptionCancelled

↓

RevenueAdjusted

---

# AI事件消费规范

AI Service：

只消费

禁止修改源数据

例如：

StrategyCreated

↓

AI分析

生成：

AITaskCompleted

而非：

修改Strategy

---

# Trace规范

所有事件：

必须携带：

trace\_id

span\_id

支持：

OpenTelemetry

---

# Metrics规范

统一指标：

event\_produced\_total

event\_consumed\_total

event\_failed\_total

event\_retry\_total

event\_dlq\_total

---

# SLA

事件投递成功率：

99\.99%

---

事件延迟：

P95 \< 1s

P99 \< 5s

---

DLQ率：

\< 0\.1%

---

# MVP范围

必须实现：

统一Envelope

统一Topic

统一Version

统一Idempotency

统一Retry

统一DLQ

统一Trace

统一Schema Registry

---

# 长期演进

v2：

Event Sourcing

---

v3：

CQRS \+ Event Sourcing

---

v4：

Workflow Engine（Temporal）


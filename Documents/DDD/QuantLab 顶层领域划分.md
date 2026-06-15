# QuantLab 顶层领域划分

# 模型领域overview

```Plain Text
QuantLab
│
├── Identity Domain
│
├── Strategy Domain
│
├── Formula Domain
│
├── Market Domain
│
├── Backtest Domain
│
├── Ranking Domain
│
├── Community Domain
│
├── AI Domain
│
├── Notification Domain
│
├── Portfolio Domain
│
└── Billing Domain
```

对应：

```Plain Text
User Service
Strategy Service
Formula Engine
Market Data Service
Backtest Engine
Ranking Service
Community Service
AI Service
Notification Service
Portfolio Service
Billing Service
```

---

# 服务边界原则

非常重要：

## User Service拥有

```Plain Text
User

Profile

Membership

Role

Permission

Follow
```

---

其它服务禁止保存：

```Plain Text
nickname
avatar
membership
```

只能缓存快照。

---

## Strategy Service拥有

```Plain Text
Strategy

StrategyVersion

StrategyTag

StrategyPublish
```

---

只有它能修改策略。

---

## Formula Engine拥有

```Plain Text
Formula

AST

Parser

Compiler
```

---

不保存策略。

---

## Market Data拥有

```Plain Text
Security

KLine

Factor

Financial

Index
```

---

所有服务都不能直接访问第三方数据源。

---

## Backtest Engine拥有

```Plain Text
BacktestJob

Portfolio

Position

Order

Trade

Report
```

---

只有回测引擎能生成交易记录。

---

## Ranking Service拥有

```Plain Text
Ranking

RankingSnapshot

Score

TrustScore
```

---

只读其它服务。

---

## Community拥有

```Plain Text
Post

Article

Comment

Like

Favorite

Feed
```

---

## AI拥有

```Plain Text
Prompt

Conversation

ResearchTask

AIReport
```

---

## Notification拥有

```Plain Text
Notification

Subscription

Template

DeliveryRecord
```

---

## Portfolio Service拥有

```Plain Text
Portfolio

PortfolioItem

RebalanceRule

PortfolioPublish
```

---

只有它能修改组合。

---

## Billing Service拥有

```Plain Text
Membership

Order

Payment

RevenueShare

Settlement

Coupon

Invoice
```

---

只有它能处理支付与结算。

---

# QuantLab统一事件总线

建议直接建立：

```Plain Text
quantlab-event-bus
```

Kafka Topic体系：

```Plain Text
user-events

strategy-events

backtest-events

ranking-events

community-events

ai-events

notification-events

portfolio-events

billing-events
```

---

# Event Naming Standard

统一格式：

```Plain Text
<Entity><Action>
```

例如：

```Plain Text
UserCreated

StrategyPublished

BacktestFinished

RankingEntered
```

---

不要：

```Plain Text
create_user

strategy_publish_success
```

---

# User Service Event Model

---

UserCreated

用户注册

```Plain Text
{
  "event":"UserCreated",
  "user_id":"u123"
}
```

---

UserProfileUpdated

---

MembershipUpgraded

---

MembershipExpired

---

UserFollowed

```Plain Text
{
  "follower_id":"u1",
  "following_id":"u2"
}
```

---

UserVerified

---

# Strategy Service Event Model

---

StrategyCreated

---

StrategyUpdated

---

StrategyVersionCreated

---

StrategyPublished

重要事件

```Plain Text
{
  "strategy_id":"s001",
  "author_id":"u123"
}
```

---

StrategyArchived

---

StrategyDeleted

---

# Formula Engine Event

Formula 编译失败一般无需广播。

因此只建议：

---

FormulaValidated

---

FormulaCompilationFailed

---

主要用于审计。

---

# Market Data Event

---

MarketDataUpdated

---

FactorUpdated

---

FinancialUpdated

---

TradingCalendarUpdated

---

DataVersionCreated

非常重要

```Plain Text
{
  "version":"2026.06.15"
}
```

---

# Backtest Event Model

这是核心。

---

BacktestCreated

---

BacktestStarted

---

BacktestFinished

最重要

```Plain Text
{
  "backtest_id":"b001",
  "strategy_id":"s001",
  "return":0.42
}
```

---

BacktestFailed

---

BacktestReportGenerated

---

BacktestShared

---

# Ranking Event Model

---

RankingCalculated

---

RankingSnapshotCreated

---

StrategyEnteredRanking

```Plain Text
{
  "strategy_id":"s001",
  "ranking":"return"
}
```

---

StrategyDroppedRanking

---

TrustScoreUpdated

---

# Community Event Model

---

PostCreated

---

ArticlePublished

---

CommentCreated

---

CommentReplied

---

StrategyLiked

---

StrategyFavorited

---

AuthorFollowed

---

FeedGenerated

---

# AI Event Model

---

AIConversationStarted

---

StrategyGeneratedByAI

---

BacktestAnalyzedByAI

---

AIReportGenerated

```Plain Text
{
  "report_id":"r001"
}
```

---

AIOptimizationCompleted

---

# Portfolio Event Model

---

PortfolioCreated

---

PortfolioUpdated

---

PortfolioPublished

重要事件

```Plain Text
{
  "portfolio_id":"p001",
  "owner_id":"u123"
}
```

---

PortfolioItemAdded

---

PortfolioItemRemoved

---

PortfolioWeightsUpdated

---

PortfolioRebalanced

---

PortfolioBacktestCompleted

---

# Billing Event Model

---

MembershipActivated

```Plain Text
{
  "user_id":"u123",
  "tier":"PRO"
}
```

---

MembershipExpired

---

OrderPaid

最核心事件

```Plain Text
{
  "order_id":"ord_001",
  "user_id":"u123",
  "amount":99.9
}
```

---

OrderRefunded

---

RevenueCalculated

```Plain Text
{
  "order_id":"ord_001",
  "creator_id":"u456",
  "platform_amount":20,
  "creator_amount":80
}
```

---

SettlementPaid

---

# Notification Event Model

通知一般消费事件。

但仍然需要：

---

NotificationCreated

---

NotificationSent

---

NotificationDelivered

---

NotificationRead

---

NotificationClicked

---

# 全平台核心业务流

这是最重要的部分。

---

## 流程1

策略发布

```Plain Text
StrategyCreated
        ↓
StrategyPublished
        ↓
Community生成策略动态
        ↓
Notification发送关注提醒
```

---

## 流程2

回测完成

```Plain Text
BacktestFinished
        ↓
Ranking重新计算
        ↓
生成榜单
        ↓
触发通知
```

---

## 流程3

策略进入排行榜

```Plain Text
StrategyEnteredRanking
        ↓
Community曝光
        ↓
Notification通知作者
```

---

## 流程4

AI生成策略

```Plain Text
StrategyGeneratedByAI
        ↓
StrategyCreated
        ↓
BacktestCreated
        ↓
BacktestFinished
```

---

## 流程5

组合发布与回测

```Plain Text
PortfolioCreated
        ↓
PortfolioPublished
        ↓
BacktestCreated
        ↓
BacktestFinished
        ↓
Ranking重新计算
        ↓
Community曝光
```

---

## 流程6

会员购买

```Plain Text
OrderCreated
        ↓
PaymentSucceeded
        ↓
OrderPaid
        ↓
MembershipActivated
        ↓
Notification通知用户
```

---

## 流程7

创作者收益结算

```Plain Text
OrderPaid
        ↓
RevenueCalculated
        ↓
SettlementRequested
        ↓
SettlementApproved
        ↓
SettlementPaid
        ↓
Notification通知作者
```

---




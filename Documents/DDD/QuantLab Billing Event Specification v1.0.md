# QuantLab Billing Event Specification v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Platform Architecture Team

Related Documents:

- Billing Service PRD

- QuantLab Monetization Model v1\.0

- QuantLab Domain Model v1\.0

- QuantLab Event Specification v1\.0

---

# 文档目标

定义 Billing Domain 的统一事件模型。

统一：

- 会员事件

- 订阅事件

- 订单事件

- 支付事件

- 分账事件

- 结算事件

- 退款事件

用于：

- 服务解耦

- Saga事务编排

- 数据同步

- 通知触发

- 排行榜统计

- 收益计算

---

# Event Design Principles

## 原则1

事件描述：

过去发生的事实

正确：

```Plain Text
OrderPaid
```

错误：

```Plain Text
PayOrder
```

---

## 原则2

事件不可修改

事件一旦发布：

```Plain Text
Immutable
```

---

## 原则3

事件可重放

支持：

```Plain Text
Replay
```

---

## 原则4

事件最终一致

采用：

```Plain Text
Eventual Consistency
```

---

# Kafka Topic规划

Billing统一Topic：

```Plain Text
billing-events
```

---

DLQ：

```Plain Text
billing-events-dlq
```

---

未来扩展：

```Plain Text
billing-membership-events

billing-payment-events

billing-settlement-events
```

---

# Event Envelope

所有事件统一结构：

```JSON
{
  "event_id": "uuid",
  "event_type": "OrderPaid",
  "aggregate_type": "Order",
  "aggregate_id": "ord_xxx",
  "version": 1,
  "occurred_at": "2026-01-01T00:00:00Z",
  "producer": "billing-service",
  "trace_id": "trace_xxx",
  "data": {}
}
```

---

# Membership Events

Aggregate：

```Plain Text
Membership
```

---

# MembershipCreated

触发：

首次创建会员

---

Payload：

```JSON
{
  "membership_id": "",
  "user_id": "",
  "tier": "PRO"
}
```

---

消费者：

```Plain Text
User Service
Notification Service
```

---

# MembershipActivated

触发：

支付成功

会员正式生效

---

Payload：

```JSON
{
  "membership_id": "",
  "user_id": "",
  "tier": "MASTER"
}
```

---

消费者：

```Plain Text
Permission Service
User Service
```

---

# MembershipRenewed

触发：

自动续费成功

---

Payload：

```JSON
{
  "membership_id": "",
  "user_id": "",
  "expired_at": ""
}
```

---

# MembershipExpired

触发：

会员过期

---

消费者：

```Plain Text
Permission Service
Notification Service
```

---

# MembershipCancelled

触发：

用户取消自动续费

---

# Subscription Events

Aggregate：

```Plain Text
Subscription
```

---

# SubscriptionCreated

触发：

用户创建订阅

---

Payload：

```JSON
{
  "subscription_id": "",
  "resource_type": "STRATEGY",
  "resource_id": "",
  "subscriber_id": ""
}
```

---

# SubscriptionActivated

触发：

支付成功

---

消费者：

```Plain Text
Strategy Service
Portfolio Service
Notification Service
```

---

# SubscriptionRenewed

触发：

自动续费成功

---

# SubscriptionExpired

触发：

订阅过期

---

消费者：

```Plain Text
Strategy Service
Portfolio Service
Permission Service
```

---

# SubscriptionCancelled

触发：

用户取消订阅

---

# Order Events

Aggregate：

```Plain Text
Order
```

---

# OrderCreated

触发：

订单创建

---

Payload：

```JSON
{
  "order_id": "",
  "user_id": "",
  "amount": 99.9,
  "currency": "USD"
}
```

---

消费者：

```Plain Text
Payment Service
```

---

# OrderPaid

最核心事件。

---

触发：

支付成功

---

Payload：

```JSON
{
  "order_id": "",
  "user_id": "",
  "amount": 99.9,
  "currency": "USD"
}
```

---

消费者：

```Plain Text
Membership
Subscription
Revenue
Notification
Ranking
```

---

# OrderFailed

触发：

支付失败

---

# OrderCancelled

触发：

订单取消

---

# OrderRefunded

触发：

退款成功

---

消费者：

```Plain Text
Subscription
Membership
Revenue
Notification
```

---

# Payment Events

Aggregate：

```Plain Text
Payment
```

---

# PaymentInitiated

触发：

调用支付网关

---

# PaymentSucceeded

触发：

第三方支付回调成功

---

Payload：

```JSON
{
  "payment_id": "",
  "order_id": "",
  "channel": "STRIPE",
  "provider_txn_id": ""
}
```

---

消费者：

```Plain Text
Order Service
```

---

# PaymentFailed

触发：

支付失败

---

# PaymentRefunded

触发：

退款完成

---

# Revenue Events

Aggregate：

```Plain Text
RevenueShare
```

---

# RevenueCalculated

触发：

订单支付成功

分账计算完成

---

Payload：

```JSON
{
  "order_id": "",
  "creator_id": "",
  "gross_amount": 100,
  "platform_amount": 20,
  "creator_amount": 80
}
```

---

消费者：

```Plain Text
Creator Center
Settlement Service
```

---

# RevenueAdjusted

触发：

退款

风控处罚

收益修正

---

# RevenueSettled

触发：

收益结算完成

---

# Settlement Events

Aggregate：

```Plain Text
Settlement
```

---

# SettlementRequested

触发：

作者申请提现

---

Payload：

```JSON
{
  "settlement_id": "",
  "creator_id": "",
  "amount": 1000
}
```

---

消费者：

```Plain Text
Finance Service
Risk Service
```

---

# SettlementApproved

触发：

审核通过

---

# SettlementRejected

触发：

审核拒绝

---

# SettlementPaid

触发：

提现完成

---

消费者：

```Plain Text
Notification Service
Creator Center
```

---

# Coupon Events

Aggregate：

```Plain Text
Coupon
```

---

# CouponCreated

触发：

创建优惠券

---

# CouponApplied

触发：

订单使用优惠券

---

# CouponExpired

触发：

优惠券过期

---

# Invoice Events

Aggregate：

```Plain Text
Invoice
```

---

# InvoiceRequested

触发：

用户申请发票

---

# InvoiceIssued

触发：

发票开具成功

---

消费者：

```Plain Text
Notification Service
```

---

# Billing Saga

Billing存在多个跨服务事务。

---

# Membership Purchase Saga

流程：

```Plain Text
Create Order
    ↓

Payment Succeeded
    ↓

Order Paid
    ↓

Membership Activated
    ↓

Permission Updated
    ↓

Notification Sent
```

---

补偿：

```Plain Text
Payment Failed
     ↓

Order Cancelled
```

---

# Strategy Subscription Saga

流程：

```Plain Text
Create Order
      ↓

Payment Success
      ↓

Subscription Activated
      ↓

Revenue Calculated
      ↓

Notification Sent
```

---

# Refund Saga

流程：

```Plain Text
Refund Requested
      ↓

Refund Approved
      ↓

Payment Refunded
      ↓

Order Refunded
      ↓

Subscription Cancelled
      ↓

Revenue Adjusted
```

---

# Auto Renew Saga

流程：

```Plain Text
Scheduler
    ↓

Create Renewal Order
    ↓

Payment Success
    ↓

Subscription Renewed
```

---

# Event Versioning

采用：

```Plain Text
Schema Versioning
```

---

例如：

```JSON
{
  "version": 2
}
```

---

兼容原则：

```Plain Text
只能增加字段

不能删除字段
```

---

# 幂等设计

每个事件：

```Plain Text
event_id
```

全局唯一。

---

消费表：

```SQL
processed_events
```

记录：

```Plain Text
event_id
consumer
processed_at
```

---

# Retry Policy

失败：

```Plain Text
3次立即重试
```

---

然后：

```Plain Text
指数退避
```

---

最终：

```Plain Text
DLQ
```

---

# DLQ策略

Topic：

```Plain Text
billing-events-dlq
```

---

保存：

```Plain Text
30天
```

---

# Audit Events

重要事件：

```Plain Text
PaymentSucceeded

PaymentRefunded

RevenueSettled

SettlementPaid
```

必须永久保存。

---

# Event Metrics

监控：

```Plain Text
TPS

Consumer Lag

Retry Count

DLQ Count

Success Rate
```

---

# 与其它服务的集成

User Service：

消费：

```Plain Text
MembershipActivated
MembershipExpired
```

---

Strategy Service：

消费：

```Plain Text
SubscriptionActivated
SubscriptionExpired
```

---

Portfolio Service：

消费：

```Plain Text
SubscriptionActivated
SubscriptionExpired
```

---

Ranking Service：

消费：

```Plain Text
RevenueCalculated
RevenueSettled
```

---

Community Service：

消费：

```Plain Text
SubscriptionActivated
```

---

Notification Service：

消费：

```Plain Text
全部核心事件
```

---

# Billing Event Flow总览

```Plain Text
OrderCreated
      ↓

PaymentSucceeded
      ↓

OrderPaid
      ↓

MembershipActivated
      ↓

SubscriptionActivated
      ↓

RevenueCalculated
      ↓

SettlementRequested
      ↓

SettlementPaid
```

---

# 最终事件地图

```Plain Text
Membership
 ├ MembershipCreated
 ├ MembershipActivated
 ├ MembershipRenewed
 ├ MembershipExpired
 └ MembershipCancelled

Subscription
 ├ SubscriptionCreated
 ├ SubscriptionActivated
 ├ SubscriptionRenewed
 ├ SubscriptionExpired
 └ SubscriptionCancelled

Order
 ├ OrderCreated
 ├ OrderPaid
 ├ OrderFailed
 ├ OrderCancelled
 └ OrderRefunded

Payment
 ├ PaymentInitiated
 ├ PaymentSucceeded
 ├ PaymentFailed
 └ PaymentRefunded

Revenue
 ├ RevenueCalculated
 ├ RevenueAdjusted
 └ RevenueSettled

Settlement
 ├ SettlementRequested
 ├ SettlementApproved
 ├ SettlementRejected
 └ SettlementPaid

Coupon
 ├ CouponCreated
 ├ CouponApplied
 └ CouponExpired

Invoice
 ├ InvoiceRequested
 └ InvoiceIssued
```

# Billing Domain 文档链路

```Plain Text
Billing PRD
        ↓
Monetization Model
        ↓
Domain Model (Billing Domain)
        ↓
Billing Event Specification
```

后续工程落地文档（Billing Service TD）将覆盖：

- 目录结构

- 数据库设计

- API 设计

- 缓存设计

- 权限设计

- 部署架构

- 性能指标


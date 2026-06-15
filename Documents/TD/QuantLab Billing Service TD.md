# QuantLab Billing Service TD

Version: 1\.0

Status: Technical Design

Owner: Platform Architecture Team

Domain: Billing

Related Documents:

- Billing Service PRD

- QuantLab Monetization Model v1\.0

- QuantLab Domain Model v1\.0

- QuantLab Billing Event Specification v1\.0

- QuantLab API Design Standard v1\.0

---

# 服务定位

Billing Service 是 QuantLab 的商业化核心服务。

负责：

```Plain Text
Membership

Subscription

Order

Payment

Revenue Share

Settlement

Coupon

Invoice
```

统一管理：

```Plain Text
会员收费

策略订阅

组合订阅

创作者收益

提现结算
```

---

# 服务边界

## Billing负责

```Plain Text
订单

支付

订阅

会员

分账

结算
```

---

## Billing不负责

用户信息

↓

User Service

---

策略管理

↓

Strategy Service

---

组合管理

↓

Portfolio Service

---

通知发送

↓

Notification Service

---

# 微服务目录结构

```Plain Text
billing-service
│
├── cmd
│   └── server
│
├── internal
│
│   ├── application
│   │
│   ├── domain
│   │
│   ├── infrastructure
│   │
│   ├── interfaces
│   │
│   └── jobs
│
├── api
│
├── configs
│
├── deployments
│
└── scripts
```

---

# DDD结构

## Domain

```Plain Text
domain
│
├── membership
│
├── subscription
│
├── order
│
├── payment
│
├── revenue
│
├── settlement
│
├── coupon
│
└── invoice
```

---

## Application

```Plain Text
application
│
├── membership
│
├── subscription
│
├── payment
│
├── settlement
│
└── revenue
```

---

## Infrastructure

```Plain Text
infrastructure
│
├── postgresql
│
├── redis
│
├── kafka
│
├── stripe
│
├── paypal
│
├── alipay
│
└── wechatpay
```

---

# 数据库设计

数据库：

```SQL
quantlab_billing
```

---

# memberships

```SQL
CREATE TABLE memberships (
    id BIGINT PRIMARY KEY,

    membership_id VARCHAR(64) UNIQUE,

    user_id BIGINT,

    tier VARCHAR(32),

    status VARCHAR(32),

    auto_renew BOOLEAN,

    started_at DATETIME,

    expired_at DATETIME,

    created_at DATETIME,

    updated_at DATETIME
);
```

---

索引：

```SQL
idx_user_id
idx_status
idx_expired_at
```

---

# subscriptions

```SQL
CREATE TABLE subscriptions (
    id BIGINT PRIMARY KEY,

    subscription_id VARCHAR(64),

    subscriber_id BIGINT,

    resource_type VARCHAR(32),

    resource_id VARCHAR(64),

    plan_id VARCHAR(64),

    status VARCHAR(32),

    auto_renew BOOLEAN,

    started_at DATETIME,

    expired_at DATETIME,

    created_at DATETIME
);
```

---

索引：

```SQL
idx_subscriber_id
idx_resource
idx_status
```

---

# orders

核心表。

```SQL
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,

    order_id VARCHAR(64),

    user_id BIGINT,

    order_type VARCHAR(64),

    amount DECIMAL(18,2),

    currency VARCHAR(16),

    status VARCHAR(32),

    coupon_id VARCHAR(64),

    created_at DATETIME,

    paid_at DATETIME
);
```

---

索引：

```SQL
idx_user_id
idx_order_type
idx_status
```

---

# payments

```SQL
CREATE TABLE payments (
    id BIGINT PRIMARY KEY,

    payment_id VARCHAR(64),

    order_id VARCHAR(64),

    channel VARCHAR(32),

    provider_txn_id VARCHAR(128),

    amount DECIMAL(18,2),

    currency VARCHAR(16),

    status VARCHAR(32),

    paid_at DATETIME
);
```

---

唯一索引：

```SQL
uk_provider_txn_id
```

用于防重复回调。

---

# revenue\_shares

```SQL
CREATE TABLE revenue_shares (
    id BIGINT PRIMARY KEY,

    revenue_share_id VARCHAR(64),

    order_id VARCHAR(64),

    creator_id BIGINT,

    gross_amount DECIMAL(18,2),

    platform_amount DECIMAL(18,2),

    creator_amount DECIMAL(18,2),

    status VARCHAR(32),

    created_at DATETIME
);
```

---

# settlements

```SQL
CREATE TABLE settlements (
    id BIGINT PRIMARY KEY,

    settlement_id VARCHAR(64),

    creator_id BIGINT,

    amount DECIMAL(18,2),

    currency VARCHAR(16),

    status VARCHAR(32),

    requested_at DATETIME,

    paid_at DATETIME
);
```

---

# coupons

```SQL
CREATE TABLE coupons (
    id BIGINT PRIMARY KEY,

    coupon_id VARCHAR(64),

    code VARCHAR(64),

    discount_type VARCHAR(32),

    value DECIMAL(18,2),

    max_usage INT,

    used_count INT,

    expired_at DATETIME
);
```

---

# invoices

```SQL
CREATE TABLE invoices (
    id BIGINT PRIMARY KEY,

    invoice_id VARCHAR(64),

    user_id BIGINT,

    order_id VARCHAR(64),

    amount DECIMAL(18,2),

    currency VARCHAR(16),

    status VARCHAR(32),

    issued_at DATETIME
);
```

---

# processed\_events

幂等表。

```SQL
CREATE TABLE processed_events (
    event_id VARCHAR(128),

    consumer VARCHAR(128),

    processed_at DATETIME,

    PRIMARY KEY(event_id, consumer)
);
```

---

# API设计

统一前缀：

```HTTP
/api/v1/billing
```

---

# Membership API

---

## 查询会员

```HTTP
GET /memberships/current
```

---

返回：

```JSON
{
  "tier":"PRO",
  "status":"ACTIVE",
  "expired_at":"..."
}
```

---

## 开通会员

```HTTP
POST /memberships/purchase
```

---

请求：

```JSON
{
  "tier":"PRO",
  "billing_cycle":"YEARLY"
}
```

---

# Subscription API

---

## 创建订阅

```HTTP
POST /subscriptions
```

---

```JSON
{
  "resource_type":"STRATEGY",
  "resource_id":"st_123"
}
```

---

## 我的订阅

```HTTP
GET /subscriptions
```

---

## 取消订阅

```HTTP
POST /subscriptions/{id}/cancel
```

---

# Order API

---

## 创建订单

```HTTP
POST /orders
```

---

## 查询订单

```HTTP
GET /orders/{id}
```

---

## 我的订单

```HTTP
GET /orders
```

---

# Payment API

---

## 创建支付

```HTTP
POST /payments
```

---

返回：

```JSON
{
  "payment_url":"..."
}
```

---

## 支付回调

```HTTP
POST /payments/webhook
```

---

# Revenue API

---

## 收益统计

```HTTP
GET /creator/revenues
```

---

## 收益明细

```HTTP
GET /creator/revenues/details
```

---

# Settlement API

---

## 提现申请

```HTTP
POST /settlements
```

---

## 提现记录

```HTTP
GET /settlements
```

---

# Coupon API

---

## 验证优惠券

```HTTP
POST /coupons/validate
```

---

# Invoice API

---

## 发票申请

```HTTP
POST /invoices
```

---

# Kafka设计

Topic：

```Plain Text
billing-events
```

---

消费者组：

```Plain Text
billing-service

notification-service

ranking-service

user-service

portfolio-service

strategy-service
```

---

# 发布事件

Membership：

```Plain Text
MembershipCreated

MembershipActivated

MembershipExpired
```

---

Subscription：

```Plain Text
SubscriptionCreated

SubscriptionActivated

SubscriptionExpired
```

---

Order：

```Plain Text
OrderCreated

OrderPaid

OrderRefunded
```

---

Payment：

```Plain Text
PaymentSucceeded

PaymentFailed
```

---

Revenue：

```Plain Text
RevenueCalculated

RevenueSettled
```

---

Settlement：

```Plain Text
SettlementRequested

SettlementPaid
```

---

# Redis设计

---

会员缓存

```Plain Text
membership:{user_id}
```

TTL：

```Plain Text
30 min
```

---

订阅缓存

```Plain Text
subscription:{user_id}
```

TTL：

```Plain Text
10 min
```

---

订单缓存

```Plain Text
order:{order_id}
```

TTL：

```Plain Text
5 min
```

---

收益缓存

```Plain Text
creator:revenue:{creator_id}
```

TTL：

```Plain Text
30 min
```

---

# 权限设计

Permission Model：

---

会员：

```Plain Text
billing.membership.purchase
```

---

订阅：

```Plain Text
billing.subscription.create
billing.subscription.cancel
```

---

提现：

```Plain Text
billing.settlement.request
```

---

管理员：

```Plain Text
billing.order.manage

billing.payment.manage

billing.settlement.manage
```

---

# 风控设计

检测：

---

异常支付

---

短时间大量退款

---

自买自卖

---

异常提现

---

重复支付回调

---

策略：

```Plain Text
LIMIT

FREEZE

BAN
```

---

# 自动续费任务

Job：

```Plain Text
membership-renew-job
```

---

执行：

```Plain Text
每天 02:00
```

---

流程：

```Plain Text
扫描即将到期会员

创建续费订单

发起自动扣费
```

---

# 订阅续费任务

Job：

```Plain Text
subscription-renew-job
```

---

流程：

```Plain Text
扫描订阅

自动续费

发送通知
```

---

# 对账任务

Job：

```Plain Text
payment-reconciliation-job
```

---

周期：

```Plain Text
每日
```

---

渠道：

```Plain Text
Stripe

PayPal

Alipay

WeChat
```

---

# 分账任务

Job：

```Plain Text
revenue-calculation-job
```

---

触发：

```Plain Text
OrderPaid
```

---

生成：

```Plain Text
RevenueShare
```

---

# 提现任务

Job：

```Plain Text
settlement-job
```

---

流程：

```Plain Text
审核

打款

更新状态
```

---

# 支付架构

统一支付抽象：

```Go
type PaymentGateway interface {

    CreatePayment()

    QueryPayment()

    Refund()

    VerifyWebhook()
}
```

---

实现：

```Plain Text
StripeGateway

PaypalGateway

AlipayGateway

WechatGateway

PayPayGateway
```

---

# 部署架构

Deployment：

```YAML
billing-service
replicas: 3
```

---

资源：

```YAML
cpu: 1

memory: 2Gi
```

---

# 高可用

PostgreSQL：

```Plain Text
Primary
+
Replica
```

---

Redis：

```Plain Text
Cluster
```

---

Kafka：

```Plain Text
3 Brokers
```

---

# 可观测性

Metrics：

```Plain Text
订单数

支付成功率

退款率

提现金额

MRR

ARR
```

---

Logs：

```Plain Text
支付日志

回调日志

提现日志
```

---

Tracing：

```Plain Text
Order

Payment

Revenue

Settlement
```

---

# SLA

API成功率：

```Plain Text
99.9%
```

---

支付成功率：

```Plain Text
99.95%
```

---

事件投递成功率：

```Plain Text
99.99%
```

---

# 性能指标

订单创建：

```Plain Text
P99 < 100ms
```

---

会员查询：

```Plain Text
P99 < 50ms
```

---

订单查询：

```Plain Text
P99 < 100ms
```

---

支付回调：

```Plain Text
P99 < 200ms
```

---

吞吐：

```Plain Text
1000 TPS
```

---

# MVP 裁剪方案

QuantLab 第一版仅需实现：

```Plain Text
Membership
Order
Payment
Strategy Subscription
Revenue Share
```

暂缓：

```Plain Text
Invoice
Coupon
Enterprise Billing
Tax
Multi Currency Settlement
```

---

# Billing Service 在平台中的位置

```Plain Text
User Service
Strategy Service
Formula Engine
Market Data Service
Backtest Engine
Portfolio Service
Ranking Service
Community Service
Notification Service
AI Service
Billing Service
```

共 11 个核心服务。Billing Service 是平台会员体系、订阅市场、创作者经济、收益分账的统一商业化基础设施。


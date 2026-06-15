# QuantLab Notification Service TD

Version: 1\.0

Module: Notification Service

Priority: P1（Supporting Domain）

Status: Draft

Owner: Architecture Team

Dependencies:

```Plain Text
User Service
Strategy Service
Portfolio Service
Backtest Engine
Ranking Service
Community Service
Event Bus(Kafka)
```

---

# 服务定位

Notification Service 是 QuantLab 的统一消息中心（Unified Notification Center）。

负责：

```Plain Text
站内通知
系统通知
邮件通知
Webhook通知
订阅通知
异步消息分发
通知偏好管理
```

不负责：

```Plain Text
业务逻辑
排行榜计算
社区互动
策略分析
```

Notification Service 本质上是：

```Plain Text
Event Consumer
+
Message Dispatcher
```

---

# 服务职责

Notification Service 监听平台事件：

```Plain Text
StrategyPublished
PortfolioPublished
CommentCreated
ContentLiked
UserFollowed
MembershipExpired
RankingUpdated
```

并转换为：

```Plain Text
In-App Notification
Email
Webhook
Push
```

---

# DDD设计

Notification Domain：

```Plain Text
Notification
Preference
Subscription
DeliveryTask
Template
```

---

# 聚合根

## Notification

```Plain Text
type Notification struct {
    ID

    UserID

    Type

    Title

    Content

    Status

    ReadAt

    CreatedAt
}
```

---

# NotificationType

统一枚举：

```Plain Text
SYSTEM

COMMENT

LIKE

FOLLOW

MENTION

RANKING

STRATEGY

PORTFOLIO

BACKTEST

MEMBERSHIP
```

---

# NotificationStatus

```Plain Text
PENDING

DELIVERED

READ

FAILED
```

---

# NotificationPreference

用户通知偏好

```Plain Text
type NotificationPreference struct {
    UserID

    InAppEnabled

    EmailEnabled

    WebhookEnabled

    PushEnabled
}
```

---

# Subscription

订阅关系

未来支持：

```Plain Text
关注策略

关注组合

关注作者
```

---

```Plain Text
type Subscription struct {
    SubscriberID

    ObjectType

    ObjectID
}
```

---

# DeliveryTask

发送任务

```Plain Text
type DeliveryTask struct {
    ID

    NotificationID

    Channel

    RetryCount

    Status
}
```

---

# Template

通知模板

```Plain Text
type Template struct {
    Type

    TitleTemplate

    ContentTemplate
}
```

---

# 服务目录结构

```Plain Text
notification-service
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
│   │   ├── notification
│   │   ├── preference
│   │   ├── subscription
│   │   └── template
│   │
│   ├── infrastructure
│   │   ├── kafka
│   │   ├── redis
│   │   ├── postgresql
│   │   ├── email
│   │   └── webhook
│   │
│   ├── consumer
│   │
│   ├── dispatcher
│   │
│   └── scheduler
│
├── configs
│
└── deploy
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_notification
```

---

# notification

```Plain Text
CREATE TABLE notification
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT NOT NULL,

    type VARCHAR(64),

    title VARCHAR(256),

    content TEXT,

    status VARCHAR(32),

    read_at DATETIME NULL,

    created_at DATETIME,

    INDEX idx_user(user_id),
    INDEX idx_status(status)
);
```

---

# notification\_preference

```Plain Text
CREATE TABLE notification_preference
(
    user_id BIGINT PRIMARY KEY,

    in_app_enabled BOOLEAN,

    email_enabled BOOLEAN,

    webhook_enabled BOOLEAN,

    push_enabled BOOLEAN,

    updated_at DATETIME
);
```

---

# notification\_subscription

```Plain Text
CREATE TABLE notification_subscription
(
    id BIGINT PRIMARY KEY,

    subscriber_id BIGINT,

    object_type VARCHAR(32),

    object_id BIGINT,

    created_at DATETIME,

    UNIQUE KEY uk_sub(
        subscriber_id,
        object_type,
        object_id
    )
);
```

---

# notification\_delivery

```Plain Text
CREATE TABLE notification_delivery
(
    id BIGINT PRIMARY KEY,

    notification_id BIGINT,

    channel VARCHAR(32),

    status VARCHAR(32),

    retry_count INT,

    created_at DATETIME
);
```

---

# notification\_template

```Plain Text
CREATE TABLE notification_template
(
    id BIGINT PRIMARY KEY,

    type VARCHAR(64),

    title_template TEXT,

    content_template TEXT,

    updated_at DATETIME
);
```

---

# processed\_event

幂等表

```Plain Text
CREATE TABLE processed_event
(
    event_id VARCHAR(128) PRIMARY KEY,

    processed_at DATETIME
);
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/notifications
```

---

# 获取通知列表

```Plain Text
GET /
```

参数：

```Plain Text
?page=1
&page_size=20
```

---

返回：

```Plain Text
{
  "total":120,
  "items":[]
}
```

---

# 获取未读数量

```Plain Text
GET /unread-count
```

---

返回：

```Plain Text
{
  "count":15
}
```

---

# 标记已读

```Plain Text
POST /{id}/read
```

---

# 全部已读

```Plain Text
POST /read-all
```

---

# 删除通知

```Plain Text
DELETE /{id}
```

软删除。

---

# 获取通知偏好

```Plain Text
GET /preferences
```

---

# 更新通知偏好

```Plain Text
PUT /preferences
```

Request：

```Plain Text
{
  "email_enabled":true,
  "push_enabled":false
}
```

---

# 获取订阅列表

```Plain Text
GET /subscriptions
```

---

# 创建订阅

```Plain Text
POST /subscriptions
```

```Plain Text
{
  "object_type":"PORTFOLIO",
  "object_id":10001
}
```

---

# 取消订阅

```Plain Text
DELETE /subscriptions/{id}
```

---

# 事件消费架构

Notification Service 不生产业务事件。

主要消费：

```Plain Text
strategy-events
portfolio-events
backtest-events
ranking-events
community-events
user-events
```

---

# 消费策略事件

监听：

```Plain Text
StrategyPublished
```

通知：

```Plain Text
关注者
```

---

模板：

```Plain Text
您关注的策略《{strategy_name}》已发布新版本
```

---

# 消费组合事件

监听：

```Plain Text
PortfolioPublished
```

通知：

```Plain Text
关注者
```

---

# 消费社区事件

监听：

```Plain Text
CommentCreated
ContentLiked
UserFollowed
```

---

生成：

```Plain Text
COMMENT

LIKE

FOLLOW
```

通知。

---

# 消费回测事件

监听：

```Plain Text
BacktestCompleted
BacktestFailed
```

---

通知作者。

---

# 消费排行榜事件

监听：

```Plain Text
RankingUpdated
TopStrategyChanged
TopPortfolioChanged
```

---

通知：

```Plain Text
策略作者

组合作者
```

---

# 消费会员事件

监听：

```Plain Text
MembershipExpired
MembershipUpgraded
```

---

通知用户。

---

# 通知模板引擎

采用：

```Plain Text
Go Template
```

例如：

```Plain Text
{{ .Nickname }}
您的组合
{{ .PortfolioName }}
已进入收益榜TOP10
```

---

# 渠道架构

支持：

```Plain Text
IN_APP

EMAIL

WEBHOOK
```

---

未来：

```Plain Text
PUSH

SMS

WECHAT
```

---

# In\-App通知

默认开启。

写入：

```Plain Text
notification
```

表。

---

# Email通知

Provider：

```Plain Text
SMTP

Amazon SES

Resend
```

任选其一。

---

# Webhook通知

用于：

```Plain Text
量化机器人

企业集成
```

---

配置：

```Plain Text
{
  "url":"https://xxx.com/webhook"
}
```

---

# Redis缓存设计

未读数量：

```Plain Text
notification:unread:{user_id}
```

TTL：

```Plain Text
10min
```

---

用户偏好：

```Plain Text
notification:pref:{user_id}
```

TTL：

```Plain Text
1h
```

---

订阅关系：

```Plain Text
notification:sub:{object_type}:{object_id}
```

TTL：

```Plain Text
30min
```

---

# Kafka Consumer Group

```Plain Text
notification-consumer-group
```

---

实例：

```Plain Text
notification-worker
```

水平扩展。

---

# 重试机制

失败：

```Plain Text
1s

5s

30s

300s

1800s
```

---

最大：

```Plain Text
5次
```

---

# DLQ

失败进入：

```Plain Text
notification-dlq
```

---

人工处理。

---

# 权限设计

普通用户：

```Plain Text
读取自己的通知
```

---

管理员：

```Plain Text
系统通知管理
```

---

超级管理员：

```Plain Text
模板管理
通知重发
```

---

# 系统通知

管理员发送：

```Plain Text
POST /admin/system-notification
```

---

支持：

```Plain Text
全员

指定会员等级

指定用户
```

---

# 部署架构

```Plain Text
Kafka
  ↓
Notification Consumer
  ↓
Notification Service
  ↓
PostgreSQL

Notification Dispatcher
  ↓
Email

Webhook
```

---

# K8S部署

```Plain Text
notification-api

2 Pods
```

---

```Plain Text
notification-consumer

5 Pods
```

---

```Plain Text
notification-dispatcher

5 Pods
```

---

# 性能指标

通知写入：

```Plain Text
P95 < 50ms
```

---

通知查询：

```Plain Text
P95 < 20ms
```

---

未读数查询：

```Plain Text
P95 < 5ms
```

---

事件消费：

```Plain Text
5000 msg/s
```

---

# 可观测性

Metrics：

```Plain Text
notification_created_total

notification_delivered_total

notification_failed_total

notification_unread_total
```

---

Trace：

```Plain Text
OpenTelemetry
```

---

Logs：

```Plain Text
JSON Structured Log
```

---

# SLA

```Plain Text
Availability
99.95%
```

```Plain Text
Notification Delivery Success
99.9%
```

```Plain Text
P95 Delivery Latency
< 3s
```

---

# MVP范围

必须实现：

✅ In\-App Notification

✅ Comment Notification

✅ Like Notification

✅ Follow Notification

✅ Backtest Notification

✅ Portfolio Notification

✅ Strategy Notification

✅ Preference Management

✅ Kafka Consumer

✅ Template Engine

---

暂缓：

❌ Push

❌ SMS

❌ WeChat

❌ Webhook Marketplace

❌ Workflow Notification

---

# 后续演进

V2 引入 NotificationRule 引擎：

```Plain Text
Event → Rule Engine → Notification
```

```Plain Text
type NotificationRule struct {
    EventType
    TargetType
    Channels
    TemplateID
}
```

新增通知场景（如 PortfolioEnteredTop100、StrategyWinRateAbove80）只需配置规则，无需修改代码。


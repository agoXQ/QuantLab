# QuantLab Notification Service PRD

Version：1\.0

Module：Notification Service

Priority：P1

Status：Draft

Owner：Product Team

Last Update：2026

---

# 模块定位

Notification Service 是 QuantLab 的统一消息与通知中心。

负责：

- 站内通知 

- 消息推送 

- 邮件通知 

- 系统公告 

- 用户订阅 

- 通知偏好管理 

- 运营消息管理 

不负责：

- 用户认证 

- 策略执行 

- 回测计算 

---

# 产品目标

让用户及时获得重要信息。

提升：

- 用户活跃度 

- 用户留存率 

- 策略传播效率 

---

# 核心原则

原则1

重要事件必须通知

---

原则2

用户可配置通知

---

原则3

避免通知骚扰

---

原则4

通知必须可追踪

---

原则5

支持多渠道发送

---

原则6

所有通知来源于事件

---

原则7

通知中心统一管理

---

原则8

用户拥有最终控制权

---

原则9

通知服务必须异步化

---

原则10

通知目标是提升活跃与留存

---

# 系统架构

```Plain Text
Notification Service

├── Notification Center
├── Subscription Center
├── Message Template Center
├── Push Gateway
├── Email Gateway
├── Event Router
├── User Preference Center
└── Operation Center
```

---

# 通知类型

分为：

```Plain Text
System Notification

Business Notification

Social Notification

Operation Notification
```

---

# System Notification

系统通知

示例：

```Plain Text
系统升级

维护公告

服务异常

版本更新
```

---

# Business Notification

业务通知

示例：

```Plain Text
回测完成

回测失败

策略审核通过

策略审核拒绝

策略发布成功
```

---

# Social Notification

社区通知

示例：

```Plain Text
收到点赞

收到评论

收到收藏

新增关注

回复通知
```

---

# Ranking Notification

排行榜通知

---

示例：

```Plain Text
您的策略进入收益榜 Top10

您的策略排名上升 20 位

您的策略跌出榜单
```

---

# Follow Notification

关注通知

---

示例：

```Plain Text
关注的作者发布新策略

关注的作者发布文章

关注的作者获得荣誉
```

---

# AI Notification

AI服务通知

---

示例：

```Plain Text
AI分析完成

AI研究报告生成完成

AI优化建议已生成
```

---

# Market Notification（V2）

市场通知

---

示例：

```Plain Text
策略命中股票池

观察股票突破新高

行业热度上升
```

---

# Notification Center

统一通知中心

---

展示：

```Plain Text
全部通知

未读通知

已读通知

归档通知
```

---

# 消息状态

状态：

```Plain Text
UNREAD

READ

ARCHIVED

DELETED
```

---

# 通知优先级

等级：

```Plain Text
P0 Critical

P1 Important

P2 Normal

P3 Low
```

---

# P0示例

```Plain Text
账户安全

支付异常

系统故障
```

---

# P1示例

```Plain Text
回测完成

排行榜变化

策略审核
```

---

# P2示例

```Plain Text
点赞

收藏

关注
```

---

# 通知渠道

支持：

```Plain Text
站内消息

Email

Web Push

Mobile Push
```

---

未来支持：

```Plain Text
企业微信

钉钉

Webhook

Slack
```

---

# Subscription Center

订阅中心

---

用户可订阅：

```Plain Text
作者

策略

排行榜

市场主题
```

---

# 作者订阅

订阅后获得：

```Plain Text
策略更新

文章更新

荣誉变化
```

---

# 策略订阅

订阅策略后：

```Plain Text
策略更新

回测更新

收益变化
```

---

# 榜单订阅

订阅榜单：

```Plain Text
收益榜

胜率榜

综合榜
```

---

通知：

```Plain Text
进入榜单

排名变化
```

---

# 通知偏好中心

用户可配置：

```Plain Text
是否接收

接收渠道

接收频率
```

---

# 静默模式

支持：

```Plain Text
22:00 ~ 08:00
```

禁止推送。

---

# 消息模板中心

统一模板管理。

---

示例：

```Plain Text
【回测完成】

策略：

{{strategy_name}}

收益率：

{{return_rate}}
```

---

# 模板变量

支持：

```Plain Text
用户昵称

策略名称

收益率

排名

作者名称
```

---

# Event Router

事件总线。

---

接收：

```Plain Text
Strategy Service

Backtest Engine

Ranking Service

Community Service

AI Service
```

产生的事件。

---

# 事件模型

统一格式：

```Plain Text
{
  "event_type":"BACKTEST_FINISHED",
  "user_id":"1001",
  "data":{}
}
```

---

# 异步处理

所有通知异步发送。

---

避免：

阻塞主业务流程。

---

# Push Gateway

负责：

```Plain Text
Web Push

APP Push
```

---

统一发送。

---

# Email Gateway

负责：

```Plain Text
邮件通知
```

---

场景：

```Plain Text
注册验证

密码重置

周报推送
```

---

# 周报系统

自动生成：

```Plain Text
本周收益

本周回测

排行榜变化

推荐策略
```

---

# 月报系统

自动生成：

```Plain Text
月度表现

策略分析

成长情况
```

---

# 运营消息

运营后台支持：

```Plain Text
全站公告

定向推送

活动通知
```

---

# 消息分组

支持：

```Plain Text
系统

策略

社区

排行榜

AI
```

---

# 未读统计

展示：

```Plain Text
总未读

分类未读
```

---

# 消息搜索

支持搜索：

```Plain Text
策略名称

作者名称

通知类型
```

---

# 消息留存

建议：

```Plain Text
普通通知：

180天

系统公告：

永久保留
```

---

# 通知风控

限制：

```Plain Text
单用户

每天最大推送次数
```

---

避免：

通知轰炸。

---

# 通知分析

统计：

```Plain Text
发送量

到达率

打开率

点击率
```

---

# 用户行为分析

统计：

```Plain Text
通知点击后回测率

通知点击后留存率

通知点击后策略发布率
```

---

# MVP范围

必须实现：

✅ 站内通知

✅ 回测完成通知

✅ 排行榜通知

✅ 评论通知

✅ 点赞通知

✅ 关注通知

✅ 通知中心

✅ 通知偏好

---

暂缓：

❌ App Push

❌ Webhook

❌ 企业微信

❌ Slack

❌ 市场预警

---

# KPI

通知打开率

通知点击率

通知召回率

用户活跃率

用户留存率

---

# 核心事件清单

建议平台统一定义 Event Catalog：

```Plain Text
StrategyCreated
StrategyPublished
StrategyUpdated

BacktestStarted
BacktestFinished
BacktestFailed

RankingEntered
RankingUpdated
RankingDropped

CommentCreated
ReplyCreated
LikeCreated
FavoriteCreated

FollowCreated

AIReportGenerated
AIOptimizationCompleted
```

---


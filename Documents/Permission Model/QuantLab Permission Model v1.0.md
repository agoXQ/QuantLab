# QuantLab Permission Model v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Platform Team

Scope:

```Plain Text
User Service
Strategy Service
Portfolio Service
Backtest Engine
Ranking Service
Community Service
Notification Service
AI Service
Admin Console
```

---

# 文档目标

统一解决以下问题：

```Plain Text
谁可以访问什么功能？

谁可以创建什么资源？

谁可以查看什么数据？

会员等级如何控制？

AI额度如何控制？

回测额度如何控制？

管理员权限如何控制？
```

避免出现：

```Plain Text
User Service有自己的权限体系

AI Service有自己的额度体系

Portfolio有自己的会员体系

最终形成权限碎片化
```

---

# 权限模型设计原则

QuantLab采用：

```Plain Text
RBAC
+
Subscription
+
Quota
+
Feature Flag
```

四层模型。

---

# 整体架构

```Plain Text
User
                  │
                  ▼
         Membership Layer
                  │
                  ▼
          Role Layer
                  │
                  ▼
        Feature Layer
                  │
                  ▼
          Quota Layer
                  │
                  ▼
            Resource
```

---

# 四层权限模型

## Layer 1：Membership

决定：

```Plain Text
商业权限
```

例如：

```Plain Text
Free

Pro

Master

Enterprise
```

---

## Layer 2：Role

决定：

```Plain Text
系统权限
```

例如：

```Plain Text
USER

MODERATOR

ADMIN

SUPER_ADMIN
```

---

## Layer 3：Feature

决定：

```Plain Text
是否拥有功能
```

例如：

```Plain Text
AI Strategy Generation

Portfolio Publish

Private Portfolio
```

---

## Layer 4：Quota

决定：

```Plain Text
能用多少
```

例如：

```Plain Text
每日AI次数

最大策略数量

最大回测次数
```

---

# Membership设计

统一会员等级：

```Plain Text
FREE

PRO

MASTER

ENTERPRISE
```

---

# FREE

适用于：

```Plain Text
普通用户
```

权限：

```Plain Text
浏览策略
浏览排行榜

创建策略
创建组合

基础回测
```

限制：

```Plain Text
AI受限

资源受限
```

---

# PRO

适用于：

```Plain Text
付费个人用户
```

开放：

```Plain Text
全部量化能力
```

---

# MASTER

适用于：

```Plain Text
职业投资者

KOL

研究员
```

开放：

```Plain Text
高级AI

高级回测

私有组合
```

---

# ENTERPRISE

适用于：

```Plain Text
机构用户
```

开放：

```Plain Text
团队空间

API Access

Webhook

专属资源
```

---

# Membership数据模型

```Plain Text
type Membership struct {
    UserID

    Tier

    StartedAt

    ExpiredAt
}
```

---

# Role模型

统一角色：

```Plain Text
USER

MODERATOR

ADMIN

SUPER_ADMIN
```

---

# USER

权限：

```Plain Text
平台普通使用
```

---

# MODERATOR

权限：

```Plain Text
社区审核

举报处理

内容封禁
```

---

# ADMIN

权限：

```Plain Text
用户管理

排行榜管理

系统通知
```

---

# SUPER\_ADMIN

权限：

```Plain Text
全平台权限
```

---

# RBAC数据模型

```Plain Text
type Role struct {
    ID
    Name
}
```

---

```Plain Text
type UserRole struct {
    UserID
    RoleID
}
```

---

# Permission模型

统一命名：

```Plain Text
resource.action
```

例如：

```Plain Text
strategy.create

strategy.update

strategy.publish

portfolio.create

portfolio.publish

backtest.run

ranking.view

community.post

ai.generate
```

---

# 核心权限清单

## Strategy

```Plain Text
strategy.create

strategy.update

strategy.delete

strategy.publish

strategy.fork

strategy.private
```

---

## Portfolio

```Plain Text
portfolio.create

portfolio.update

portfolio.delete

portfolio.publish

portfolio.private
```

---

## Backtest

```Plain Text
backtest.run

backtest.export
```

---

## Community

```Plain Text
community.post

community.comment

community.delete
```

---

## Ranking

```Plain Text
ranking.view

ranking.manage
```

---

## AI

```Plain Text
ai.chat

ai.generate.strategy

ai.generate.portfolio

ai.optimize.strategy

ai.optimize.portfolio
```

---

## Admin

```Plain Text
admin.user.manage

admin.content.manage

admin.notification.manage
```

---

# Feature模型

Feature是：

```Plain Text
产品功能开关
```

---

统一命名：

```Plain Text
FEATURE_*
```

---

# 功能列表

```Plain Text
FEATURE_AI_CHAT

FEATURE_AI_STRATEGY_GENERATION

FEATURE_AI_PORTFOLIO_GENERATION

FEATURE_PRIVATE_STRATEGY

FEATURE_PRIVATE_PORTFOLIO

FEATURE_ADVANCED_BACKTEST

FEATURE_EXPORT_BACKTEST

FEATURE_WEBHOOK

FEATURE_API_ACCESS

FEATURE_TEAM_WORKSPACE
```

---

# Membership与Feature映射

---

# Quota模型

Quota用于：

```Plain Text
次数限制

数量限制

容量限制
```

---

# Quota分类

```Plain Text
DAILY

MONTHLY

TOTAL
```

---

# Free额度

## AI

```Plain Text
20次/天
```

---

## 策略

```Plain Text
最多50个
```

---

## 组合

```Plain Text
最多20个
```

---

## 回测

```Plain Text
20次/天
```

---

## 社区发帖

```Plain Text
20篇/月
```

---

# Pro额度

## AI

```Plain Text
500次/天
```

---

## 策略

```Plain Text
500个
```

---

## 组合

```Plain Text
200个
```

---

## 回测

```Plain Text
500次/天
```

---

# Master额度

```Plain Text
无限
```

但保留风控上限。

---

# Enterprise额度

采用：

```Plain Text
Contract Based
```

合同配置。

---

# Quota数据模型

```Plain Text
type UserQuota struct {
    UserID

    Resource

    Used

    Limit

    ResetAt
}
```

---

# Resource枚举

```Plain Text
AI_CHAT

AI_STRATEGY

AI_PORTFOLIO

STRATEGY

PORTFOLIO

BACKTEST

COMMUNITY_POST
```

---

# 权限判定流程

```Plain Text
User
 ↓
Authentication
 ↓
Role Check
 ↓
Membership Check
 ↓
Feature Check
 ↓
Quota Check
 ↓
Execute
```

---

# 示例

用户执行：

```Plain Text
POST /ai/strategies/generate
```

系统检查：

```Plain Text
已登录？

有ai.generate.strategy权限？

会员允许？

Feature开启？

Quota未超限？
```

全部通过：

```Plain Text
执行
```

---

# 资源可见性模型

统一：

```Plain Text
PUBLIC

UNLISTED

PRIVATE
```

---

# Strategy Visibility

```Plain Text
type StrategyVisibility string
```

---

```Plain Text
PUBLIC

PRIVATE
```

---

# Portfolio Visibility

```Plain Text
PUBLIC

PRIVATE
```

---

未来：

```Plain Text
SUBSCRIBERS_ONLY
```

---

# Community Visibility

```Plain Text
PUBLIC

FOLLOWERS_ONLY
```

---

# API权限矩阵

---

# 社区权限矩阵

---

# 管理后台权限

模块：

```Plain Text
User Management

Content Management

Ranking Management

Notification Management

System Config
```

---

# 风控权限

触发：

```Plain Text
异常回测

恶意刷榜

API滥用

AI滥用
```

---

状态：

```Plain Text
NORMAL

LIMITED

BANNED
```

---

# UserStatus模型

```Plain Text
type UserStatus string
```

---

```Plain Text
ACTIVE

SUSPENDED

BANNED
```

---

# 权限缓存

Redis：

```Plain Text
perm:user:{user_id}
```

TTL：

```Plain Text
30min
```

---

会员：

```Plain Text
membership:{user_id}
```

TTL：

```Plain Text
30min
```

---

Quota：

```Plain Text
quota:{user_id}
```

TTL：

```Plain Text
5min
```

---

# 权限事件

Topic：

```Plain Text
user-events
```

---

事件：

```Plain Text
MembershipUpgraded

MembershipExpired

UserRoleChanged

QuotaExceeded

UserBanned
```

---

# 审计日志

记录：

```Plain Text
Who

When

Action

Resource

Result
```

---

例如：

```Plain Text
ADMIN

Delete Strategy

Strategy#1001

SUCCESS
```

---

# SaaS计费模型（建议）

实际上 QuantLab 不应该只有：

```Plain Text
Free
Pro
Master
```

建议未来采用：

## Free

```Plain Text
社区用户
```

---

## Quant Pro

```Plain Text
个人量化投资者
```

---

## Quant Master

```Plain Text
策略作者
KOL
研究员
```

---

## Quant Studio

```Plain Text
团队版
```

---

## Enterprise

```Plain Text
机构版
```

---

# 当前设计的一个重要补充（建议纳入 v1\.1）

到目前为止，我们设计了：

```Plain Text
Strategy
Portfolio
Backtest
Ranking
Community
AI
```

但是缺少一个核心商业化对象：

```Plain Text
Subscription（订阅）
```

例如：

```Plain Text
订阅策略作者

订阅组合

订阅高级策略

订阅研究报告
```

因此建议新增领域：

```Plain Text
Subscription Domain
```

后续可能演进为：

```Plain Text
Creator Economy
```

即：

```Plain Text
策略作者
     ↓
发布策略
     ↓
用户订阅
     ↓
收益分成
```

这会直接影响：

```Plain Text
Portfolio Service
Community Service
Ranking Service
Notification Service
```

甚至可能成为 QuantLab 后期最大的收入来源之一。

---

# Permission Model 在整体架构中的位置

至此平台基础架构层已经形成：

```Plain Text
Platform PRD
      ↓

Domain Model
      ↓

Event Specification
      ↓

API Specification
      ↓

Permission Model
      ↓

10个核心服务TD
```




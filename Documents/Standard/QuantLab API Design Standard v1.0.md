# QuantLab API Design Standard v1\.0

**Document Version:** v1\.0
**Project:** QuantLab
**Document Type:** API Design Standard
**Status:** Approved
**Last Updated:** 2026\-06

---

# Introduction

## 1\.1 Purpose

本文档定义 QuantLab 平台统一 API 设计规范。

目标：

- 保持 API 风格一致

- 降低前后端协作成本

- 保证接口长期兼容性

- 支撑微服务演进

- 支撑 OpenAPI 自动化治理

- 支撑 SDK 自动生成

---

## 1\.2 Scope

适用于：

```Plain Text
REST API
Internal API
Public API
Partner API
Admin API
Mobile API
```

---

## 1\.3 Design Principles

所有 API 设计必须遵循以下原则。

### Principle 1：API First

开发流程：

```Plain Text
API Design
    ↓
Review
    ↓
OpenAPI Contract
    ↓
Implementation
```

禁止：

```Plain Text
先写代码再补接口文档
```

---

### Principle 2：Contract First

API Contract 是系统间唯一协议。

禁止：

```Plain Text
通过代码推断接口
```

---

### Principle 3：Backward Compatible

接口必须向后兼容。

允许：

```Plain Text
新增字段
新增接口
```

禁止：

```Plain Text
删除字段
修改字段类型
修改字段语义
```

---

### Principle 4：Consistency

同类资源必须保持一致风格。

---

# API Classification

## 2\.1 Public API

面向用户。

例如：

```Plain Text
/api/v1/strategies
/api/v1/factors
```

---

## 2\.2 Internal API

微服务之间调用。

例如：

```Plain Text
/internal/v1/users
/internal/v1/billing
```

---

## 2\.3 Admin API

管理后台。

例如：

```Plain Text
/admin/v1/users
/admin/v1/subscriptions
```

---

## 2\.4 Webhook API

事件推送。

例如：

```Plain Text
/webhooks/stripe
/webhooks/github
```

---

# URL Design Standard

## 3\.1 Resource\-Oriented Design

统一采用 RESTful 风格。

正确：

```HTTP
GET /api/v1/strategies

POST /api/v1/strategies

GET /api/v1/strategies/{id}

PATCH /api/v1/strategies/{id}

DELETE /api/v1/strategies/{id}
```

---

错误：

```HTTP
POST /createStrategy

POST /queryStrategy

GET /getStrategy
```

---

## 3\.2 Naming Convention

统一：

```Plain Text
snake_case
```

例如：

```Plain Text
/api/v1/backtest_jobs
/api/v1/factor_groups
```

---

禁止：

```Plain Text
/api/v1/backtestJobs
/api/v1/FactorGroups
```

---

## 3\.3 URL Hierarchy

允许：

```Plain Text
/api/v1/strategies/{id}/backtests
```

最多三级。

禁止：

```Plain Text
/api/v1/a/b/c/d/e/f
```

---

# HTTP Method Standard

## 4\.1 GET

查询资源。

必须：

```Plain Text
幂等
无副作用
```

---

## 4\.2 POST

创建资源。

例如：

```HTTP
POST /api/v1/strategies
```

---

## 4\.3 PUT

完整更新。

例如：

```HTTP
PUT /api/v1/users/{id}
```

---

## 4\.4 PATCH

部分更新。

推荐：

```HTTP
PATCH /api/v1/users/{id}
```

---

## 4\.5 DELETE

删除资源。

例如：

```HTTP
DELETE /api/v1/strategies/{id}
```

---

# Request Design Standard

## 5\.1 JSON Only

统一：

```HTTP
Content-Type: application/json
```

---

## 5\.2 Request Body

正确：

```JSON
{
  "name": "Momentum Strategy",
  "description": "..."
}
```

---

禁止：

```JSON
{
  "strategyName": "...",
  "strategyDesc": "..."
}
```

---

统一使用：

```Plain Text
snake_case
```

---

## 5\.3 Validation

所有输入必须验证。

包括：

```Plain Text
Required

Length

Range

Enum

Format

Business Rule
```

---

# Response Design Standard

## 6\.1 Unified Response Structure

统一返回格式：

```JSON
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## 6\.2 Success Response

```JSON
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "stg_001"
  }
}
```

---

## 6\.3 Error Response

```JSON
{
  "code": 10001,
  "message": "strategy_not_found",
  "request_id": "req_xxx"
}
```

---

## 6\.4 Metadata

分页接口统一：

```JSON
{
  "code": 0,
  "message": "success",
  "data": [],
  "meta": {}
}
```

---

# Resource Identifier Standard

## 7\.1 Public ID

禁止暴露数据库自增 ID。

错误：

```JSON
{
  "id": 123
}
```

---

正确：

```JSON
{
  "id": "stg_01HJY..."
}
```

---

## 7\.2 ID Prefix

统一前缀。

例如：

```Plain Text
usr_
sub_
inv_
stg_
fac_
bt_
```

---

示例：

```Plain Text
usr_01HJ...
inv_01HJ...
```

---

# Pagination Standard

## 8\.1 Cursor Pagination

推荐：

```HTTP
GET /strategies?cursor=xxx&limit=20
```

---

返回：

```JSON
{
  "data": [],
  "meta": {
    "next_cursor": "...",
    "has_more": true
  }
}
```

---

## 8\.2 Offset Pagination

仅用于后台管理系统。

```HTTP
?page=1&page_size=20
```

---

# Filtering Standard

## 9\.1 Query Parameter

例如：

```HTTP
GET /strategies?status=active
```

---

## 9\.2 Multi Filter

```HTTP
GET /strategies?
status=active&
category=momentum
```

---

## 9\.3 Sort

```HTTP
GET /strategies?sort=-created_at
```

含义：

```Plain Text
DESC
```

---

```HTTP
GET /strategies?sort=created_at
```

含义：

```Plain Text
ASC
```

---

# Error Code Standard

## 10\.1 Error Structure

统一：

```JSON
{
  "code": 20001,
  "message": "subscription_expired",
  "request_id": "req_xxx"
}
```

---

## 10\.2 Error Category

### 1xxxx

Validation

---

### 2xxxx

Business Error

---

### 3xxxx

Permission Error

---

### 4xxxx

Resource Error

---

### 5xxxx

System Error

---

## 10\.3 Example

```Plain Text
10001 Invalid Parameter

20001 Subscription Expired

30001 Permission Denied

40001 Strategy Not Found

50001 Internal Error
```

---

# Authentication Standard

## 11\.1 Authentication

统一：

```Plain Text
OAuth2
JWT
```

---

Header：

```HTTP
Authorization: Bearer xxx
```

---

## 11\.2 User Context

服务内部统一：

```JSON
{
  "user_id": "usr_xxx",
  "tenant_id": "tenant_xxx"
}
```

---

# Authorization Standard

## 12\.1 RBAC

统一：

```Plain Text
Role

Permission

Resource
```

---

## 12\.2 Permission Naming

统一：

```Plain Text
strategy.read

strategy.write

billing.read

billing.manage
```

---

# Idempotency Standard

## 13\.1 Required APIs

必须支持幂等：

```Plain Text
Payment

Subscription

Billing

Order
```

---

## 13\.2 Header

```HTTP
Idempotency-Key: uuid
```

---

## 13\.3 Response

同一请求：

返回相同结果。

---

# Webhook Standard

## 14\.1 Event Structure

统一：

```JSON
{
  "event_id": "",
  "event_name": "",
  "event_version": "1.0",
  "occurred_at": "",
  "payload": {}
}
```

---

## 14\.2 Retry Policy

```Plain Text
1 min

5 min

15 min

1 hour

6 hour
```

---

## 14\.3 Signature Verification

必须支持：

```Plain Text
HMAC SHA256
```

---

# Async API Standard

## 15\.1 Long Running Job

适用于：

```Plain Text
Backtest

Factor Calculation

Screening

Report Export
```

---

## 15\.2 Response

```JSON
{
  "job_id": "job_xxx",
  "status": "pending"
}
```

---

## 15\.3 Query Job

```HTTP
GET /jobs/{job_id}
```

---

# API Versioning Standard

## 16\.1 Version Strategy

统一：

```Plain Text
/api/v1
/api/v2
```

---

## 16\.2 Deprecation

接口废弃必须提前：

```Plain Text
90 Days
```

通知。

---

## 16\.3 Sunset Header

推荐：

```HTTP
Sunset: 2026-12-31
```

---

# OpenAPI Standard

## 17\.1 Mandatory OpenAPI

所有接口必须生成：

```Plain Text
OpenAPI 3.x
```

---

## 17\.2 Source Of Truth

统一：

```Plain Text
OpenAPI Contract
```

---

## 17\.3 SDK Generation

自动生成：

```Plain Text
Python SDK

Go SDK

TypeScript SDK
```

---

# Observability Standard

## 18\.1 Required Headers

所有请求必须包含：

```HTTP
X-Request-Id
```

---

推荐：

```HTTP
X-Trace-Id
```

---

## 18\.2 Logging

必须记录：

```Plain Text
Request

Response

Latency

Status Code
```

---

# Rate Limiting Standard

## 19\.1 Public API

默认：

```Plain Text
100 requests/minute
```

---

## 19\.2 Premium Plan

根据订阅等级扩展。

---

## 19\.3 Headers

```HTTP
X-RateLimit-Limit

X-RateLimit-Remaining

X-RateLimit-Reset
```

---

# QuantLab Domain API Guidelines

## 20\.1 Strategy APIs

资源：

```Plain Text
Strategy
StrategyVersion
StrategyTag
```

---

## 20\.2 Formula APIs

资源：

```Plain Text
Formula
AST
FormulaTest
```

---

## 20\.3 Backtest APIs

资源：

```Plain Text
BacktestJob
BacktestReport
Order
Position
```

---

## 20\.4 Portfolio APIs

资源：

```Plain Text
Portfolio
PortfolioItem
RebalanceRule
```

---

## 20\.5 Ranking APIs

资源：

```Plain Text
Ranking
RankingSnapshot
Score
```

---

## 20\.6 Community APIs

资源：

```Plain Text
Content
Comment
Like
Favorite
```

---

## 20\.7 Billing APIs

资源：

```Plain Text
Membership
Subscription
Order
Payment
Invoice
Coupon
```

---

## 20\.8 User APIs

资源：

```Plain Text
User
Profile
Role
Permission
```

---

## 20\.9 Market Data APIs

资源：

```Plain Text
Security
KLine
Financial
Factor
Index
```

---

## 20\.10 AI APIs

资源：

```Plain Text
AITask
Conversation
AIReport
```

---

## 20\.11 Notification APIs

资源：

```Plain Text
Notification
Subscription
Template
```

---

# API Review Checklist

上线前必须检查：

✓ Resource Naming

✓ OpenAPI Completed

✓ Validation Defined

✓ Error Codes Defined

✓ Authentication Enabled

✓ Authorization Enabled

✓ Logging Enabled

✓ Rate Limiting Enabled

✓ Monitoring Enabled

✓ Backward Compatibility Verified

---

# Final Principle

API 是产品，而不是代码实现细节。

任何 API 设计必须满足：

Consistency \> Simplicity \> Extensibility \> Convenience

未经评审不得破坏既有 Contract。


# QuantLab API Specification v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Architecture Team

Applicable Services:

```Plain Text
user-service
strategy-service
formula-engine
portfolio-service
market-data-service
backtest-engine
ranking-service
community-service
notification-service
ai-service
```

---

# API设计原则

## Principle 1

资源导向（Resource Oriented）

正确：

```Plain Text
GET /strategies/{id}
```

错误：

```Plain Text
POST /queryStrategy
```

---

## Principle 2

名词优先

正确：

```Plain Text
/strategies
/portfolios
/backtests
```

错误：

```Plain Text
/getStrategy
/createPortfolio
```

---

## Principle 3

统一复数

统一：

```Plain Text
/users
/strategies
/portfolios
/backtests
```

禁止：

```Plain Text
/user
/strategy
```

---

## Principle 4

无状态

所有请求：

```Plain Text
Stateless
```

---

## Principle 5

API First

所有功能：

```Plain Text
先定义API
再开发
```

---

# API版本规范

统一：

```Plain Text
/api/v1
```

例如：

```Plain Text
/api/v1/strategies
/api/v1/portfolios
```

---

未来：

```Plain Text
/api/v2
```

---

禁止：

```Plain Text
/api/strategy/v1
```

---

# URL规范

统一：

```Plain Text
/api/v1/{resource}
```

---

示例：

```Plain Text
/api/v1/users

/api/v1/strategies

/api/v1/portfolios

/api/v1/backtests
```

---

# HTTP Method规范

查询：

```Plain Text
GET
```

---

创建：

```Plain Text
POST
```

---

更新：

```Plain Text
PUT
```

---

局部更新：

```Plain Text
PATCH
```

---

删除：

```Plain Text
DELETE
```

---

# 统一响应结构

所有API：

```Plain Text
{
  "request_id":"req_xxx",

  "code":"SUCCESS",

  "message":"success",

  "data": {}
}
```

---

# 成功响应

```Plain Text
{
  "request_id":"req_123",

  "code":"SUCCESS",

  "message":"success",

  "data":{
    "id":1001
  }
}
```

---

# 失败响应

```Plain Text
{
  "request_id":"req_123",

  "code":"STRATEGY_NOT_FOUND",

  "message":"strategy not found",

  "data":null
}
```

---

# Formula Screen API

## POST /api/v1/formulas/screen

服务端公式选股接口。

请求：

```JSON
{
  "formula": "ROE > 15 AND PE < 20",
  "as_of_date": "2026-06-27",
  "universe_filter": {
    "market": "CN",
    "exchange": "SSE",
    "industry": "银行",
    "asset_type": "STOCK",
    "status": "LISTED",
    "stock_codes": []
  },
  "limit": 500
}
```

字段说明：

- `formula` 必填，DSL 选股公式。
- `as_of_date` 可选，默认服务端当前日期。
- `universe_filter.exchange` 可选，不传则不过滤交易所。
- `universe_filter.industry` 可选，不传则不过滤行业。
- `universe_filter.asset_type` 可选，不传默认为 `STOCK`。
- `universe_filter.status` 可选，不传默认为 `LISTED`。
- `universe_filter.stock_codes` 可选，传入时作为手工股票池。
- `limit` 可选，默认 500，表示最多返回结果条数，服务端限制最大值；它不截断参与计算的候选股票池。
- `data_version` 不由前端传入，后端默认使用最新 Market Data 版本。

错误响应会返回可直接展示给前端用户的 `error` 文本，例如股票池为空、最新数据版本不存在、Market Data 查询失败或公式执行失败。

响应：

```JSON
{
  "formula_hash": "sha256...",
  "plan_type": "FILTER",
  "data_version": "2026.06.27",
  "universe_size": 500,
  "items": [
    {
      "stock_code": "600519",
      "stock_name": "贵州茅台",
      "exchange": "SSE",
      "industry": "白酒",
      "score": null,
      "selected": true
    }
  ]
}
```

---

# Request ID规范

Header：

```Plain Text
X-Request-ID
```

---

不存在：

自动生成。

---

全链路透传。

---

# Trace ID规范

Header：

```Plain Text
X-Trace-ID
```

---

OpenTelemetry兼容。

---

# 分页规范

统一：

```Plain Text
?page=1&page_size=20
```

---

返回：

```Plain Text
{
  "items":[],
  "pagination":{
      "page":1,
      "page_size":20,
      "total":100,
      "total_pages":5
  }
}
```

---

# 排序规范

统一：

```Plain Text
?sort_by=created_at
&order=desc
```

---

支持：

```Plain Text
asc
desc
```

---

# 过滤规范

统一：

```Plain Text
?status=PUBLISHED
```

---

多个：

```Plain Text
?status=PUBLISHED,DRAFT
```

---

# 搜索规范

统一：

```Plain Text
?q=roe strategy
```

---

禁止：

```Plain Text
?keyword=
```

统一使用：

```Plain Text
q
```

---

# 时间格式

统一：

ISO8601

```Plain Text
2026-06-13T15:00:00Z
```

---

禁止：

```Plain Text
2026/06/13
```

---

# 金额规范

统一：

```Plain Text
{
  "cost":"12.35"
}
```

字符串传输。

---

禁止：

```Plain Text
{
  "cost":12.35
}
```

避免精度问题。

---

# 百分比规范

统一：

```Plain Text
{
  "annual_return":0.185
}
```

表示：

```Plain Text
18.5%
```

---

禁止：

```Plain Text
{
  "annual_return":"18.5%"
}
```

---

# ID规范

统一：

```Plain Text
Snowflake
```

类型：

```Plain Text
{
  "id":"190000000001"
}
```

---

统一字符串返回。

避免JS精度问题。

---

# 认证规范

Header：

```Plain Text
Authorization: Bearer <token>
```

---

JWT：

```Plain Text
RS256
```

---

# 用户上下文

Header：

```Plain Text
X-User-ID
```

---

仅网关注入。

禁止客户端传递。

---

# 幂等规范

创建类接口：

支持：

```Plain Text
Idempotency-Key
```

---

例如：

```Plain Text
POST /backtests
```

---

重复请求：

返回第一次结果。

---

# 错误码规范

格式：

```Plain Text
DOMAIN_ERROR
```

---

示例：

```Plain Text
USER_NOT_FOUND

USER_BANNED

STRATEGY_NOT_FOUND

PORTFOLIO_NOT_FOUND

BACKTEST_FAILED

AI_QUOTA_EXCEEDED
```

---

# HTTP状态码规范

成功：

```Plain Text
200
```

---

创建：

```Plain Text
201
```

---

参数错误：

```Plain Text
400
```

---

未认证：

```Plain Text
401
```

---

权限不足：

```Plain Text
403
```

---

不存在：

```Plain Text
404
```

---

冲突：

```Plain Text
409
```

---

限流：

```Plain Text
429
```

---

系统错误：

```Plain Text
500
```

---

# API权限模型

统一：

```Plain Text
PUBLIC

AUTHENTICATED

PREMIUM

ADMIN
```

---

示例：

```Plain Text
GET /rankings

PUBLIC
```

---

```Plain Text
POST /backtests

AUTHENTICATED
```

---

```Plain Text
POST /ai/strategies/generate

PREMIUM
```

---

# 限流规范

Header：

```Plain Text
X-RateLimit-Limit

X-RateLimit-Remaining

X-RateLimit-Reset
```

---

# OpenAPI规范

所有服务：

必须生成：

```Plain Text
OpenAPI 3.1
```

---

路径：

```Plain Text
/openapi.json
```

---

# SDK规范

统一生成：

```Plain Text
TypeScript SDK

Python SDK

Go SDK
```

---

# Webhook规范

签名：

```Plain Text
X-Signature
```

---

算法：

```Plain Text
HMAC-SHA256
```

---

# 文件上传规范

统一：

```Plain Text
POST /files
```

---

返回：

```Plain Text
{
  "file_id":"123",
  "url":"..."
}
```

---

# API审计日志

记录：

```Plain Text
UserID

RequestID

Path

Method

Duration

Status
```

---

# API监控指标

统一：

```Plain Text
http_requests_total

http_request_duration

http_errors_total

http_4xx_total

http_5xx_total
```

---

# API SLA

内部API：

```Plain Text
P95 < 100ms
```

---

查询API：

```Plain Text
P95 < 50ms
```

---

AI API：

```Plain Text
P95 < 15s
```

---

可用性：

```Plain Text
99.95%
```

---

# API网关规范

统一入口：

```Plain Text
api.quantlab.com
```

---

路由：

```Plain Text
/users

/strategies

/portfolios

/backtests

/rankings

/community

/notifications

/ai
```

---

# GraphQL（未来）

V1：

```Plain Text
REST Only
```

---

V2：

增加：

```Plain Text
GraphQL Gateway
```

---

# AI Agent API（未来）

预留：

```Plain Text
POST /agent/tasks
```

---

支持：

```Plain Text
Create Strategy

Create Portfolio

Run Backtest

Analyze Result
```

---

# API成熟度模型

Level 1：

```Plain Text
REST
```

---

Level 2：

```Plain Text
REST + OpenAPI
```

---

Level 3：

```Plain Text
REST + Event + SDK
```

---

Level 4：

```Plain Text
AI Agent Native Platform
```

---

# QuantLab API统一资源目录

核心资源：

```Plain Text
/users

/strategies

/strategy-versions

/portfolios

/portfolio-versions

/backtests

/rankings

/contents

/comments

/notifications

/ai/tasks

/ai/reports
```

---

# 与前面文档的关系

本规范约束：

```Plain Text
User Service TD
Strategy Service TD
Portfolio Service TD
Backtest Engine TD
Ranking Service TD
Community Service TD
Notification Service TD
AI Service TD
```

所有未来新增接口必须遵循本规范。

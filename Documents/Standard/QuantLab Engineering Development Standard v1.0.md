# QuantLab Engineering Development Standard v1\.0

**Document Version:** v1\.0
**Project:** QuantLab
**Document Type:** Engineering Standard
**Status:** Approved
**Last Updated:** 2026\-06

---

# Introduction

## 1\.1 Purpose

本文档定义 QuantLab 项目统一工程开发规范。

目标：

- 提高代码可维护性

- 保证系统架构一致性

- 降低协作成本

- 提高系统稳定性

- 规范 AI Assisted Coding 输出

- 支撑长期演进

---

## 1\.2 Scope

适用于：

- Backend Service (Go)
- Quant Engine (Go)
- Formula Engine (Go)
- Market Data Service (Go)
- Billing Service (Go)
- User Service (Go)
- Strategy Service (Go)
- Portfolio Service (Go)
- AI Agent (Go)
- Infrastructure Components

---

## 1\.3 Engineering Principles

所有技术决策必须遵守：

### Principle 1：Domain First

业务模型优先于技术实现。

### Principle 2：Maintainability First

可维护性优先于短期开发速度。

### Principle 3：Explicit Over Implicit

显式优于隐式。

### Principle 4：Consistency Over Preference

统一优于个人习惯。

### Principle 5：Automation First

所有重复工作优先自动化。

---

# Project Architecture Standard

## 2\.1 Architecture Pattern

统一采用：

Go \+ DDD \+ Clean Architecture \+ Event Driven Architecture

---

## 2\.2 Layer Definition

每个业务域必须包含：

```Plain Text
Domain
Application
Infrastructure
Interfaces
```

---

### Domain Layer

职责：

- Entity

- Value Object

- Aggregate

- Domain Service

- Domain Event

- Repository Interface

禁止：

- SQL

- HTTP

- MQ SDK

- Cache SDK

---

### Application Layer

职责：

- Use Case

- Command

- Query

- Transaction Orchestration

禁止：

- 复杂业务规则

---

### Infrastructure Layer

职责：

- Database

- Redis

- MQ

- Third Party API

---

### Interface Layer

职责：

- REST API

- gRPC

- Message Consumer

---

# Repository Structure Standard

## 3\.1 Monorepo Structure

```Plain Text
quantlab/

├── services/
│
├── engines/
│
├── sdk/
│
├── infrastructure/
│
├── docs/
│
├── scripts/
│
├── deployment/
│
└── tools/
```

---

## 3\.2 Service Structure

```Plain Text
billing/

├── domain/
├── application/
├── infrastructure/
├── interfaces/
├── tests/
└── docs/
```

---

# Python Development Standard（辅助脚本）

## 4\.1 Version

Python \>= 3\.12

适用范围：数据采集脚本、运维工具、AI 服务原型。核心业务服务使用 Go。

---

## 4\.2 Style Guide

遵循：

PEP8

Black

isort

ruff

mypy

---

## 4\.3 Naming Convention

### Class

```Python
BillingService
FactorCalculator
StrategyExecutor
```

PascalCase

---

### Function

```Python
create_strategy()
run_backtest()
calculate_factor()
```

snake\_case

---

### Constant

```Python
MAX_RETRY_COUNT
DEFAULT_TIMEOUT
```

UPPER\_CASE

---

### Boolean

必须：

```Python
is_active
is_deleted
has_permission
can_trade
```

禁止：

```Python
flag
status
```

---

## 4\.4 Type Hint

强制要求。

```Python
def create_user(
    request: CreateUserRequest
) -> User:
```

禁止：

```Python
def create_user(req):
```

---

## 4\.5 Dependency Injection

禁止：

```Python
db = Database()
```

必须：

```Python
class UserService:
    def __init__(
        self,
        repository: UserRepository
    ):
        ...
```

---

# Go Development Standard（主力语言）

## 5\.1 Version

Go \>= 1\.24

适用范围：所有核心业务服务与计算引擎。

---

## 5\.2 Style Guide

统一：

gofmt

golangci\-lint

go vet

---

## 5\.3 Package Structure

```Plain Text
internal/

├── domain
├── application
├── infrastructure
└── interfaces
```

---

## 5\.4 Interface Design

优先小接口原则。

正确：

```Go
type UserReader interface {
    GetByID(ctx context.Context, id int64)
}
```

禁止：

```Go
type UserManager interface {
    Create()
    Update()
    Delete()
    Query()
    ...
}
```

---

## 5\.5 Context

所有外部调用必须传递 Context。

```Go
func CreateUser(
    ctx context.Context,
)
```

---

## 5\.6 Error Handling

必须：

```Go
if err != nil {
    return err
}
```

禁止：

```Go
panic(err)
```

业务代码中不得使用 panic。

---

# C\+\+ Development Standard

## 6\.1 Version

C\+\+20

---

## 6\.2 Style Guide

统一采用：

Google C\+\+ Style Guide

clang\-format

clang\-tidy

---

## 6\.3 Namespace

必须：

```C++
namespace quantlab::backtest
{
}
```

禁止：

```C++
using namespace std;
```

---

## 6\.4 Smart Pointer

必须：

```C++
std::unique_ptr
std::shared_ptr
```

禁止：

```C++
new
delete
```

直接暴露。

---

## 6\.5 RAII

所有资源管理必须遵循 RAII。

包括：

- File

- Socket

- Memory

- Thread

---

## 6\.6 Exception

核心计算路径：

禁止异常控制流程。

使用：

```C++
std::expected
```

进行错误返回。

---

# API Design Standard（Go Gin 框架）

## 7\.1 URL Convention

正确：

```Plain Text
GET /api/v1/strategies

POST /api/v1/strategies

GET /api/v1/strategies/{id}
```

---

## 7\.2 Response Structure

统一：

```JSON
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## 7\.3 API Versioning

统一：

```Plain Text
/api/v1
/api/v2
```

禁止：

```Plain Text
/api/new
```

---

# Database Design Standard

## 8\.1 Naming Convention

统一：

snake\_case

---

## 8\.2 Table Naming

```SQL
strategy
formula
backtest_job
billing_invoice
portfolio
```

---

## 8\.3 Audit Fields

所有业务表必须包含：

```SQL
created_at
updated_at
created_by
updated_by
```

---

## 8\.4 Soft Delete

统一：

```SQL
deleted_at
```

禁止：

```SQL
is_deleted
```

---

# Event Driven Standard

## 9\.1 Event Naming

统一：

```Plain Text
<Entity><Action>
```

与 Event Specification 保持一致。

例如：

```Plain Text
SubscriptionCreated
OrderPaid

StrategyCreated

UserCreated
```

---

## 9\.2 Event Structure

```JSON
{
  "event_id": "",
  "event_name": "",
  "occurred_at": "",
  "aggregate_id": "",
  "payload": {}
}
```

---

## 9\.3 Event Version

必须：

```JSON
{
  "event_version": "1.0"
}
```

---

# Logging Standard

## 10\.1 Structured Logging

统一 JSON。

---

## 10\.2 Required Fields

```JSON
{
  "trace_id": "",
  "request_id": "",
  "user_id": ""
}
```

---

## 10\.3 Log Level

```Plain Text
DEBUG
INFO
WARN
ERROR
FATAL
```

---

# Testing Standard

## 11\.1 Test Pyramid

```Plain Text
E2E

Integration

Unit Test
```

---

## 11\.2 Coverage

核心领域：

≥ 90%

应用层：

≥ 80%

整体：

≥ 75%

---

## 11\.3 Mandatory Test Scope

必须覆盖：

- Billing

- Subscription

- Payment

- Strategy

- Formula

- Backtest

- Portfolio

---

# Git Workflow Standard

## 12\.1 Branch Strategy

```Plain Text
main

develop

feature/*
bugfix/*
release/*
hotfix/*
```

---

## 12\.2 Commit Convention

Conventional Commit

```Plain Text
feat:
fix:
refactor:
docs:
test:
chore:
```

---

## 12\.3 Pull Request

必须包含：

```Plain Text
Background

Change

Testing

Risk

Rollback Plan
```

---

# Code Review Standard

## 13\.1 Mandatory Review

所有代码必须 Review。

禁止直接提交 Main。

---

## 13\.2 Review Focus

重点检查：

- Architecture

- Domain Logic

- Security

- Performance

- Test Coverage

---

# CI/CD Standard

## 14\.1 Pipeline

```Plain Text
Lint

Unit Test

Integration Test

Build

Security Scan

Deploy
```

---

## 14\.2 Deployment Strategy

支持：

- Blue Green

- Canary

禁止：

直接覆盖生产环境。

---

# Security Standard

## 15\.1 Secret Management

禁止：

```Go
password := "123456"
```

必须：

Secret Manager

Environment Variable

Vault

---

## 15\.2 Authentication

统一：

OAuth2

JWT

RBAC

---

## 15\.3 Audit Log

必须记录：

- Login

- Permission Change

- Billing Action

- Subscription Action

---

# Documentation Standard

## 16\.1 Required Documents

每个服务必须包含：

```Plain Text
README.md

Architecture.md

API.md

DomainModel.md

ADR.md
```

---

## 16\.2 Architecture Decision Record

路径：

```Plain Text
docs/adr/
```

格式：

```Plain Text
Context

Decision

Alternatives

Consequences
```

---

# AI Coding Governance

## 17\.1 AI Generated Code

AI 生成代码不得直接进入生产环境。

必须：

- Lint

- Test

- Review

后方可合并。

---

## 17\.2 AI Prompt Archive

重要 Prompt 必须存档：

```Plain Text
docs/prompts/
```

用于知识沉淀。

---

# Engineering Quality Gates

代码合并前必须满足：

✓ Build Success

✓ Test Passed

✓ Lint Passed

✓ Review Approved

✓ Security Scan Passed

✓ Documentation Updated

否则禁止 Merge。

---

# Final Principle

代码是资产，不是消耗品。

任何设计与实现都必须优先考虑：

可维护性 \> 可扩展性 \> 性能 \> 开发速度

除非经过 ADR 审批，否则不得违反本原则。


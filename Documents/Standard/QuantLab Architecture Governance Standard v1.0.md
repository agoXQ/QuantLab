# QuantLab Architecture Governance Standard v1\.0

**Document Version:** v1\.0
**Project:** QuantLab
**Document Type:** Architecture Governance Standard
**Status:** Approved
**Last Updated:** 2026\-06

---

# Introduction

## 1\.1 Purpose

本文档用于定义 QuantLab 平台的统一架构治理规范（Architecture Governance）。

架构治理的目标不是限制开发，而是保证：

- 系统长期可演进

- 架构保持一致性

- 技术债可控

- 服务边界清晰

- 性能与成本平衡

- 多团队协作可持续

本规范高于项目级开发规范。

---

## 1\.2 Governance Scope

适用于：

```Plain Text
Backend Services
Microservices
Data Platform
Quant Engine
AI Platform
Billing Platform
Infrastructure Platform
```

---

## 1\.3 Architecture Principles

所有架构设计必须遵循以下原则。

---

### Principle 1：Business Driven

业务决定架构。

禁止：

```Plain Text
为了使用某技术而设计业务
```

必须：

```Plain Text
业务需求驱动技术选型
```

---

### Principle 2：Domain First

领域边界优先。

架构边界必须来源于领域模型。

禁止：

```Plain Text
按照数据库拆服务
按照开发人员拆服务
按照页面拆服务
```

---

### Principle 3：Evolutionary Architecture

架构必须允许演进。

禁止：

```Plain Text
一次性设计未来十年架构
```

提倡：

```Plain Text
当前最优
可扩展
可替换
```

---

### Principle 4：Loose Coupling

系统间低耦合。

优先：

```Plain Text
Event
API Contract
Interface
```

避免：

```Plain Text
共享数据库
共享代码
共享状态
```

---

### Principle 5：Observability First

不可观测的系统不可维护。

所有架构必须具备：

```Plain Text
Logging
Metrics
Tracing
Audit
```

---

# Architecture Governance Organization

## 2\.1 Architecture Ownership

架构决策分级管理。

---

### Level 1

System Architecture

负责：

```Plain Text
整体架构
技术路线
服务边界
核心技术栈
```

---

### Level 2

Domain Architecture

负责：

```Plain Text
领域拆分
聚合设计
事件设计
```

---

### Level 3

Service Architecture

负责：

```Plain Text
服务内部实现
模块设计
接口设计
```

---

# System Architecture Standard

## 3\.1 Target Architecture

统一采用：

```Plain Text
DDD
+
Clean Architecture
+
Event Driven Architecture
+
Microservice Architecture
```

---

## 3\.2 Logical Architecture

```Plain Text
Client Layer

↓

API Gateway

↓

Application Services

↓

Domain Services

↓

Infrastructure

↓

Database / MQ / Cache
```

---

## 3\.3 Service Classification

所有服务必须属于以下类别之一。

---

### Domain Service

核心业务服务。

例如：

```Plain Text
User Service
Billing Service
Strategy Service
Portfolio Service
```

---

### Platform Service

平台能力服务。

例如：

```Plain Text
Notification Service
Workflow Service
Auth Service
```

---

### Infrastructure Service

基础设施服务。

例如：

```Plain Text
Redis
Kafka
PostgreSQL
ClickHouse
```

---

### Engine Service

计算引擎。

例如：

```Plain Text
Backtest Engine
Formula Engine
Ranking Engine
```

---

# Bounded Context Governance

## 4\.1 Core Domains

QuantLab 核心领域划分：

```Plain Text
Identity Domain

Billing Domain

Strategy Domain

Formula Domain

Market Data Domain

Portfolio Domain

Backtest Domain

Ranking Domain

Community Domain

AI Domain

Notification Domain
```

---

## 4\.2 Domain Independence

领域之间禁止：

```Plain Text
直接访问数据库

共享 Entity

共享 Repository
```

允许：

```Plain Text
REST API

gRPC

Domain Event
```

---

## 4\.3 Cross Domain Communication

优先级：

```Plain Text
Domain Event
    ↓
gRPC
    ↓
REST API
```

---

# Service Governance Standard

## 5\.1 Service Boundary

一个服务只能负责一个领域。

禁止：

```Plain Text
User + Billing

Billing + Strategy

Factor + Portfolio
```

放入同一服务。

---

## 5\.2 Service Size Rule

推荐：

```Plain Text
5~20 Aggregate

10~50 API

5万~15万 LOC
```

---

## 5\.3 Shared Library Rule

允许共享：

```Plain Text
SDK

Common Protocol

Logging

Tracing
```

禁止共享：

```Plain Text
Business Logic

Domain Model

Repository
```

---

# Event Architecture Governance

## 6\.1 Event Classification

---

### Domain Event

领域内事件。

例如：

```Plain Text
UserCreated

OrderPaid

SubscriptionCreated
```

---

### Integration Event

跨领域事件。

例如：

```Plain Text
OrderPaid

MembershipExpired
```

---

### System Event

系统事件。

例如：

```Plain Text
DeploymentCompleted

BacktestFailed
```

---

## 6\.2 Event Naming Convention

统一格式：

```Plain Text
<Entity><Action>
```

与 Event Specification 保持一致。

例如：

```Plain Text
OrderPaid

UserCreated

BacktestCompleted
```

---

## 6\.3 Event Immutability

事件一经发布：

禁止修改。

必须：

```Plain Text
Version Upgrade
```

---

## 6\.4 Event Retention

保留时间：

```Plain Text
Domain Event:
365 Days

Audit Event:
7 Years

Operational Event:
90 Days
```

---

# API Governance Standard

## 7\.1 API First

开发流程：

```Plain Text
API Design

↓

Review

↓

OpenAPI

↓

Implementation
```

---

## 7\.2 API Compatibility

禁止：

```Plain Text
删除字段

修改字段类型
```

允许：

```Plain Text
新增字段
```

---

## 7\.3 API Versioning

统一：

```Plain Text
/api/v1
/api/v2
```

---

# Data Architecture Governance

## 8\.1 Data Ownership

一个数据只能属于一个领域。

例如：

```Plain Text
User → Identity Domain

Order → Billing Domain

Strategy → Strategy Domain

Portfolio → Portfolio Domain
```

---

## 8\.2 Database Ownership

一个服务拥有一个数据库。

禁止：

```Plain Text
多个服务共享数据库
```

---

## 8\.3 Data Synchronization

采用：

```Plain Text
CDC

Event

ETL
```

禁止：

```Plain Text
跨服务 Join
```

---

# Storage Governance

## 9\.1 PostgreSQL

适用于：

```Plain Text
OLTP
Transaction
Billing
User
```

---

## 9\.2 PostgreSQL（时序扩展）

适用于：

```Plain Text
Market Data (TimescaleDB)

Backtest Result

Analytics
```

---

## 9\.3 Redis

适用于：

```Plain Text
Cache

Session

Rate Limit

Distributed Lock
```

---

## 9\.4 Object Storage

适用于：

```Plain Text
Report

Export

Snapshot

Dataset
```

---

# Infrastructure Governance

## 10\.1 Containerization

所有服务必须容器化。

统一：

```Plain Text
Docker
```

---

## 10\.2 Orchestration

统一：

```Plain Text
Kubernetes
```

---

## 10\.3 Configuration

统一：

```Plain Text
Environment Variable

Config Service
```

禁止：

```Plain Text
硬编码配置
```

---

# Security Architecture Governance

## 11\.1 Zero Trust Principle

所有请求必须认证。

禁止：

```Plain Text
内网默认可信
```

---

## 11\.2 Authentication

统一：

```Plain Text
OAuth2

JWT
```

---

## 11\.3 Authorization

统一：

```Plain Text
RBAC
```

后续扩展：

```Plain Text
ABAC
```

---

## 11\.4 Secret Management

统一：

```Plain Text
Vault

Secret Manager
```

---

# Observability Governance

## 12\.1 Logging

统一：

```Plain Text
Structured JSON Logging
```

---

## 12\.2 Metrics

统一：

```Plain Text
Prometheus
```

---

## 12\.3 Tracing

统一：

```Plain Text
OpenTelemetry
```

---

## 12\.4 Dashboard

统一：

```Plain Text
Grafana
```

---

# Performance Governance

## 13\.1 Performance Budget

所有服务必须定义：

```Plain Text
Latency

QPS

Concurrency

Storage

Cost
```

---

## 13\.2 SLA

核心服务：

```Plain Text
99.9%
```

---

关键服务：

```Plain Text
99.95%
```

---

## 13\.3 Capacity Planning

每季度进行：

```Plain Text
Load Test

Stress Test

Capacity Review
```

---

# Architecture Review Board

## 14\.1 Mandatory Review

以下变更必须经过架构评审：

```Plain Text
新增微服务

新增数据库

新增MQ

新增缓存系统

领域边界调整

核心技术栈变更
```

---

## 14\.2 Review Output

必须产出：

```Plain Text
ADR

Architecture Diagram

Risk Assessment
```

---

# Architecture Decision Record \(ADR\)

## 15\.1 Mandatory ADR

以下事项必须记录 ADR：

```Plain Text
数据库选型

MQ选型

缓存选型

架构模式变更

核心框架变更

第三方服务引入
```

---

## 15\.2 ADR Template

```Plain Text
Title

Status

Context

Decision

Alternatives

Consequences

References
```

---

# Technical Debt Governance

## 16\.1 Debt Classification

P1

架构级问题

P2

服务级问题

P3

代码级问题

---

## 16\.2 Debt Tracking

统一记录：

```Plain Text
docs/tech-debt/
```

---

## 16\.3 Debt Review

每月评审一次。

---

# QuantLab Target Architecture

## 17\.1 Strategic Direction

未来架构目标：

```Plain Text
Cloud Native

Microservices

Event Driven

AI Native

Observability First

Multi Tenant Ready
```

---

## 17\.2 Architecture Evolution Path

Phase 1

```Plain Text
Modular Monolith
```

↓

Phase 2

```Plain Text
Domain Service Split
```

↓

Phase 3

```Plain Text
Event Driven Platform
```

↓

Phase 4

```Plain Text
AI Native Quant Platform
```

---

# Governance Enforcement

以下事项禁止上线：

- 无 ADR 的重大架构变更

- 无 Review 的服务拆分

- 跨领域数据库访问

- 未定义 SLA 的核心服务

- 不符合领域边界的服务设计

---

# Final Principle

架构不是为了展示技术能力。

架构存在的唯一目的：

支撑业务长期稳定演进。

任何技术决策必须满足：

Business Value \> Maintainability \> Scalability \> Performance \> Technical Preference

否则不得通过架构评审。


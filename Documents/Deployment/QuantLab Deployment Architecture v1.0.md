# QuantLab Deployment Architecture v1\.0

Version: 1\.0

Status: Architecture Baseline

Owner: Platform Architecture Team

Scope:

```Plain Text
QuantLab 全平台部署架构
MVP → Growth → Enterprise
```

覆盖：

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

Kafka
Redis
MySQL
Qdrant
Object Storage
API Gateway
Monitoring
CI/CD
```

---

# 架构目标

## MVP阶段

目标：

```Plain Text
1000 DAU

100万行情数据/天

1000次回测/天

100次AI调用/天
```

---

## Growth阶段

目标：

```Plain Text
10万注册用户

1万DAU

10万回测/天

10万AI调用/天
```

---

## Enterprise阶段

目标：

```Plain Text
100万用户

10万DAU

百万级回测任务

机构级客户
```

---

# 总体架构

```Plain Text
Internet
                       │
                       ▼
               ┌──────────────┐
               │  CDN / WAF   │
               └──────┬───────┘
                      │
                      ▼
              ┌───────────────┐
              │ API Gateway   │
              └──────┬────────┘
                     │
      ┌──────────────┼──────────────┐
      ▼              ▼              ▼

 User Service   Strategy Service   Portfolio Service

      ▼              ▼              ▼

 Formula Engine   Backtest Engine  Ranking Service

      ▼              ▼              ▼

 Community      Notification      AI Service

                     │
                     ▼

                 Kafka

                     │

     ┌───────────────┼───────────────┐

     ▼               ▼               ▼

   MySQL           Redis          Qdrant

                     │

                     ▼

              Object Storage
```

---

# 技术栈统一规范

## Backend

推荐：

```Plain Text
Go 1.25+
```

原因：

```Plain Text
高并发
低资源消耗
微服务生态成熟
```

---

## API

```Plain Text
REST + OpenAPI
```

---

## Event Bus

```Plain Text
Kafka
```

---

## Cache

```Plain Text
Redis Cluster
```

---

## Database

```Plain Text
MySQL 8.x
```

---

## Vector DB

```Plain Text
Qdrant
```

---

## Storage

```Plain Text
S3 Compatible
```

推荐：

```Plain Text
MinIO（私有化）

AWS S3（公有云）
```

---

# Kubernetes部署原则

全部服务容器化。

统一：

```Plain Text
Docker
+
Kubernetes
```

---

# K8S命名空间规划

```Plain Text
quantlab-prod

quantlab-staging

quantlab-dev
```

---

# Pod划分

## user\-service

```Plain Text
Deployment

Replicas: 2
```

---

## strategy\-service

```Plain Text
Deployment

Replicas: 3
```

---

## portfolio\-service

```Plain Text
Deployment

Replicas: 3
```

---

## formula\-engine

```Plain Text
Deployment

Replicas: 5
```

公式计算频繁。

---

## backtest\-engine

```Plain Text
Deployment

Replicas: 10
```

CPU密集。

---

## ranking\-service

```Plain Text
Deployment

Replicas: 2
```

---

## community\-service

```Plain Text
Deployment

Replicas: 2
```

---

## notification\-service

```Plain Text
Deployment

Replicas: 3
```

---

## ai\-service

```Plain Text
Deployment

Replicas: 5
```

---

# API Gateway

推荐：

```Plain Text
Kong Gateway
```

或者：

```Plain Text
APISIX
```

---

职责：

```Plain Text
认证

限流

审计

灰度发布

统一路由
```

---

统一入口：

```Plain Text
api.quantlab.com
```

---

# Service Mesh（未来）

MVP：

```Plain Text
不使用
```

---

Growth：

推荐：

```Plain Text
Istio
```

---

提供：

```Plain Text
mTLS

Traffic Split

Tracing
```

---

# 数据库部署

## MySQL集群

推荐：

```Plain Text
Primary
+
2 Replica
```

---

拓扑：

```Plain Text
MySQL Primary
        │
        ▼
 Replica-1

 Replica-2
```

---

职责：

```Plain Text
写主库

读从库
```

---

# 数据库拆分策略

MVP：

```Plain Text
单实例
```

---

Growth：

拆分：

```Plain Text
quantlab_user

quantlab_strategy

quantlab_portfolio

quantlab_backtest

quantlab_community

quantlab_ai
```

---

# Redis部署

推荐：

```Plain Text
Redis Cluster
```

---

节点：

```Plain Text
3 Master

3 Replica
```

---

职责：

```Plain Text
权限缓存

排行榜缓存

通知缓存

会话缓存

AI缓存
```

---

# Kafka部署

推荐：

```Plain Text
Kafka KRaft
```

不再依赖 ZooKeeper。

---

节点：

```Plain Text
3 Broker
```

MVP。

---

Growth：

```Plain Text
5 Broker
```

---

# Topic规划

```Plain Text
user-events

strategy-events

portfolio-events

backtest-events

ranking-events

community-events

notification-events

ai-events
```

---

# Kafka保留策略

事件：

```Plain Text
7 Days
```

---

DLQ：

```Plain Text
30 Days
```

---

审计：

```Plain Text
90 Days
```

---

# Qdrant部署

Collection：

```Plain Text
strategy_docs

portfolio_docs

backtest_docs

community_docs
```

---

MVP：

```Plain Text
单节点
```

---

Growth：

```Plain Text
3节点集群
```

---

# 对象存储

存储：

```Plain Text
头像

附件

回测导出

AI报告

社区图片
```

---

目录规划：

```Plain Text
/user

/strategy

/portfolio

/backtest

/community

/ai
```

---

# 行情数据架构（重点）

Market Data是平台核心。

---

数据来源：

```Plain Text
聚宽
Tushare
AkShare
交易所
第三方供应商
```

---

架构：

```Plain Text
Data Collector
      ↓
Raw Storage
      ↓
Normalize
      ↓
Market Data Service
```

---

# 行情存储设计

不要全部放MySQL。

推荐：

```Plain Text
ClickHouse
```

---

原因：

```Plain Text
时序查询快

OLAP优秀

压缩率高
```

---

# 数据库职责

MySQL：

```Plain Text
业务数据
```

---

ClickHouse：

```Plain Text
行情数据

K线

因子

回测结果
```

---

# Backtest集群

Backtest Engine建议独立。

---

原因：

```Plain Text
CPU密集

耗时长

容易拖垮业务服务
```

---

部署：

```Plain Text
Backtest Worker Pool
```

---

拓扑：

```Plain Text
Backtest Scheduler
        │
        ▼
Kafka Queue
        │
        ▼
Backtest Workers
```

---

# AI架构

推荐：

```Plain Text
AI Gateway
```

统一管理：

```Plain Text
OpenAI

Claude

DeepSeek

Qwen
```

---

拓扑：

```Plain Text
AI Service
      │
      ▼
LLM Router
      │
 ┌────┼────┐
 ▼    ▼    ▼

GPT Claude DeepSeek
```

---

# AI成本隔离

单独命名空间：

```Plain Text
quantlab-ai
```

---

避免：

```Plain Text
AI爆量
影响核心业务
```

---

# 可观测性架构

统一：

```Plain Text
Metrics
Logs
Tracing
```

---

# Metrics

推荐：

```Plain Text
Prometheus
```

---

指标：

```Plain Text
QPS

Latency

CPU

Memory

Kafka Lag
```

---

# Dashboard

推荐：

```Plain Text
Grafana
```

---

统一大盘：

```Plain Text
System

Business

AI

Backtest
```

---

# 日志架构

推荐：

```Plain Text
Loki
```

或者：

```Plain Text
ELK
```

---

格式：

```Plain Text
{
  "trace_id":"",
  "user_id":"",
  "service":"",
  "message":""
}
```

---

# Tracing

推荐：

```Plain Text
OpenTelemetry
```

---

采集：

```Plain Text
HTTP

Kafka

DB
```

---

展示：

```Plain Text
Jaeger
```

---

# CI/CD架构

推荐：

```Plain Text
GitHub
+
GitHub Actions
```

---

流程：

```Plain Text
Commit
 ↓
Test
 ↓
Build
 ↓
Docker
 ↓
Push Registry
 ↓
Deploy
```

---

# 环境规划

## Dev

```Plain Text
开发环境
```

---

## Staging

```Plain Text
预发布环境
```

---

## Production

```Plain Text
生产环境
```

---

严格隔离。

---

# 发布策略

推荐：

```Plain Text
Blue Green
```

---

未来：

```Plain Text
Canary
```

---

# 安全架构

统一：

```Plain Text
HTTPS

TLS1.3
```

---

内部：

```Plain Text
mTLS
```

未来开启。

---

# Secrets管理

禁止：

```Plain Text
配置文件明文密码
```

---

推荐：

```Plain Text
Vault
```

或者：

```Plain Text
K8S Secret
```

---

# 备份策略

MySQL：

```Plain Text
每日全量

每小时Binlog
```

---

Redis：

```Plain Text
RDB
+
AOF
```

---

Kafka：

```Plain Text
Mirror Backup
```

---

# 容灾等级

RPO：

```Plain Text
≤ 5分钟
```

---

RTO：

```Plain Text
≤ 30分钟
```

---

# 成本控制

MVP阶段建议：

```Plain Text
4 Core

8 GB

×3节点
```

即可运行。

---

预计：

```Plain Text
200~500 USD/月
```

公有云。

---

# MVP部署拓扑（推荐）

```Plain Text
K8S Cluster (3 Nodes)

Node-1
 ├ User
 ├ Strategy
 ├ Portfolio
 ├ Community

Node-2
 ├ Formula
 ├ Ranking
 ├ Notification

Node-3
 ├ Backtest
 ├ AI
 ├ Market Data
```

---

中间件：

```Plain Text
MySQL

Redis

Kafka

MinIO
```

单独部署。

---

# Growth阶段拓扑

```Plain Text
业务集群

计算集群

AI集群
```

分离。

---

```Plain Text
Business Cluster
```

运行：

```Plain Text
User

Strategy

Portfolio

Community
```

---

```Plain Text
Compute Cluster
```

运行：

```Plain Text
Backtest

Formula

Ranking
```

---

```Plain Text
AI Cluster
```

运行：

```Plain Text
AI Service

Qdrant
```

---

# Enterprise阶段

增加：

```Plain Text
Multi Region
```

例如：

```Plain Text
Tokyo

Singapore

Hong Kong
```

---

支持：

```Plain Text
Active-Passive
```

容灾。

---

# 当前架构缺失的一块

经过前面所有设计，我们实际上已经有：

```Plain Text
User
Strategy
Portfolio
Backtest
Ranking
Community
Notification
AI
```

但缺少：

```Plain Text
Billing Service
```

以及：

```Plain Text
Subscription Service
```

---

这两个服务未来会负责：

```Plain Text
会员付费

策略订阅

组合订阅

创作者收益

佣金结算
```




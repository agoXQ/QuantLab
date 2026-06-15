# QuantLab Portfolio Service TD

Version: 1\.0

Module: Portfolio Service

Priority: P0（Core Domain）

Status: Draft

Owner: Architecture Team

Dependencies:

```Plain Text
User Service
Strategy Service
Backtest Engine
Ranking Service
Community Service
Notification Service
```

---

# 服务定位

Portfolio Service 是 QuantLab 的核心领域服务之一。

负责：

```Plain Text
组合管理
策略配置
权重管理
调仓规则
组合发布
组合生命周期
```

不负责：

```Plain Text
策略执行
组合回测计算
排行榜计算
社区互动
```

这些能力分别由：

```Plain Text
Backtest Engine
Ranking Service
Community Service
```

完成。

---

# 服务职责边界

## 拥有数据

```Plain Text
Portfolio
PortfolioItem
PortfolioVersion
RebalanceRule
PortfolioSnapshot
```

---

## 引用数据

```Plain Text
Strategy
User
```

---

## 输出数据

```Plain Text
Portfolio Metadata

Portfolio Composition

Portfolio Allocation
```

---

# 目录结构

```Plain Text
portfolio-service
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
│   │   ├── portfolio
│   │   ├── rebalance
│   │   └── version
│   │
│   ├── infrastructure
│   │   ├── persistence
│   │   ├── cache
│   │   └── mq
│   │
│   ├── event
│   │
│   └── scheduler
│
├── configs
│
└── deploy
```

---

# DDD聚合设计

---

## Portfolio Aggregate

聚合根

```Plain Text
type Portfolio struct {
    ID
    OwnerID

    Name
    Description

    Visibility
    Status

    Version

    CreatedAt
    UpdatedAt
}
```

---

## PortfolioItem Entity

```Plain Text
type PortfolioItem struct {
    PortfolioID

    StrategyID

    Weight

    SortOrder
}
```

---

## RebalanceRule Entity

```Plain Text
type RebalanceRule struct {
    PortfolioID

    Frequency

    Method

    Threshold
}
```

---

## PortfolioVersion Entity

```Plain Text
type PortfolioVersion struct {
    PortfolioID

    Version

    Snapshot
}
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_portfolio
```

---

## portfolio

组合主表

```Plain Text
CREATE TABLE portfolio
(
    id BIGINT PRIMARY KEY,

    owner_id BIGINT NOT NULL,

    name VARCHAR(128) NOT NULL,

    description TEXT,

    visibility VARCHAR(32),

    status VARCHAR(32),

    current_version INT,

    created_at DATETIME,

    updated_at DATETIME,

    deleted_at DATETIME NULL,

    INDEX idx_owner(owner_id),
    INDEX idx_status(status)
);
```

---

## portfolio\_item

组合成分

```Plain Text
CREATE TABLE portfolio_item
(
    id BIGINT PRIMARY KEY,

    portfolio_id BIGINT NOT NULL,

    strategy_id BIGINT NOT NULL,

    weight DECIMAL(8,4),

    sort_order INT,

    created_at DATETIME,

    INDEX idx_portfolio(portfolio_id),
    INDEX idx_strategy(strategy_id)
);
```

---

## portfolio\_version

版本快照

```Plain Text
CREATE TABLE portfolio_version
(
    id BIGINT PRIMARY KEY,

    portfolio_id BIGINT,

    version INT,

    snapshot_json JSON,

    created_at DATETIME,

    UNIQUE KEY uk_version(portfolio_id, version)
);
```

---

## rebalance\_rule

调仓规则

```Plain Text
CREATE TABLE rebalance_rule
(
    id BIGINT PRIMARY KEY,

    portfolio_id BIGINT,

    frequency VARCHAR(32),

    method VARCHAR(32),

    threshold DECIMAL(8,4),

    created_at DATETIME
);
```

---

## portfolio\_publish

发布记录

```Plain Text
CREATE TABLE portfolio_publish
(
    id BIGINT PRIMARY KEY,

    portfolio_id BIGINT,

    version INT,

    published_at DATETIME
);
```

---

## portfolio\_snapshot

每日快照

用于：

```Plain Text
收益曲线
历史分析
排名追踪
```

---

```Plain Text
CREATE TABLE portfolio_snapshot
(
    portfolio_id BIGINT,

    trade_date DATE,

    nav DECIMAL(18,6),

    return_rate DECIMAL(18,6),

    drawdown DECIMAL(18,6),

    PRIMARY KEY(portfolio_id, trade_date)
);
```

---

# 状态模型

统一平台状态：

```Plain Text
DRAFT

PUBLISHED

ARCHIVED

DELETED
```

---

状态流转：

```Plain Text
DRAFT
   ↓
PUBLISHED
   ↓
ARCHIVED
```

---

# 可见性模型

```Plain Text
PRIVATE

PUBLIC

UNLISTED
```

---

# 权重模型

## Manual Weight

```Plain Text
手动配置
```

---

## Equal Weight

```Plain Text
自动等权
```

---

## Risk Weight

```Plain Text
风险平价
```

---

未来：

```Plain Text
AI Weight
```

---

# 权重校验

要求：

```Plain Text
Σ(weight)=100%
```

---

误差：

```Plain Text
±0.001
```

---

自动归一化：

```Plain Text
Normalize()
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/portfolios
```

---

## 创建组合

```Plain Text
POST /
```

Request：

```Plain Text
{
  "name":"价值成长组合",
  "description":"长期投资",
  "visibility":"PUBLIC"
}
```

---

Response：

```Plain Text
{
  "portfolio_id":10001
}
```

---

## 获取组合

```Plain Text
GET /{id}
```

---

## 更新组合

```Plain Text
PUT /{id}
```

---

## 删除组合

```Plain Text
DELETE /{id}
```

软删除。

---

## 添加策略

```Plain Text
POST /{id}/items
```

---

Request：

```Plain Text
{
  "strategy_id":2001,
  "weight":40
}
```

---

## 删除策略

```Plain Text
DELETE /{id}/items/{itemId}
```

---

## 调整权重

```Plain Text
PUT /{id}/weights
```

Request：

```Plain Text
{
  "items":[
    {
      "strategy_id":1,
      "weight":60
    },
    {
      "strategy_id":2,
      "weight":40
    }
  ]
}
```

---

## 创建版本

```Plain Text
POST /{id}/versions
```

---

## 获取版本列表

```Plain Text
GET /{id}/versions
```

---

## 回滚版本

```Plain Text
POST /{id}/rollback
```

---

## 发布组合

```Plain Text
POST /{id}/publish
```

---

## 创建回测任务

```Plain Text
POST /{id}/backtests
```

实际调用：

```Plain Text
Backtest Engine
```

---

## 获取分析报告

```Plain Text
GET /{id}/analytics
```

---

返回：

```Plain Text
{
  "annual_return":0.21,
  "sharpe":1.4,
  "max_drawdown":0.18
}
```

---

## 获取净值曲线

```Plain Text
GET /{id}/equity-curve
```

---

## 获取持仓构成

```Plain Text
GET /{id}/composition
```

---

# 事件设计

Topic：

```Plain Text
portfolio-events
```

---

## PortfolioCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioCreated",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "producer": "portfolio-service",
  "payload": {
    "portfolio_id": 1001,
    "owner_id": 2001
  }
}
```

---

## PortfolioUpdated

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioUpdated",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "payload": {
    "portfolio_id": 1001
  }
}
```

---

## PortfolioPublished

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioPublished",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "payload": {
    "portfolio_id": 1001,
    "version": 5
  }
}
```

---

## PortfolioItemAdded

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioItemAdded",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "payload": {
    "portfolio_id": 1001,
    "strategy_id": 3001
  }
}
```

---

## PortfolioWeightsUpdated

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioWeightsUpdated",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "payload": {
    "portfolio_id": 1001
  }
}
```

---

## PortfolioBacktestCompleted

来自 Backtest Engine：

```JSON
{
  "event_id": "uuid",
  "event_type": "PortfolioBacktestCompleted",
  "aggregate_type": "PORTFOLIO",
  "aggregate_id": "1001",
  "payload": {
    "portfolio_id": 1001,
    "annual_return": 0.25
  }
}
```

---

# 事件消费者

## Community Service

监听：

```Plain Text
PortfolioPublished
```

创建：

```Plain Text
Content(PORTFOLIO)
```

---

## Ranking Service

监听：

```Plain Text
PortfolioBacktestCompleted
```

更新排行榜。

---

## Notification Service

监听：

```Plain Text
PortfolioPublished
```

通知关注者。

---

## AI Service

监听：

```Plain Text
PortfolioCreated

PortfolioUpdated
```

生成分析。

---

# Redis缓存设计

---

组合详情

```Plain Text
portfolio:{id}
```

TTL：

```Plain Text
30min
```

---

组合构成

```Plain Text
portfolio:composition:{id}
```

TTL：

```Plain Text
10min
```

---

分析结果

```Plain Text
portfolio:analytics:{id}
```

TTL：

```Plain Text
1h
```

---

热门组合

```Plain Text
portfolio:hot
```

TTL：

```Plain Text
1min
```

---

# 权限设计

## Free

```Plain Text
最多3个组合

最多20个策略
```

---

## Pro

```Plain Text
最多50个组合

最多200个策略
```

---

## Master

```Plain Text
无限制
```

---

# 配额控制

创建组合：

```Plain Text
Rate Limit

10/min
```

---

发布组合：

```Plain Text
5/min
```

---

回测请求：

```Plain Text
20/day
```

Free用户。

---

# 部署架构

```Plain Text
Portfolio API
      ↓
Portfolio Service
      ↓
PostgreSQL
      ↓
Redis
      ↓
Kafka
```

---

# K8S部署

```Plain Text
portfolio-api

3 Pods

↓

portfolio-worker

5 Pods

↓

scheduler

2 Pods
```

---

# 高可用设计

PostgreSQL：

```Plain Text
1 Primary
2 Replica
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

# 性能指标

组合查询：

```Plain Text
P95 < 30ms
```

---

组合更新：

```Plain Text
P95 < 100ms
```

---

创建组合：

```Plain Text
P95 < 200ms
```

---

QPS：

```Plain Text
2000+
```

---

# 可观测性

Metrics：

```Plain Text
portfolio_created_total

portfolio_published_total

portfolio_backtest_total

portfolio_rebalance_total
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
RTO
15min
```

```Plain Text
RPO
5min
```

---

# MVP范围

必须实现：

✅ Portfolio

✅ PortfolioItem

✅ 权重配置

✅ PortfolioVersion

✅ Portfolio Publish

✅ Portfolio Analytics

✅ Community集成

✅ Ranking集成

---

暂缓：

❌ 自动调仓

❌ AI配置权重

❌ 跟投

❌ 订阅

❌ 实盘同步

---

# 后续演进

V2 支持组合嵌套（Fund of Funds）：

```Plain Text
type PortfolioItem struct {
    ItemType   // STRATEGY | PORTFOLIO
    ItemID
    Weight
}
```

示例：

```Plain Text
稳健组合
├── 红利组合 40%
├── ETF轮动组合 30%
└── 成长组合 30%
```

此抽象使 Portfolio Service 从策略容器升级为真正的资产配置引擎。


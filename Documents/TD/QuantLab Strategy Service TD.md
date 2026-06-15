# QuantLab Strategy Service TD

Version: 1\.0

Module: Strategy Service

Priority: P0

Status: Draft

Owner: Architecture Team

---

# 服务定位

Strategy Service 负责管理平台所有策略资产。

职责：

```Plain Text
策略创建
策略编辑
策略版本管理
策略发布
策略标签
策略可见性
策略复制
策略收藏统计
策略生命周期管理
```

不负责：

```Plain Text
公式解析
回测执行
排行榜计算
社区评论
```

---

# 服务边界（Bounded Context）

拥有：

```Plain Text
Strategy
StrategyVersion
StrategyTag
StrategyPublish
StrategyFork
StrategyFavoriteSnapshot
```

只读引用：

```Plain Text
User
Membership
```

禁止拥有：

```Plain Text
Backtest
Ranking
Comment
Article
```

---

# DDD聚合设计

---

## Aggregate \#1

### Strategy Aggregate

聚合根：

```Plain Text
Strategy
```

包含：

```Plain Text
StrategyMeta
StrategyConfig
StrategyVisibility
```

---

职责：

```Plain Text
创建
编辑
删除
发布
归档
```

---

## Aggregate \#2

### StrategyVersion Aggregate

聚合根：

```Plain Text
StrategyVersion
```

---

职责：

```Plain Text
保存快照
版本回滚
版本比较
```

---

## Aggregate \#3

### StrategyFork Aggregate

聚合根：

```Plain Text
StrategyFork
```

---

职责：

```Plain Text
复制策略
记录来源
构建传播链
```

---

# 策略生命周期

```Plain Text
Draft
 ↓
Configured
 ↓
Backtested
 ↓
Published
 ↓
Archived
```

---

## Draft

刚创建，未配置完成。

允许：

```Plain Text
编辑
删除
保存
```

---

## Configured

规则配置完成，可以回测。

---

## Backtested

至少执行过一次回测。

---

## Published

已公开，参与社区与排行榜。

允许：

```Plain Text
浏览
收藏
复制
评论
```

---

## Archived

废弃策略，不参与榜单。

禁止：

```Plain Text
新用户访问
```

---

# 技术架构

```Plain Text
strategy-service

├── api
├── application
├── domain
├── infrastructure
├── event
├── cache
├── search
└── scheduler
```

---

# 目录结构

```Plain Text
strategy-service
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── application
│   │   ├── command
│   │   ├── query
│   │   └── service
│   │
│   ├── domain
│   │   ├── strategy
│   │   ├── version
│   │   ├── publish
│   │   ├── fork
│   │   └── tag
│   │
│   ├── infrastructure
│   │   ├── postgresql
│   │   ├── redis
│   │   ├── kafka
│   │   └── elasticsearch
│   │
│   ├── event
│   │
│   └── scheduler
│
└── deploy
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_strategy
```

---

## strategy表

策略主表

```Plain Text
CREATE TABLE strategy
(
    id BIGINT PRIMARY KEY,

    author_id BIGINT,

    current_version_id BIGINT,

    title VARCHAR(256),

    description TEXT,

    status VARCHAR(32),

    visibility VARCHAR(32),

    category VARCHAR(64),

    created_at DATETIME,

    updated_at DATETIME,

    deleted_at DATETIME NULL
);
```

---

## strategy\_version表

核心表

```Plain Text
CREATE TABLE strategy_version
(
    id BIGINT PRIMARY KEY,

    strategy_id BIGINT,

    version_no VARCHAR(32),

    formula_text LONGTEXT,

    buy_rule LONGTEXT,

    sell_rule LONGTEXT,

    risk_rule LONGTEXT,

    parameter_json JSON,

    change_log TEXT,

    created_by BIGINT,

    created_at DATETIME
);
```

---

## strategy\_publish表

```Plain Text
CREATE TABLE strategy_publish
(
    id BIGINT PRIMARY KEY,

    strategy_id BIGINT,

    version_id BIGINT,

    publish_time DATETIME,

    publish_status VARCHAR(32)
);
```

---

## strategy\_tag表

```Plain Text
CREATE TABLE strategy_tag
(
    strategy_id BIGINT,

    tag_id BIGINT,

    PRIMARY KEY(strategy_id, tag_id)
);
```

---

## tag表

```Plain Text
CREATE TABLE tag
(
    id BIGINT PRIMARY KEY,

    name VARCHAR(64),

    category VARCHAR(64)
);
```

---

## strategy\_fork表

```Plain Text
CREATE TABLE strategy_fork
(
    id BIGINT PRIMARY KEY,

    source_strategy_id BIGINT,

    target_strategy_id BIGINT,

    creator_id BIGINT,

    created_at DATETIME
);
```

---

## strategy\_stat表

统计快照表

```Plain Text
CREATE TABLE strategy_stat
(
    strategy_id BIGINT PRIMARY KEY,

    view_count BIGINT,

    favorite_count BIGINT,

    fork_count BIGINT,

    comment_count BIGINT
);
```

---

## strategy\_snapshot表

用于排行榜冻结

```Plain Text
CREATE TABLE strategy_snapshot
(
    id BIGINT PRIMARY KEY,

    strategy_id BIGINT,

    version_id BIGINT,

    snapshot_json JSON,

    created_at DATETIME
);
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/strategies
```

---

## 创建策略

```Plain Text
POST /
```

Request

```Plain Text
{
  "title":"ROE价值选股",
  "description":"..."
}
```

---

## 获取策略

```Plain Text
GET /{strategyId}
```

---

## 更新策略

```Plain Text
PUT /{strategyId}
```

---

## 保存版本

```Plain Text
POST /{strategyId}/versions
```

---

## 获取版本列表

```Plain Text
GET /{strategyId}/versions
```

---

## 获取指定版本

```Plain Text
GET /versions/{versionId}
```

---

## 发布策略

```Plain Text
POST /{strategyId}/publish
```

---

## 归档策略

```Plain Text
POST /{strategyId}/archive
```

---

## Fork策略

```Plain Text
POST /{strategyId}/fork
```

---

## 搜索策略

```Plain Text
GET /search
```

参数：

```Plain Text
keyword
tag
author
category
sort
```

---

## 获取热门策略

```Plain Text
GET /popular
```

---

## 获取作者策略

```Plain Text
GET /users/{uid}/strategies
```

---

# 事件设计

Topic：

```Plain Text
strategy-events
```

---

## StrategyCreated

```Plain Text
{
  "event":"StrategyCreated",
  "strategy_id":1001,
  "author_id":2001
}
```

---

## StrategyUpdated

```Plain Text
{
  "event":"StrategyUpdated",
  "strategy_id":1001
}
```

---

## StrategyVersionCreated

```Plain Text
{
  "event":"StrategyVersionCreated",
  "strategy_id":1001,
  "version":"v1.2"
}
```

---

## StrategyPublished

```Plain Text
{
  "event":"StrategyPublished",
  "strategy_id":1001,
  "version_id":3001
}
```

---

消费者：

```Plain Text
Community Service
Ranking Service
Notification Service
```

---

## StrategyArchived

```Plain Text
{
  "event":"StrategyArchived",
  "strategy_id":1001
}
```

---

## StrategyForked

```Plain Text
{
  "event":"StrategyForked",
  "source_strategy_id":1,
  "target_strategy_id":2
}
```

---

# 缓存设计

Redis

---

策略详情：

```Plain Text
strategy:{id}
```

TTL：

```Plain Text
30min
```

---

策略统计：

```Plain Text
strategy:stat:{id}
```

TTL：

```Plain Text
10min
```

---

作者策略列表：

```Plain Text
user:strategies:{uid}
```

TTL：

```Plain Text
15min
```

---

热门策略：

```Plain Text
strategy:popular
```

TTL：

```Plain Text
5min
```

---

# 搜索设计

ElasticSearch

索引：

```Plain Text
strategy_index
```

---

字段：

```Plain Text
{
  "strategy_id":1,
  "title":"",
  "description":"",
  "tags":[],
  "author_id":1,
  "status":"PUBLISHED"
}
```

---

支持：

```Plain Text
全文搜索
标签搜索
作者搜索
组合搜索
```

---

# 权限设计

---

Role

```Plain Text
USER
CREATOR
Pro
ADMIN
```

---

Permission

```Plain Text
strategy:create

strategy:update

strategy:publish

strategy:archive

strategy:fork
```

---

发布限制：

```Plain Text
Creator以上
```

---

高级策略数量限制：

```Plain Text
Membership控制
```

---

# 数据一致性设计

采用：

```Plain Text
Transaction Outbox Pattern
```

---

流程：

```Plain Text
PostgreSQL Commit
      ↓
Outbox Event
      ↓
Kafka
```

---

避免：

```Plain Text
数据已提交
事件丢失
```

---

# 部署架构

```Plain Text
API Gateway
      ↓
Strategy Service

      ↓

PostgreSQL Cluster

Redis Cluster

ElasticSearch Cluster

Kafka Cluster
```

---

# 高可用设计

K8S

```Plain Text
3 Pods
```

---

PostgreSQL

```Plain Text
1 Primary
2 Replica
```

---

Redis

```Plain Text
3 Master
3 Replica
```

---

ElasticSearch

```Plain Text
3 Node
```

---

Kafka

```Plain Text
3 Broker
```

---

# 性能指标

创建策略：

```Plain Text
P95 < 100ms
```

---

获取策略：

```Plain Text
P95 < 50ms
```

---

搜索策略：

```Plain Text
P95 < 300ms
```

---

发布策略：

```Plain Text
P95 < 150ms
```

---

QPS目标：

```Plain Text
查询
10000 QPS

创建
1000 QPS

搜索
3000 QPS
```

---

# 可观测性

Metrics

```Plain Text
strategy_created_total

strategy_published_total

strategy_fork_total

strategy_search_total
```

---

Trace

```Plain Text
OpenTelemetry
```

---

Log

```Plain Text
JSON Structured Log
```

---

# SLA

```Plain Text
Availability
99.95%
```

---

```Plain Text
RTO
30分钟
```

---

```Plain Text
RPO
5分钟
```

---

# MVP范围

必须实现：

✅ 创建策略

✅ 编辑策略

✅ 版本管理

✅ 发布策略

✅ Fork策略

✅ 标签体系

✅ Elasticsearch搜索

✅ Kafka事件

✅ Redis缓存

---

暂缓：

❌ 协同编辑

❌ Git风格Diff

❌ 策略市场收费

❌ 策略授权交易

❌ 团队策略空间

---

# 后续演进

当策略规模达到 100 万级别时，可考虑将 Strategy Service 拆分为：

```Plain Text
strategy-service
│
├── strategy-core-service
│
├── strategy-version-service
│
└── strategy-search-service
```

版本数据增长速度远大于策略元数据，独立拆分有利于存储与查询的独立扩展。


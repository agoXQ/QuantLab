# QuantLab Database Design Standard v1.0

**Document Version:** v1.0
**Project:** QuantLab
**Document Type:** Database Design Standard
**Status:** Approved
**Last Updated:** 2026-06

---

# Introduction

## 1.1 Purpose

本文档定义 QuantLab 平台统一数据库设计规范。

目标：

- 保证数据模型一致性
- 支撑量化数据的高效存取
- 保证回测结果可复现
- 降低跨服务数据耦合
- 支撑数据版本管理与审计

---

## 1.2 Scope

适用于：

```Plain Text
PostgreSQL (OLTP)
TimescaleDB (时序数据)
Redis (缓存 / 队列)
Object Storage (报告 / 快照)
```

---

## 1.3 Design Principles

### Principle 1：One Database Per Service

每个微服务拥有独立数据库。

禁止跨服务直接访问数据库。

---

### Principle 2：Schema As Contract

数据库 Schema 是服务间数据契约。

变更必须通过 Migration 管理。

---

### Principle 3：Time-Series First

量化数据以时间序列为核心。

K 线、因子、财务数据均按时间维度组织。

---

### Principle 4：Read / Write Separation

写路径优先一致性与完整性。

读路径优先性能与缓存策略。

---

### Principle 5：Data Version Binding

所有回测相关数据必须绑定数据版本。

保证回测结果可复现。

---

# Naming Convention

## 2.1 Database Naming

格式：

```Plain Text
quantlab_<service>
```

示例：

```Plain Text
quantlab_user
quantlab_strategy
quantlab_billing
quantlab_market_data
quantlab_backtest
quantlab_portfolio
quantlab_community
quantlab_ranking
quantlab_notification
quantlab_ai
```

---

## 2.2 Schema Naming

默认使用 `public` schema。

多租户场景使用：

```Plain Text
tenant_<tenant_id>
```

---

## 2.3 Table Naming

统一：

```Plain Text
snake_case
单数名词
```

正确：

```SQL
strategy
backtest_job
portfolio_item
market_bar
financial_statement
```

错误：

```SQL
Strategies
backtestJobs
tbl_strategy
```

---

## 2.4 Column Naming

统一：

```Plain Text
snake_case
```

正确：

```SQL
created_at
user_id
strategy_id
max_drawdown
annual_return
```

错误：

```SQL
createdAt
userId
MaxDrawdown
```

---

## 2.5 Index Naming

格式：

```Plain Text
idx_<table>_<column>
```

示例：

```SQL
idx_strategy_owner_id
idx_backtest_job_status
idx_market_bar_symbol_date
```

唯一索引：

```Plain Text
uq_<table>_<column>
```

示例：

```SQL
uq_user_email
uq_security_code
```

---

## 2.6 Foreign Key Naming

格式：

```Plain Text
fk_<table>_<ref_table>
```

示例：

```SQL
fk_strategy_user
fk_portfolio_item_strategy
```

---

# Data Type Standard

## 3.1 Primary Key

统一使用：

```SQL
BIGINT
```

推荐 Snowflake ID。

---

## 3.2 External ID

对外暴露的 ID 使用：

```SQL
VARCHAR(32)
```

带前缀，例如：

```Plain Text
usr_01HJY...
stg_01HJY...
ord_01HJY...
```

---

## 3.3 Timestamp

统一使用：

```SQL
TIMESTAMPTZ
```

所有时间以 UTC 存储。

---

## 3.4 Boolean

统一使用：

```SQL
BOOLEAN
```

命名以 `is_` / `has_` / `can_` 开头。

---

## 3.5 Monetary Amount

统一使用：

```SQL
DECIMAL(20, 4)
```

禁止使用 FLOAT / DOUBLE 存储金额。

---

## 3.6 Price / Ratio

价格与比率使用：

```SQL
DECIMAL(20, 6)
```

---

## 3.7 Enum

使用：

```SQL
VARCHAR(32)
```

配合 CHECK 约束。

禁止使用 PostgreSQL ENUM 类型（不可扩展）。

---

## 3.8 JSON

仅用于非查询字段：

```SQL
JSONB
```

禁止对 JSONB 内部字段建立索引（除非使用 GIN）。

---

# Audit Fields Standard

## 4.1 Required Fields

所有业务表必须包含：

```SQL
created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
created_by    BIGINT
updated_by    BIGINT
```

---

## 4.2 Soft Delete

统一使用：

```SQL
deleted_at    TIMESTAMPTZ
```

禁止使用 `is_deleted BOOLEAN`。

查询时统一过滤：

```SQL
WHERE deleted_at IS NULL
```

---

## 4.3 Version Field

需要乐观锁的表：

```SQL
version    INTEGER NOT NULL DEFAULT 1
```

更新时：

```SQL
UPDATE ... SET version = version + 1 WHERE version = ?
```

---

# Time-Series Data Standard

## 5.1 Market Data Table

K 线数据使用 TimescaleDB Hypertable：

```SQL
CREATE TABLE market_bar (
    symbol       VARCHAR(16)  NOT NULL,
    trade_date   DATE         NOT NULL,
    period       VARCHAR(8)   NOT NULL,  -- day / week / month
    open         DECIMAL(20,6),
    high         DECIMAL(20,6),
    low          DECIMAL(20,6),
    close        DECIMAL(20,6),
    volume       BIGINT,
    amount       DECIMAL(20,4),
    adj_factor   DECIMAL(20,6),
    data_version VARCHAR(16)  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('market_bar', 'trade_date');

CREATE UNIQUE INDEX uq_market_bar
    ON market_bar (symbol, trade_date, period, data_version);
```

---

## 5.2 Financial Data Table

```SQL
CREATE TABLE financial_statement (
    symbol         VARCHAR(16)  NOT NULL,
    report_date    DATE         NOT NULL,
    report_type    VARCHAR(16)  NOT NULL,  -- annual / quarter
    revenue        DECIMAL(20,4),
    net_profit     DECIMAL(20,4),
    total_assets   DECIMAL(20,4),
    total_liability DECIMAL(20,4),
    net_assets     DECIMAL(20,4),
    data_version   VARCHAR(16)  NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('financial_statement', 'report_date');

CREATE UNIQUE INDEX uq_financial_statement
    ON financial_statement (symbol, report_date, report_type, data_version);
```

---

## 5.3 Factor Data Table

```SQL
CREATE TABLE factor_data (
    symbol       VARCHAR(16)  NOT NULL,
    trade_date   DATE         NOT NULL,
    factor_name  VARCHAR(64)  NOT NULL,
    factor_value DECIMAL(20,6),
    data_version VARCHAR(16)  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('factor_data', 'trade_date');

CREATE UNIQUE INDEX uq_factor_data
    ON factor_data (symbol, trade_date, factor_name, data_version);
```

---

## 5.4 Data Version Table

```SQL
CREATE TABLE data_version (
    version     VARCHAR(16) PRIMARY KEY,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

所有回测必须绑定 `data_version`。

---

# Partitioning Strategy

## 6.1 Time-Based Partition

时序数据按日期范围分区：

```SQL
-- 按月分区
CREATE TABLE market_bar_2026_01 PARTITION OF market_bar
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

---

## 6.2 Retention Policy

```SQL
-- K 线数据保留 10 年
SELECT add_retention_policy('market_bar', INTERVAL '10 years');

-- 因子数据保留 5 年
SELECT add_retention_policy('factor_data', INTERVAL '5 years');
```

---

# Index Strategy

## 7.1 Primary Index

每表必须有主键。

---

## 7.2 Query-Driven Index

索引设计基于查询模式，而非表结构。

常见查询模式：

```SQL
-- 按用户查询策略
CREATE INDEX idx_strategy_owner_status ON strategy (owner_id, status);

-- 按日期范围查询 K 线
CREATE INDEX idx_market_bar_date ON market_bar (trade_date DESC);

-- 按状态查询回测任务
CREATE INDEX idx_backtest_job_status_created ON backtest_job (status, created_at);
```

---

## 7.3 Covering Index

高频查询使用覆盖索引：

```SQL
CREATE INDEX idx_strategy_list
    ON strategy (owner_id, status, created_at DESC)
    INCLUDE (name, category);
```

---

## 7.4 Partial Index

```SQL
-- 仅索引公开策略
CREATE INDEX idx_strategy_public
    ON strategy (created_at DESC)
    WHERE visibility = 'PUBLIC' AND deleted_at IS NULL;
```

---

## 7.5 Index Rules

- 单表索引不超过 5 个（时序表除外）
- 禁止冗余索引
- 禁止未使用的索引
- 所有外键必须建索引

---

# Migration Management

## 8.1 Migration Tool

统一使用 Alembic（Python 服务）。

---

## 8.2 Migration Naming

格式：

```Plain Text
<YYYYMMDD>_<description>.py
```

示例：

```Plain Text
20260614_add_strategy_tags.py
20260615_create_portfolio_table.py
```

---

## 8.3 Migration Rules

- 每个 Migration 只做一件事
- Migration 必须可回滚（`downgrade`）
- 禁止在 Migration 中执行数据迁移（使用独立脚本）
- 禁止手动修改数据库 Schema

---

## 8.4 Zero-Downtime Migration

大表变更必须遵循：

1. 新增列（允许 NULL 或有默认值）
2. 应用代码同时支持新旧列
3. 回填数据
4. 移除旧列

禁止直接 `ALTER TABLE ... ADD COLUMN NOT NULL`。

---

# Connection Management

## 9.1 Connection Pool

统一使用连接池：

```Plain Text
Min: 2
Max: 20
Idle Timeout: 10 min
Max Lifetime: 30 min
```

---

## 9.2 Timeout

```Plain Text
Statement Timeout: 30s
Idle In Transaction Timeout: 60s
```

---

## 9.3 Read / Write Split

读写分离：

```Plain Text
Primary: Write
Replica: Read
```

ORM 层面自动路由。

---

# Redis Usage Standard

## 10.1 Key Naming

格式：

```Plain Text
<service>:<entity>:<id>
```

示例：

```Plain Text
strategy:cache:stg_001
user:session:usr_001
ranking:snapshot:return:daily
backtest:job:bt_001
```

---

## 10.2 TTL Strategy

```Plain Text
Session: 24h
Cache: 5min - 1h
Rate Limit: 1min
Ranking Snapshot: 24h
Distributed Lock: 30s
```

---

## 10.3 Data Structure Selection

```Plain Text
String: 简单缓存
Hash: 对象缓存
Sorted Set: 排行榜
List: 消息队列
String + SETNX: 分布式锁
```

---

## 10.4 Prohibited Patterns

- 禁止使用 `KEYS *`
- 禁止存储大于 1MB 的值
- 禁止将 Redis 作为持久化存储
- 禁止在 Lua 脚本中执行耗时操作

---

# Backup & Recovery

## 11.1 Backup Strategy

```Plain Text
PostgreSQL: 每日全量 + 持续 WAL 归档
Redis: RDB 每 6 小时 + AOF
```

---

## 11.2 Retention

```Plain Text
Daily Backup: 30 天
Weekly Backup: 12 周
Monthly Backup: 12 个月
```

---

## 11.3 Recovery SLA

```Plain Text
RPO: < 1 小时
RTO: < 4 小时
```

---

## 11.4 Backup Verification

每月执行一次恢复演练。

---

# Multi-Tenancy Data Isolation

## 12.1 MVP Strategy

MVP 阶段采用：

```Plain Text
Schema Per Tenant
```

---

## 12.2 Future Strategy

规模化后支持：

```Plain Text
Database Per Tenant
```

---

## 12.3 Tenant Context

所有查询必须带 Tenant ID：

```SQL
WHERE tenant_id = ?
```

禁止跨租户查询。

---

# Data Archival

## 13.1 Archival Policy

```Plain Text
Backtest Result: 2 年后归档
Notification: 180 天后归档
Audit Log: 7 年后归档
```

---

## 13.2 Archival Storage

归档至 Object Storage（S3 兼容）。

---

# Performance Guidelines

## 14.1 Query Rules

- 禁止 `SELECT *`
- 禁止无 LIMIT 的查询
- 禁止在 WHERE 子句中对列使用函数
- 批量写入使用 `COPY` 或批量 INSERT

---

## 14.2 Batch Size

```Plain Text
INSERT: 1000 rows/batch
UPDATE: 500 rows/batch
DELETE: 1000 rows/batch
```

---

## 14.3 Slow Query Threshold

```Plain Text
OLTP: > 100ms
OLAP: > 5s
```

慢查询必须记录并优化。

---

# Security

## 15.1 Access Control

- 应用账号仅授予必要权限（CRUD，禁止 DDL）
- Migration 账号独立管理
- 禁止使用 superuser 运行应用

---

## 15.2 Sensitive Data

- 密码使用 bcrypt 哈希
- PII 数据加密存储
- API Key 使用 SHA-256 哈希存储

---

## 15.3 Connection Security

- 所有连接使用 TLS
- 禁止明文传输数据库密码

---

# Monitoring

## 16.1 Required Metrics

```Plain Text
Connection Count
Active Queries
Slow Queries
Deadlock Count
Cache Hit Ratio
Replication Lag
Disk Usage
```

---

## 16.2 Alert Rules

```Plain Text
Connection > 80%: Warning
Replication Lag > 10s: Critical
Disk Usage > 85%: Warning
Deadlock > 0: Warning
```

---

# Final Principle

数据库是业务模型的物理投影。

Schema 设计必须优先反映领域模型，而非技术便利。

任何 Schema 变更必须满足：

Domain Alignment > Data Integrity > Query Performance > Storage Efficiency

未经 Migration Review 不得直接修改生产数据库。

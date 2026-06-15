# QuantLab Market Data Service TD

Version: 1\.0

Module: Market Data Service

Priority: P0

Status: Draft

Owner: Architecture Team

---

# 服务定位

Market Data Service 是 QuantLab 的统一数据中心（Single Source Of Truth）。

负责：

```Plain Text
行情数据管理

财务数据管理

因子数据管理

指数数据管理

数据清洗

复权处理

数据版本管理

数据服务接口
```

不负责：

```Plain Text
回测计算

公式解析

策略管理

排行榜计算
```

---

# 核心原则

原则1

所有服务只能从 Market Data 获取数据

---

原则2

禁止直接访问第三方数据源

---

原则3

数据版本固定

---

原则4

支持历史回溯

---

原则5

统一复权逻辑

---

# 服务边界

拥有：

```Plain Text
Security

MarketBar

FinancialStatement

Factor

Index

TradingCalendar

CorporateAction

DataVersion
```

不拥有：

```Plain Text
Strategy

Backtest

Ranking
```

---

# 数据架构

```Plain Text
Data Provider

        ↓

Ingestion Service

        ↓

Cleaner

        ↓

Adjustment Engine

        ↓

Storage

        ↓

Market Data API
```

---

# 数据来源（V1）

建议：

```Plain Text
AkShare
Tushare
BaoStock
```

组合方案。

---

# V2数据源

支持：

```Plain Text
Wind

Choice

JoinQuant

RiceQuant
```

---

# 数据分类

---

## 行情数据

```Plain Text
日线

周线

月线

分钟线
```

---

## 财务数据

```Plain Text
利润表

资产负债表

现金流量表
```

---

## 因子数据

```Plain Text
PE

PB

PS

ROE

ROA

EPS
```

---

## 指数数据

```Plain Text
沪深300

中证500

创业板指
```

---

# DDD聚合设计

---

## Security Aggregate

聚合根：

```Plain Text
Security
```

---

包含：

```Plain Text
BasicInfo

Industry

Exchange
```

---

## MarketData Aggregate

聚合根：

```Plain Text
MarketBar
```

---

## Financial Aggregate

聚合根：

```Plain Text
FinancialStatement
```

---

## Factor Aggregate

聚合根：

```Plain Text
Factor
```

---

## DataVersion Aggregate

聚合根：

```Plain Text
DataVersion
```

---

# 技术架构

```Plain Text
market-data-service

├── ingestion
├── cleaner
├── adjustment
├── storage
├── query
├── cache
├── event
└── api
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_market
```

---

## security表

```Plain Text
CREATE TABLE security
(
    id BIGINT PRIMARY KEY,

    stock_code VARCHAR(32),

    stock_name VARCHAR(128),

    exchange VARCHAR(32),

    industry VARCHAR(128),

    listing_date DATE,

    delisting_date DATE,

    status VARCHAR(32)
);
```

---

## trading\_calendar表

```Plain Text
CREATE TABLE trading_calendar
(
    trade_date DATE PRIMARY KEY,

    is_open TINYINT
);
```

---

## market\_bar表

核心表（使用 TimescaleDB Hypertable）

```SQL
CREATE TABLE market_bar
(
    stock_code VARCHAR(32),

    trade_date DATE,

    period VARCHAR(8),

    open_price DECIMAL(20,6),

    high_price DECIMAL(20,6),

    low_price DECIMAL(20,6),

    close_price DECIMAL(20,6),

    volume BIGINT,

    amount DECIMAL(20,4),

    adj_factor DECIMAL(20,6),

    data_version VARCHAR(16),

    PRIMARY KEY(stock_code, trade_date, period)
);

SELECT create_hypertable('market_bar', 'trade_date');
```

---

## market\_bar\_adjusted表

前复权数据

```Plain Text
CREATE TABLE market_bar_adjusted
(
    stock_code VARCHAR(32),

    trade_date DATE,

    close_price DECIMAL(20,4),

    adjustment_factor DECIMAL(20,8),

    PRIMARY KEY(stock_code, trade_date)
);
```

---

## corporate\_action表

```Plain Text
CREATE TABLE corporate_action
(
    id BIGINT PRIMARY KEY,

    stock_code VARCHAR(32),

    action_type VARCHAR(64),

    action_date DATE,

    factor DECIMAL(20,8)
);
```

---

## financial\_statement表

```SQL
CREATE TABLE financial_statement
(
    stock_code VARCHAR(32),

    report_date DATE,

    report_type VARCHAR(16),

    revenue DECIMAL(20,4),

    net_profit DECIMAL(20,4),

    total_assets DECIMAL(20,4),

    total_liabilities DECIMAL(20,4),

    net_assets DECIMAL(20,4),

    data_version VARCHAR(16),

    PRIMARY KEY(stock_code, report_date, report_type)
);

SELECT create_hypertable('financial_statement', 'report_date');
```

---

## factor\_data表

核心表（使用 TimescaleDB Hypertable）

```SQL
CREATE TABLE factor_data
(
    stock_code VARCHAR(32),

    trade_date DATE,

    factor_name VARCHAR(64),

    factor_value DECIMAL(20,6),

    data_version VARCHAR(16),

    PRIMARY KEY(stock_code, trade_date, factor_name)
);

SELECT create_hypertable('factor_data', 'trade_date');
```

---

## index\_bar表

```SQL
CREATE TABLE index_bar
(
    index_code VARCHAR(32),

    trade_date DATE,

    close_price DECIMAL(20,6),

    data_version VARCHAR(16),

    PRIMARY KEY(index_code, trade_date)
);

SELECT create_hypertable('index_bar', 'trade_date');
```

---

## data\_version表

```SQL
CREATE TABLE data_version
(
    version VARCHAR(16) PRIMARY KEY,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

# 数据版本机制

例如：

```Plain Text
2026.01

2026.02

2026.03
```

---

回测记录：

```Plain Text
使用数据版本：

2026.03
```

---

保证：

```Plain Text
结果可复现
```

---

# 数据清洗引擎

负责：

```Plain Text
缺失值

异常值

重复数据

停牌数据
```

处理。

---

# 复权引擎

支持：

```Plain Text
不复权

前复权

后复权
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/market
```

---

## 获取股票

```Plain Text
GET /securities/{code}
```

---

## 获取K线

```Plain Text
GET /bars
```

参数：

```Plain Text
code

start

end

adjustment
```



---

## 获取财务数据

```Plain Text
GET /financials
```

---

## 获取因子数据

```Plain Text
GET /factors
```

---

## 获取指数数据

```Plain Text
GET /indexes
```

---

## 获取交易日历

```Plain Text
GET /calendar
```

---

## 获取数据版本

```Plain Text
GET /versions
```

---

# 事件设计

Topic：

```Plain Text
market-events
```

Market Data Service 的事件主要用于内部数据更新通知，供 Backtest Engine 和 AI Service 消费。

---

## MarketDataUpdated

```JSON
{
  "event_id": "uuid",
  "event_type": "MarketDataUpdated",
  "aggregate_type": "SECURITY",
  "producer": "market-data-service",
  "payload": {
    "trade_date": "2026-06-01"
  }
}
```

---

## FactorUpdated

```JSON
{
  "event_id": "uuid",
  "event_type": "FactorUpdated",
  "aggregate_type": "SECURITY",
  "producer": "market-data-service",
  "payload": {
    "trade_date": "2026-06-01"
  }
}
```

---

## DataVersionCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "DataVersionCreated",
  "aggregate_type": "SECURITY",
  "producer": "market-data-service",
  "payload": {
    "version": "2026.06"
  }
}
```

---

消费者：

```Plain Text
Backtest Engine

AI Service
```

---

# Redis缓存设计

热门股票：

```Plain Text
market:security:{code}
```

TTL：

```Plain Text
24h
```

---

最近K线：

```Plain Text
market:bar:{code}
```

TTL：

```Plain Text
1h
```

---

因子缓存：

```Plain Text
market:factor:{code}:{date}
```

TTL：

```Plain Text
12h
```

---

# 时序数据库设计（重要）

时序数据使用 TimescaleDB（PostgreSQL 扩展）。

建议：

```Plain Text
Security
Financial

→ PostgreSQL

KLine

→ TimescaleDB

分钟线

→ TimescaleDB
```

---

推荐：

```Plain Text
PostgreSQL

+

TimescaleDB
```

架构。

---

## TimescaleDB Hypertable 设计

```SQL
CREATE TABLE market_bar
(
    stock_code VARCHAR(32),

    trade_date DATE NOT NULL,

    open_price DECIMAL(20,6),

    high_price DECIMAL(20,6),

    low_price DECIMAL(20,6),

    close_price DECIMAL(20,6),

    volume BIGINT,

    data_version VARCHAR(16)
);

SELECT create_hypertable('market_bar', 'trade_date');
CREATE INDEX idx_market_bar_symbol ON market_bar (stock_code, trade_date DESC);
```

---

# 查询优化

支持：

```Plain Text
按股票

按日期

按行业

按指数成分
```

查询。

---

# 数据同步任务

调度：

```Plain Text
每日17:00

同步当日数据
```

---

# 数据校验任务

检查：

```Plain Text
缺失

异常

重复
```

---

# 部署架构

```Plain Text
Data Provider

↓

Ingestion Worker

↓

Market Data Service

↓

PostgreSQL

↓

TimescaleDB

↓

Redis

↓

Kafka
```

---

# K8S部署

```Plain Text
market-api

3 Pods

↓

ingestion-worker

5 Pods
```

---

# 高可用设计

PostgreSQL

```Plain Text
1 Primary

2 Replica
```

---

TimescaleDB

```Plain Text
3 Node Cluster
```

---

Redis

```Plain Text
Cluster
```

---

Kafka

```Plain Text
3 Broker
```

---

# 权限设计

权限：

```Plain Text
market:query

market:financial

market:factor
```

---

会员限制：

Free

```Plain Text
最近3年数据
```

---

Pro

```Plain Text
最近10年数据
```

---

Master

```Plain Text
全量数据
```

---

# 性能指标

K线查询：

```Plain Text
P95 < 50ms
```

---

因子查询：

```Plain Text
P95 < 100ms
```

---

财务查询：

```Plain Text
P95 < 150ms
```

---

QPS目标：

```Plain Text
50000+
```

---

# 可观测性

Metrics：

```Plain Text
market_query_total

market_query_latency

data_sync_total

data_quality_error_total
```

---

Trace：

```Plain Text
OpenTelemetry
```

---

# SLA

```Plain Text
Availability

99.99%
```

---

```Plain Text
RTO

15分钟
```

---

```Plain Text
RPO

1分钟
```

---

# MVP范围

必须实现：

✅ 股票基础信息

✅ 日线数据

✅ 财务数据

✅ 因子数据

✅ 前复权

✅ 数据版本

✅ Redis缓存

✅ TimescaleDB存储

---

暂缓：

❌ Tick数据

❌ Level2数据

❌ 实时行情

❌ 期货

❌ 期权

❌ 港股

❌ 美股

---

# 后续演进

当数据规模达到千万级证券 × 十年历史数据时，可拆分为独立数据中台：

```Plain Text
market-data-platform
│
├── security-service
├── market-bar-service
├── factor-service
├── financial-service
├── adjustment-service
└── data-quality-service
```


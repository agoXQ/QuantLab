# QuantLab Backtest Engine TD

Version: 1\.0

Module: Backtest Engine

Priority: P0

Status: Draft

Owner: Architecture Team

---

# 服务定位

Backtest Engine 负责：

```Plain Text
策略执行

股票筛选

交易撮合

仓位管理

资金管理

绩效分析

回测报告生成
```

不负责：

```Plain Text
用户管理

公式解析

市场数据采集

排行榜计算
```

---

# 服务边界（Bounded Context）

拥有：

```Plain Text
BacktestJob
Portfolio
Position
Order
Trade
PerformanceReport
ExecutionContext
```

只读：

```Plain Text
Strategy
FormulaPlan
MarketData
```

---

# 核心设计原则

原则1

结果可复现

---

原则2

数据版本固定

---

原则3

回测环境隔离

---

原则4

事件驱动执行

---

原则5

支持水平扩展

---

# 整体架构

```Plain Text
Backtest API
        │
        ▼
Job Scheduler
        │
        ▼
Execution Manager
        │
        ▼
Strategy Executor
        │
        ▼
Portfolio Engine
        │
        ▼
Matching Engine
        │
        ▼
Performance Engine
        │
        ▼
Report Generator
```

---

# 目录结构

```Plain Text
backtest-engine
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── scheduler
│   │
│   ├── executor
│   │
│   ├── portfolio
│   │
│   ├── matching
│   │
│   ├── performance
│   │
│   ├── report
│   │
│   ├── event
│   │
│   ├── cache
│   │
│   └── repository
│
└── deploy
```

---

# 回测生命周期

```Plain Text
Created
 ↓
Queued
 ↓
Running
 ↓
Completed
 ↓
Archived
```

---

## 状态定义

Created：创建完成，等待入队。

Queued：已入队，等待 Worker 执行。

Running：执行中。

Completed：回测成功完成。

Failed：回测失败。

Archived：已归档。

---

# DDD聚合设计

---

## BacktestJob Aggregate

聚合根：

```Plain Text
BacktestJob
```

包含：

```Plain Text
Config
ExecutionContext
Status
```

---

## Portfolio Aggregate

聚合根：

```Plain Text
Portfolio
```

包含：

```Plain Text
Cash

Position

Trade
```

---

## Report Aggregate

聚合根：

```Plain Text
PerformanceReport
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_backtest
```

---

## backtest\_job

```Plain Text
CREATE TABLE backtest_job
(
    id BIGINT PRIMARY KEY,

    strategy_id BIGINT,

    version_id BIGINT,

    user_id BIGINT,

    status VARCHAR(32),

    start_date DATE,

    end_date DATE,

    benchmark VARCHAR(32),

    initial_capital DECIMAL(20,2),

    created_at DATETIME,

    finished_at DATETIME
);
```

---

## backtest\_config

```Plain Text
CREATE TABLE backtest_config
(
    job_id BIGINT PRIMARY KEY,

    commission_rate DECIMAL(10,6),

    slippage_rate DECIMAL(10,6),

    tax_rate DECIMAL(10,6),

    rebalance_period VARCHAR(32),

    max_position_count INT
);
```

---

## portfolio

```Plain Text
CREATE TABLE portfolio
(
    id BIGINT PRIMARY KEY,

    job_id BIGINT,

    cash DECIMAL(20,2),

    total_asset DECIMAL(20,2),

    market_value DECIMAL(20,2),

    trade_date DATE
);
```

---

## position

```Plain Text
CREATE TABLE position
(
    id BIGINT PRIMARY KEY,

    job_id BIGINT,

    stock_code VARCHAR(32),

    quantity BIGINT,

    cost_price DECIMAL(20,4),

    market_price DECIMAL(20,4),

    market_value DECIMAL(20,2),

    trade_date DATE
);
```

---

## order

```Plain Text
CREATE TABLE orders
(
    id BIGINT PRIMARY KEY,

    job_id BIGINT,

    stock_code VARCHAR(32),

    side VARCHAR(16),

    quantity BIGINT,

    price DECIMAL(20,4),

    status VARCHAR(32),

    created_at DATETIME,

    deleted_at DATETIME NULL
);
```

---

## trade

```Plain Text
CREATE TABLE trades
(
    id BIGINT PRIMARY KEY,

    order_id BIGINT,

    stock_code VARCHAR(32),

    quantity BIGINT,

    price DECIMAL(20,4),

    commission DECIMAL(20,4),

    trade_time DATETIME
);
```

---

## performance\_report

```Plain Text
CREATE TABLE performance_report
(
    job_id BIGINT PRIMARY KEY,

    annual_return DECIMAL(10,4),

    total_return DECIMAL(10,4),

    sharpe_ratio DECIMAL(10,4),

    max_drawdown DECIMAL(10,4),

    win_rate DECIMAL(10,4),

    volatility DECIMAL(10,4),

    beta DECIMAL(10,4),

    alpha DECIMAL(10,4)
);
```

---

# 回测执行流程

```Plain Text
加载策略

↓

加载执行计划

↓

加载历史数据

↓

逐日回放

↓

生成信号

↓

生成订单

↓

撮合成交

↓

更新持仓

↓

更新资金

↓

计算绩效

↓

生成报告
```

---

# Market Replay Engine

核心能力：

历史市场重放。

```Plain Text
2022-01-01

↓

2022-01-02

↓

2022-01-03
```

逐日推进。

---

## 执行上下文

ExecutionContext

```Plain Text
type ExecutionContext struct {
    CurrentDate time.Time

    Universe []string

    Portfolio Portfolio

    MarketData MarketSnapshot
}
```

---

## 策略执行器

输入：

```Plain Text
ExecutionPlan
```

输出：

```Plain Text
BuySignal

SellSignal
```

---

## 信号模型

```Plain Text
type Signal struct {
    StockCode string

    Action string

    Score float64
}
```

---

## 订单生成器

负责：

```Plain Text
信号

↓

订单
```

---

支持：

```Plain Text
Equal Weight

Risk Weight

Custom Weight
```

---

## 撮合引擎

V1规则：

```Plain Text
次日开盘成交
```

---

V2支持：

```Plain Text
当日收盘

VWAP

TWAP
```

---

## 滑点模型

固定滑点

```Plain Text
0.1%
```

---

动态滑点

```Plain Text
成交量相关
```

---

## 手续费模型

A股默认：

```Plain Text
佣金

印花税

过户费
```

参数化配置。

---

## 停牌处理

规则：

```Plain Text
停牌不可交易
```

---

## 涨跌停处理

规则：

```Plain Text
涨停不可买

跌停不可卖
```

---

# 绩效分析模块

计算：

```Plain Text
总收益率

年化收益率

超额收益

最大回撤

夏普比率

卡玛比率

索提诺比率

波动率
```

---

# 夏普比率

使用：

Sharpe=Rp−RfσpSharpe=\\frac\{R\_p\-R\_f\}\{\\sigma\_p\}Sharpe=σpRp−Rf

---

# 最大回撤

使用：

MDD=max⁡\(Peak−TroughPeak\)MDD=\\max\\left\(\\frac\{Peak\-Trough\}\{Peak\}\\right\)MDD=max\(PeakPeak−Trough\)

---

# 报告生成

输出：

```Plain Text
{
  "summary":{},
  "returns":[],
  "drawdown":[],
  "positions":[]
}
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/backtests
```

---

## 创建回测

```Plain Text
POST /
```

Request

```Plain Text
{
  "strategy_id":1001,

  "version_id":2001,

  "start_date":"2020-01-01",

  "end_date":"2024-01-01"
}
```

---

## 查询回测

```Plain Text
GET /{jobId}
```

---

## 获取回测报告

```Plain Text
GET /{jobId}/report
```

---

## 获取交易记录

```Plain Text
GET /{jobId}/trades
```

---

## 获取持仓历史

```Plain Text
GET /{jobId}/positions
```

---

## 取消回测

```Plain Text
POST /{jobId}/cancel
```

---

# 事件设计

Topic：

```Plain Text
backtest-events
```

---

## BacktestCreated

```Plain Text
{
  "job_id":1
}
```

---

## BacktestStarted

```Plain Text
{
  "job_id":1
}
```

---

## BacktestFinished

```Plain Text
{
  "job_id":1,

  "strategy_id":1001,

  "return":0.42,

  "sharpe":2.1
}
```

---

消费者：

```Plain Text
Ranking Service

Notification Service

AI Service
```

---

## BacktestFailed

```Plain Text
{
  "job_id":1,

  "error":"OUT_OF_MEMORY"
}
```

---

# Redis缓存设计

任务状态：

```Plain Text
backtest:job:{id}
```

---

报告缓存：

```Plain Text
backtest:report:{id}
```

---

TTL：

```Plain Text
24h
```

---

# 权限设计

权限：

```Plain Text
backtest:create

backtest:view

backtest:cancel
```

---

会员限制：

Free

```Plain Text
并发1个任务
```

---

Pro

```Plain Text
并发5个任务
```

---

Master

```Plain Text
并发20个任务
```

---

# 分布式执行架构（重点）

不要同步执行。

采用：

```Plain Text
API

↓

Kafka

↓

Backtest Worker

↓

Result Store
```

---

Worker池：

```Plain Text
worker-1

worker-2

worker-3

...
```

---

# K8S部署

```Plain Text
backtest-api

3 Pods

↓

backtest-worker

20 Pods

自动扩容
```

---

# 高可用设计

PostgreSQL

```Plain Text
1 Primary

2 Replica
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

# 性能指标

创建任务：

```Plain Text
P95 < 100ms
```

---

状态查询：

```Plain Text
P95 < 50ms
```

---

单任务：

```Plain Text
5000支股票

5年数据

10分钟内完成
```

---

目标吞吐：

```Plain Text
1000并发回测
```

---

# 可观测性

Metrics

```Plain Text
backtest_job_total

backtest_running_total

backtest_success_total

backtest_fail_total

backtest_duration_seconds
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

✅ 历史数据回放

✅ 股票筛选

✅ 买卖信号执行

✅ 持仓管理

✅ 绩效分析

✅ 回测报告

✅ Kafka任务队列

✅ 分布式Worker

---

暂缓：

❌ Tick级回测

❌ 期货回测

❌ 期权回测

❌ 高频回测

❌ GPU加速

---

# 后续演进

当回测并发达到万级时，可拆分为独立子服务：

```Plain Text
backtest-engine
│
├── execution-service
├── matching-service
├── portfolio-service
├── performance-service
└── report-service
```

核心链路：

```Plain Text
Strategy DSL → Formula Engine → Execution Plan → Backtest Cluster → Performance Engine → Ranking Service
```


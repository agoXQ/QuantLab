# QuantLab Formula Engine TD

Version: 1\.0

Module: Formula Engine

Priority: P0

Owner: Architecture Team

Status: Draft

---

# 服务定位

Formula Engine 负责：

```Plain Text
DSL解析
DSL校验
AST生成
语义分析
执行计划生成
公式优化
函数管理
```

不负责：

```Plain Text
策略管理
回测执行
市场数据读取
排行榜计算
```

---

# 核心目标

实现：

```Plain Text
DSL

↓

AST

↓

Logical Plan

↓

Execution Plan
```

供：

```Plain Text
Backtest Engine

Stock Screening Engine

AI Service
```

统一使用。

---

# 服务边界

拥有：

```Plain Text
Formula
AST
Function Registry
Compiler
Validator
Optimizer
```

不拥有：

```Plain Text
Strategy
Backtest
MarketData
```

---

# 总体架构

```Plain Text
formula-engine

├── lexer
├── parser
├── ast
├── validator
├── optimizer
├── compiler
├── planner
├── registry
├── cache
└── api
```

---

# 编译流程

```Plain Text
DSL Source

↓

Lexer

↓

Tokens

↓

Parser

↓

AST

↓

Validator

↓

Optimizer

↓

Execution Plan

↓

Backtest Engine
```

---

# 目录结构

```Plain Text
formula-engine
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── lexer
│   │
│   ├── parser
│   │
│   ├── ast
│   │
│   ├── validator
│   │
│   ├── optimizer
│   │
│   ├── planner
│   │
│   ├── compiler
│   │
│   ├── registry
│   │
│   ├── cache
│   │
│   └── event
│
└── deploy
```

---

# DSL输入示例

```Plain Text
ROE > 15

AND

PE < 20

AND

MA(CLOSE,5) > MA(CLOSE,20)
```

---

# Token设计

Token类型：

```Plain Text
IDENTIFIER

NUMBER

STRING

FUNCTION

OPERATOR

LPAREN

RPAREN

COMMA
```

---

# Lexer输出

```Plain Text
[
  {"type":"IDENTIFIER","value":"ROE"},
  {"type":"GT","value":">"},
  {"type":"NUMBER","value":"15"}
]
```

---

# AST设计

顶层：

```Plain Text
type Node interface{}
```

---

# BinaryExpression

```Plain Text
type BinaryExpression struct {
    Left Node
    Op   string
    Right Node
}
```

---

# FunctionCall

```Plain Text
type FunctionCall struct {
    Name string
    Args []Node
}
```

---

# Identifier

```Plain Text
type Identifier struct {
    Name string
}
```

---

# Literal

```Plain Text
type Literal struct {
    Value any
}
```

---

# AST示例

公式：

```Plain Text
ROE > 15
```

AST：

```Plain Text
{
  "type":"BinaryExpression",
  "op":">",
  "left":"ROE",
  "right":15
}
```

---

# Validator

负责：

```Plain Text
语法校验

函数校验

参数校验

类型校验

未来函数检测
```

---

# 未来函数检测

例如：

```Plain Text
REF(CLOSE,-1)
```

直接拒绝。

---

错误：

```Plain Text
{
  "code":"FUTURE_FUNCTION",
  "message":"Future data reference detected"
}
```

---

# 类型检查

禁止：

```Plain Text
PE > "ABC"
```

---

返回：

```Plain Text
{
  "code":"TYPE_ERROR"
}
```

---

# 函数注册中心

核心组件：

```Plain Text
Function Registry
```

---

# Function定义

```Plain Text
type FunctionDefinition struct {
    Name string

    Args []ArgDef

    ReturnType string

    Category string
}
```

---

## 分类

```Plain Text
Technical

Financial

Logical

Math

TimeSeries
```

---

## 技术指标函数

```Plain Text
MA

EMA

SMA

MACD

RSI

KDJ
```

---

## 财务函数

```Plain Text
PE

PB

ROE

ROA

EPS

RevenueGrowth
```

---

## 时间序列函数

```Plain Text
REF

HHV

LLV

COUNT

BARSLAST
```

---

## 优化器

目标：

减少计算量。

---

## Constant Folding

输入：

```Plain Text
1+2+3
```

优化：

```Plain Text
6
```

---

## Boolean Simplify

输入：

```Plain Text
A AND TRUE
```

优化：

```Plain Text
A
```

---

## Dead Branch Remove

输入：

```Plain Text
FALSE AND A
```

优化：

```Plain Text
FALSE
```

---

# Execution Plan

输出给：

```Plain Text
Backtest Engine
```

---

结构：

```Plain Text
{
  "plan_type":"FILTER",

  "conditions":[]
}
```

---

# Execution Node

```Plain Text
type PlanNode struct {
    Type string

    Children []PlanNode
}
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/formula
```

---

## 校验公式

```Plain Text
POST /validate
```

Request

```Plain Text
{
  "formula":"ROE > 15"
}
```

---

Response

```Plain Text
{
  "valid":true
}
```

---

## 编译公式

```Plain Text
POST /compile
```

---

返回：

```Plain Text
{
  "ast":{},
  "plan":{}
}
```

---

## AST查看

```Plain Text
POST /ast
```

---

返回：

```Plain Text
{
  "ast":{}
}
```

---

## 获取函数列表

```Plain Text
GET /functions
```

---

## 获取函数详情

```Plain Text
GET /functions/{name}
```

---

# 事件设计

Topic：

```Plain Text
strategy-events
```

Formula Engine 事件发布至 strategy-events（与 Strategy Service 共享 Topic）。

---

## FormulaValidated

```JSON
{
  "event_id": "uuid",
  "event_type": "FormulaValidated",
  "aggregate_type": "STRATEGY",
  "producer": "formula-engine",
  "payload": {
    "formula_hash": "",
    "valid": true
  }
}
```

---

## FormulaCompileFailed

```JSON
{
  "event_id": "uuid",
  "event_type": "FormulaCompileFailed",
  "aggregate_type": "STRATEGY",
  "producer": "formula-engine",
  "payload": {
    "error": "TYPE_ERROR"
  }
}
```

---

# 缓存设计

Redis

---

AST缓存：

```Plain Text
formula:ast:{hash}
```

---

执行计划缓存：

```Plain Text
formula:plan:{hash}
```

---

TTL：

```Plain Text
24h
```

---

## 缓存命中策略

用户提交：

```Plain Text
ROE > 15
```

---

计算：

```Plain Text
SHA256(formula)
```

---

作为Key。

---

# 数据库设计

数据库：

```Plain Text
quantlab_formula
```

---

## function\_registry表

```Plain Text
CREATE TABLE function_registry
(
    id BIGINT PRIMARY KEY,

    name VARCHAR(64),

    category VARCHAR(64),

    return_type VARCHAR(64),

    description TEXT
);
```

---

## function\_parameter表

```Plain Text
CREATE TABLE function_parameter
(
    id BIGINT PRIMARY KEY,

    function_id BIGINT,

    param_name VARCHAR(64),

    param_type VARCHAR(64)
);
```

---

## formula\_compile\_log表

```Plain Text
CREATE TABLE formula_compile_log
(
    id BIGINT PRIMARY KEY,

    formula_hash VARCHAR(64),

    success TINYINT,

    compile_time_ms INT,

    created_at DATETIME
);
```

---

# 权限设计

权限：

```Plain Text
formula:validate

formula:compile

formula:function:list
```

---

普通用户：

```Plain Text
validate

compile
```

---

管理员：

```Plain Text
function:register
```

---

# 部署架构

```Plain Text
API Gateway

↓

Formula Engine

↓

Redis

↓

PostgreSQL

↓

Kafka
```

---

# 高可用

Formula Engine：

```Plain Text
5 Pods
```

原因：

CPU密集型。

---

Redis：

```Plain Text
Cluster
```

---

Kafka：

```Plain Text
3 Broker
```

---

# 性能指标

Validate：

```Plain Text
P95 < 10ms
```

---

Compile：

```Plain Text
P95 < 30ms
```

---

AST生成：

```Plain Text
P95 < 20ms
```

---

QPS：

```Plain Text
Compile
5000 QPS
```

---

# 可观测性

Metrics：

```Plain Text
formula_compile_total

formula_compile_fail_total

formula_validate_total

formula_cache_hit_total
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

# MVP范围

必须实现：

✅ Lexer

✅ Parser

✅ AST

✅ Validator

✅ Function Registry

✅ Execution Plan

✅ Compile API

✅ Redis Cache

---

暂缓：

❌ JIT编译

❌ LLVM

❌ WASM执行

❌ 分布式编译

---

# 后续演进

V2 引入 IR（Intermediate Representation）层：

```Plain Text
DSL → AST → IR → Execution Plan → Backtest Engine
```

IR 层的引入使 Formula Engine 可兼容通达信公式、同花顺公式等外部 DSL，统一编译到同一执行计划。

V3 支持 JIT 编译与 WASM 执行，进一步提升回测性能。


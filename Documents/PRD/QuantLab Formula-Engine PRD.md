# QuantLab Formula\-Engine PRD

Version：1\.0

Module：Formula Engine

Priority：P0

Status：Draft

Owner：Product Team

---

# 模块定位

Formula Engine 是 QuantLab 的核心能力模块。

负责：

- 策略表达

- 公式解析

- 语法校验

- 指标计算

- 条件判断

- 股票筛选

不负责：

- 数据存储

- 回测执行

- 排行榜计算

---

# 产品目标

允许用户通过简单公式表达投资逻辑。

实现：

投资想法

↓

策略表达

↓

公式解析

↓

可执行策略

---

# 设计原则

## 原则1

兼容通达信思维

用户无需重新学习。

---

## 原则2

支持未来AI生成

公式必须结构化。

---

## 原则3

支持跨市场

未来支持：

A股

港股

美股

ETF

基金

指数

---

## 原则4

支持非程序员

避免复杂语法。

---

# DSL定位

DSL：

Domain Specific Language

策略专用语言

---

用户最终编写：

```Plain Text
ROE > 15
AND PE < 20
AND RevenueGrowth > 30
```

而不是：

```Python
if roe > 15:
    ...
```

---

# Formula Engine架构

Formula Engine

├── 编辑器

├── 语法检查器

├── 自动补全器

├── 解析器\(Parser\)

├── AST生成器

├── 执行器\(Executor\)

├── 函数库

└── 指标库

---

# 用户场景

## 场景1

选股

用户输入：

```Plain Text
ROE > 15
AND PE < 20
```

输出：

符合条件股票池

---

## 场景2

买入条件

```Plain Text
CROSS(MA(CLOSE,5),MA(CLOSE,20))
```

输出：

买入信号

---

## 场景3

卖出条件

```Plain Text
CLOSE < MA(CLOSE,20)
```

输出：

卖出信号

---

# 编辑器功能

## 7\.1 语法高亮

关键词：

AND

OR

NOT

IF

---

函数：

MA

EMA

RSI

MACD

---

变量：

OPEN

HIGH

LOW

CLOSE

VOL

---

# 7\.2 自动补全

用户输入：

```Plain Text
MA(
```

提示：

```Plain Text
MA(CLOSE,5)

MA(CLOSE,10)

MA(CLOSE,20)
```

---

# 7\.3 参数提示

输入：

```Plain Text
RSI(
```

显示：

```Plain Text
RSI(price, period)
```

---

# 7\.4 实时校验

示例：

```Plain Text
MA(CLOSE,5
```

提示：

缺少右括号

---

# 7\.5 函数文档

点击函数：

展示：

定义

参数

返回值

示例

---

# 数据类型

## Number

```Plain Text
15

20.5
```

---

## Boolean

```Plain Text
TRUE

FALSE
```

---

## Series

时间序列

例如：

```Plain Text
CLOSE
```

---

# 运算符

## 算术

```Plain Text
+
-
*
/
%
```

---

## 比较

```Plain Text
>
<
>=
<=
==
!=
```

---

## 逻辑

```Plain Text
AND

OR

NOT
```

---

# 内置变量

## 行情变量

```Plain Text
OPEN

HIGH

LOW

CLOSE

VOL

AMOUNT
```

---

## 财务变量

```Plain Text
PE

PB

PS

ROE

ROA

EPS
```

---

## 成长变量

```Plain Text
RevenueGrowth

ProfitGrowth
```

---

## 市值变量

```Plain Text
MarketCap

FloatMarketCap
```

---

# 内置函数体系

================================================

基础函数

================================================

MA\(\)

EMA\(\)

SMA\(\)

STD\(\)

ABS\(\)

MAX\(\)

MIN\(\)

---

================================================

统计函数

================================================

SUM\(\)

AVG\(\)

COUNT\(\)

---

================================================

时间序列函数

================================================

REF\(\)

BARSLAST\(\)

HHV\(\)

LLV\(\)

---

示例：

```Plain Text
REF(CLOSE,1)
```

昨日收盘价

---

================================================

信号函数

================================================

CROSS\(\)

LONGCROSS\(\)

FILTER\(\)

---

示例：

```Plain Text
CROSS(MA(CLOSE,5),MA(CLOSE,20))
```

---

================================================

趋势函数

================================================

MACD\(\)

RSI\(\)

KDJ\(\)

BOLL\(\)

ATR\(\)

---

# 公式分类

## 选股公式

返回：

Boolean

---

示例：

```Plain Text
ROE > 15
```

---

## 排序公式

返回：

Number

---

示例：

```Plain Text
ROE * ProfitGrowth
```

---

## 买入公式

返回：

Signal

---

## 卖出公式

返回：

Signal

---

# AST生成

用户公式：

```Plain Text
PE < 20
AND ROE > 15
```

转换：

```JSON
{
  "type":"AND",
  "children":[
    {
      "type":"COMPARE",
      "field":"PE",
      "operator":"<",
      "value":20
    },
    {
      "type":"COMPARE",
      "field":"ROE",
      "operator":">",
      "value":15
    }
  ]
}
```

---

# 错误体系

## 语法错误

括号缺失

参数缺失

关键字错误

---

## 引用错误

未知变量

未知函数

---

## 类型错误

Number与Boolean比较

Series与String比较

---

# 公式测试

支持：

测试公式

---

输入：

股票代码

时间范围

---

输出：

公式结果

指标值

信号结果

---

# 公式调试器

支持：

逐日执行

---

显示：

日期

变量值

函数结果

最终结果

---

用途：

定位策略问题

---

# AI辅助编写

输入：

寻找低估值高成长股票

---

生成：

```Plain Text
PE < 20
AND ROE > 15
AND RevenueGrowth > 20
```

---

# 策略模板

预置：

双均线策略

MACD策略

RSI策略

高ROE策略

成长策略

ETF轮动

---

# 权限控制

免费用户：

基础函数

---

Pro 及以上：

高级因子

高级函数

自定义函数

---

# 性能要求

公式校验

\<100ms

---

公式执行

1000只股票

\<2秒

---

# MVP范围

必须实现：

√ 编辑器

√ 语法高亮

√ 自动补全

√ 基础变量

√ 基础函数

√ AST生成

√ 公式测试

---

暂缓：

× 用户自定义函数

× 自定义指标

× 多语言DSL

× 图形化策略编辑器

---

# KPI

公式创建数

公式执行次数

公式保存数

公式成功率

策略生成率

---

# 核心原则

原则1

所有策略最终转换为AST

---

原则2

所有执行逻辑基于AST

---

原则3

Formula Engine不依赖Backtest Engine

---

原则4

DSL必须向后兼容

---

原则5

AI生成策略必须符合DSL规范


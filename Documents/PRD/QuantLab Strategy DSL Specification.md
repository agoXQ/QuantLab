# QuantLab Strategy DSL Specification

Version：1\.0

Status：Draft

Type：Language Specification

Owner：Architecture Team

Last Update：2026

---

# 文档目标

定义 QuantLab 策略语言（Strategy DSL）。

本规范用于统一：

- Formula Engine

- Strategy Service

- Backtest Engine

- AI Service

之间的数据表达方式。

---

# 设计目标

DSL必须满足：

## 易学

兼容通达信使用习惯

---

## 可扩展

支持：

A股

港股

美股

ETF

基金

指数

---

## 可执行

最终转换为AST

统一执行

---

## 可生成

支持AI生成

---

# DSL层级

DSL分为四层：

L1

表达式层

---

L2

条件层

---

L3

规则层

---

L4

策略层

---

# 基础语法

## 标识符

格式：

```Plain Text
[A-Za-z_][A-Za-z0-9_]*
```

示例：

```Plain Text
ROE

ProfitGrowth

MA
```

---

## 数字

整数：

```Plain Text
15
```

浮点：

```Plain Text
15.5
```

---

## 字符串

```Plain Text
"银行"

"新能源"
```

---

## 布尔值

```Plain Text
TRUE

FALSE
```

---

# 数据类型系统

## Number

单一数值

示例：

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

示例：

```Plain Text
CLOSE

OPEN
```

---

## Signal

交易信号

示例：

```Plain Text
BUY

SELL
```

---

## StockSet

股票集合

用于选股结果

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

# 运算优先级

最高：

```Plain Text
()
```

↓

```Plain Text
*
/
%
```

↓

```Plain Text
+
-
```

↓

```Plain Text
>
<
>=
<=
```

↓

```Plain Text
==
!=
```

↓

```Plain Text
AND
```

↓

最低：

```Plain Text
OR
```

---

# 变量体系

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

# 函数规范

统一格式：

```Plain Text
FUNCTION(arg1,arg2,...)
```

---

函数名：

大小写不敏感

以下等价：

```Plain Text
MA()

ma()

Ma()
```

---

# 基础函数

## MA

移动平均

```Plain Text
MA(series, period)
```

示例：

```Plain Text
MA(CLOSE,20)
```

---

## EMA

指数移动平均

```Plain Text
EMA(series, period)
```

---

## SMA

平滑移动平均

```Plain Text
SMA(series, period)
```

---

# 引用函数

## REF

引用历史数据

```Plain Text
REF(series,n)
```

示例：

```Plain Text
REF(CLOSE,1)
```

昨日收盘价

---

# 区间函数

## HHV

最高值

```Plain Text
HHV(series,n)
```

---

## LLV

最低值

```Plain Text
LLV(series,n)
```

---

# 统计函数

## COUNT

```Plain Text
COUNT(condition,n)
```

---

## SUM

```Plain Text
SUM(series,n)
```

---

## AVG

```Plain Text
AVG(series,n)
```

---

# 信号函数

## CROSS

上穿

```Plain Text
CROSS(a,b)
```

---

定义：

```Plain Text
a(t-1) <= b(t-1)

AND

a(t) > b(t)
```

---

## LONGCROSS

持续上穿

```Plain Text
LONGCROSS(a,b,n)
```

---

# 技术指标函数

## RSI

```Plain Text
RSI(period)
```

---

## MACD

```Plain Text
MACD(fast,slow,signal)
```

---

## KDJ

```Plain Text
KDJ(n,m1,m2)
```

---

## BOLL

```Plain Text
BOLL(n,k)
```

---

# 表达式规范

合法：

```Plain Text
PE < 20
```

---

合法：

```Plain Text
MA(CLOSE,5) > MA(CLOSE,20)
```

---

合法：

```Plain Text
ROE > 15
AND PE < 20
```

---

非法：

```Plain Text
PE < AND 20
```

---

# 规则定义

## 选股规则

返回：

Boolean

---

示例：

```Plain Text
ROE > 15
AND PE < 20
```

---

## 买入规则

返回：

Signal

---

示例：

```Plain Text
CROSS(MA(CLOSE,5),MA(CLOSE,20))
```

---

## 卖出规则

返回：

Signal

---

示例：

```Plain Text
CLOSE < MA(CLOSE,20)
```

---

## 排序规则

返回：

Number

---

示例：

```Plain Text
ROE * ProfitGrowth
```

---

# 策略对象标准

统一JSON格式：

\{

"stock\_select": "\.\.\.",

"buy\_rule": "\.\.\.",

"sell\_rule": "\.\.\.",

"ranking\_rule": "\.\.\.",

"risk\_rule": \{\}

\}

---

# AST标准

所有DSL最终转换为AST。

示例：

PE \< 20

转换：

\{

"type": "CompareExpr",

"operator": "\<",

"left": \{

"type": "Identifier",

"value": "PE"

\},

"right": \{

"type": "Literal",

"value": 20

\}

\}

---

# 错误码规范

1000

语法错误

---

1001

未知变量

---

1002

未知函数

---

1003

参数数量错误

---

1004

类型错误

---

1005

除零错误

---

# 扩展机制

V1支持：

内置函数

---

V2支持：

用户自定义函数

---

V3支持：

用户自定义指标

---

V4支持：

插件式指标库

---

# 向后兼容原则

V2必须兼容V1

V3必须兼容V2

任何已保存策略不得失效

---

# AI兼容规范

AI生成DSL时：

必须输出合法DSL

禁止输出Python

禁止输出SQL

禁止输出自然语言

仅允许DSL表达式

---

# DSL核心原则

原则1

所有策略最终AST化

原则2

所有执行AST驱动

原则3

DSL不依赖具体市场

原则4

DSL不依赖具体数据库

原则5

DSL是平台统一语言


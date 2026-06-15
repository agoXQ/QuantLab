# QuantLab AI Service PRD

Version：1\.0

Module：AI Service

Priority：P0

Status：Draft

Owner：Product Team

Last Update：2026

---

# 模块定位

AI Service 是 QuantLab 的智能研究中心。

负责：

- 策略生成 

- 策略解释 

- 策略优化 

- 回测分析 

- 因子发现 

- 策略推荐 

- 智能问答 

不负责：

- 行情存储 

- 回测执行 

- 用户认证 

---

# 产品愿景

让用户从：

```Plain Text
手工研究策略
```

变成：

```Plain Text
AI辅助研究策略
```

最终实现：

```Plain Text
人机协同量化研究
```

---

# 产品目标

降低量化门槛。

帮助用户：

- 不懂编程也能创建策略 

- 不懂量化也能理解策略 

- 快速发现有效策略 

---

# AI Service架构

```Plain Text
AI Service

├── AI Chat
├── Strategy Generator
├── Formula Copilot
├── Backtest Analyst
├── Strategy Optimizer
├── Strategy Recommender
├── Knowledge Base
├── RAG Engine
└── LLM Gateway
```

---

# AI Chat

统一聊天入口。

---

用户可以询问：

```Plain Text
什么是ROE？

什么是双均线策略？

为什么我的策略亏损？
```

---

# Strategy Generator

最核心能力。

---

输入：

```Plain Text
寻找高成长低估值股票
```

---

AI输出：

```Plain Text
ROE > 15
AND PE < 20
AND RevenueGrowth > 30
```

---

同时生成：

策略说明

适用市场

风险提示

---

# 自然语言生成DSL

输入：

```Plain Text
5日均线上穿20日均线买入
跌破20日均线卖出
```

---

输出：

```Plain Text
BUY:

CROSS(
 MA(CLOSE,5),
 MA(CLOSE,20)
)

SELL:

CLOSE < MA(CLOSE,20)
```

---

# Strategy Wizard

策略向导。

---

AI主动提问：

```Plain Text
偏好什么市场？

A股？

ETF？

港股？
```

---

逐步生成策略。

---

# Formula Copilot

公式助手。

---

支持：

自动补全

错误修复

语法解释

函数推荐

---

示例：

```Plain Text
用户：

PE < 20 AND
```

---

AI提示：

缺少后续条件。

---

# Formula Explain

公式解释。

---

输入：

```Plain Text
ROE > 15
AND PE < 20
```

---

输出：

```Plain Text
选择盈利能力较强且估值较低的公司。
```

---

# Strategy Explain

解释完整策略。

---

内容：

选股逻辑

买入逻辑

卖出逻辑

风控逻辑

适用行情

---

# Backtest Analyst

分析回测结果。

---

输入：

回测结果

---

输出：

收益来源

风险来源

关键交易

失效阶段

---

# 回测诊断

示例：

```Plain Text
收益率：

12%
```

---

AI分析：

```Plain Text
收益主要来源于：

新能源板块

2024年上涨阶段
```

---

# 策略优化器

输入：

策略

- 

回测结果

---

输出：

优化建议

---

示例：

```Plain Text
当前：

MA(5)
```

---

建议：

```Plain Text
MA(10)
```

---

# 参数优化助手

自动分析：

参数敏感性

---

输出：

最佳参数区间

---

避免：

参数过拟合

---

# 因子发现助手

AI发现：

潜在有效因子

---

示例：

```Plain Text
ROE

+
ProfitGrowth

+
Momentum
```

---

形成新策略。

---

# 策略推荐系统

依据：

用户画像

浏览记录

策略偏好

---

推荐：

类似策略

热门策略

潜力策略

---

# AI排行榜分析

每日分析：

收益榜

胜率榜

热门榜

---

生成：

榜单解读

---

# 社区内容生成

自动生成：

策略摘要

策略点评

周报

月报

---

# AI研究报告

输入：

策略

---

输出：

完整研究报告

---

内容：

策略逻辑

收益分析

风险分析

市场适应性

---

# AI知识库

包含：

DSL文档

量化知识

指标知识

行业知识

投资理论

---

# RAG系统

检索来源：

策略库

回测库

排行榜

知识库

社区内容

---

# AI学习对象

未来可学习：

优秀策略

优秀作者

排行榜策略

---

形成：

策略知识图谱

---

# AI策略评分

分析：

逻辑质量

风险水平

复杂度

可解释性

---

输出：

AI评分

---

# AI策略审查

检查：

未来函数

幸存者偏差

数据泄露

参数过拟合

---

输出：

风险提示

---

# AI风险提示

示例：

```Plain Text
策略收益率较高

但交易次数过少

存在统计偏差风险
```

---

# AI市场分析

用户提问：

```Plain Text
当前市场适合什么策略？
```

---

AI分析：

市场状态

风格轮动

行业机会

---

# AI选股助手

用户提问：

```Plain Text
帮我找符合策略的股票
```

---

AI返回：

股票列表

原因分析

---

# AI观察名单

自动发现：

符合策略股票

---

加入观察池。

---

# AI工作流

未来支持：

```Plain Text
创建策略

↓

执行回测

↓

分析结果

↓

优化参数

↓

再次回测

↓

生成报告
```

自动完成。

---

# Prompt管理中心

统一管理：

Prompt模板

Prompt版本

Prompt测试

---

# 模型管理

支持：

OpenAI

Claude

Gemini

DeepSeek

Qwen

私有模型

---

# LLM Gateway

统一调用层。

---

负责：

路由

限流

缓存

日志

---

# AI权限体系

Free：

每日20次

---

Pro：

每日200次

---

Master：

无限使用

---

# Token管理

统计：

消耗量

成本

用户额度

---

# AI缓存

缓存：

策略解释

回测分析

排行榜解读

---

降低成本。

---

# AI反馈系统

用户可评价：

有帮助

无帮助

---

用于Prompt优化。

---

# MVP范围

必须实现：

√ AI Chat

√ Strategy Generator

√ DSL生成

√ Formula Explain

√ Backtest Analyst

√ 策略优化建议

---

暂缓：

× AI自动回测

× AI自动调参

× 因子挖掘

× AutoML

× AI交易代理

---

# KPI

AI使用人数

AI调用次数

策略生成次数

AI生成策略回测率

AI生成策略发布率

用户满意度

---

# 核心原则

原则1

AI辅助决策

不替代用户决策

---

原则2

所有AI结果可追溯

---

原则3

AI生成策略必须符合DSL规范

---

原则4

AI分析必须引用真实数据

---

原则5

AI服务于量化研究流程

---

# QuantLab AI终极形态

V1：

AI助手

↓

V2：

AI研究员

↓

V3：

AI量化分析师

↓

V4：

AI策略团队

↓

V5：

Autonomous Quant Lab

---

# AI Service 在整体架构中的位置

最终架构会变成：

```Plain Text
QuantLab

├── User Service
├── Strategy Service
├── Formula Engine
├── Market Data Service
├── Backtest Engine
├── Ranking Service
├── Community Service
├── AI Service
└── Notification Service
```

# AI Service 子系统划分

## AI Copilot（交互层）

定位：高频交互入口，面向所有用户。

能力：

- DSL生成

- 公式解释

- 策略问答

- 编辑器辅助

---

## AI Research（研究层）

定位：高价值研究能力，面向付费用户。

能力：

- 回测分析

- 参数优化

- 策略评分

- 榜单分析

- 研究报告生成

---

## 商业化策略

AI Copilot：所有用户可用（含免费版）。

AI Research：Master 及以上会员可用。


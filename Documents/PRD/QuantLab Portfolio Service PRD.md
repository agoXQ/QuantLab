# QuantLab Portfolio Service PRD

Version: 1\.0

Module: Portfolio Service

Priority: P0

Status: Draft

Owner: Product Team

Last Update: 2026-06

Dependencies:

- Strategy Service

- Backtest Engine

- Market Data Service

- Community Service

- Ranking Service

---

# 产品定位

Portfolio Service 是 QuantLab 的：

```Plain Text
Strategy Allocation Platform
```

负责：

```Plain Text
策略组合管理

组合回测

组合调仓

组合排行榜

组合分享

组合研究
```

Portfolio 是连接：

```Plain Text
Strategy
↓
Backtest
↓
Ranking
↓
Community
```

的核心桥梁。

---

# 为什么需要 Portfolio

当前：

```Plain Text
策略A

ROE > 15
PE < 20
```

输出：

```Plain Text
股票池
```

但真实投资中：

用户不会只运行一个策略。

例如：

```Plain Text
价值策略 40%

成长策略 30%

红利策略 20%

现金策略 10%
```

最终形成：

```Plain Text
Portfolio
```

---

# 产品目标

V1目标：

支持用户创建：

```Plain Text
策略组合
```

并完成：

```Plain Text
组合回测

组合分析

组合分享
```

---

V2目标：

支持：

```Plain Text
自动调仓

组合排行榜

组合订阅
```

---

V3目标：

支持：

```Plain Text
组合跟投

实盘同步

策略市场
```

---

# 核心能力

Portfolio Service 提供：

```Plain Text
Portfolio CRUD

Portfolio Allocation

Portfolio Backtest

Portfolio Analytics

Portfolio Publish

Portfolio Ranking
```

---

# 用户故事

## 用户故事1

作为量化用户

我希望：

```Plain Text
将多个策略组成一个组合
```

从而：

```Plain Text
获得更稳定收益
```

---

## 用户故事2

作为投资者

我希望：

```Plain Text
查看优秀组合
```

从而：

```Plain Text
学习资产配置
```

---

## 用户故事3

作为作者

我希望：

```Plain Text
发布我的组合
```

从而：

```Plain Text
获得关注
```

---

# 领域模型

---

## Portfolio

组合

---

属性：

```Plain Text
ID

Owner

Name

Description

Visibility

Status
```

---

示例：

```Plain Text
价值成长组合
```

---

## Portfolio Item

组合项

---

属性：

```Plain Text
PortfolioID

StrategyID

Weight
```

---

示例：

```Plain Text
价值策略 40%

成长策略 30%

红利策略 30%
```

---

# 组合类型

V1支持：

---

Strategy Portfolio

```Plain Text
策略组合
```

---

未来：

---

Stock Portfolio

```Plain Text
股票组合
```

---

ETF Portfolio

```Plain Text
ETF组合
```

---

Mixed Portfolio

```Plain Text
混合组合
```

---

# 创建组合

用户输入：

```Plain Text
组合名称

描述

策略列表

权重
```

---

示例：

```Plain Text
价值成长组合
```

包含：

```Plain Text
价值策略 40%

成长策略 40%

红利策略 20%
```

---

# 权重规则

总和：

```Plain Text
100%
```

---

校验：

```Plain Text
必须=100
```

---

支持：

```Plain Text
自动归一化
```

---

# 权重模式

---

Manual

手动权重

---

Equal Weight

等权重

---

Risk Weight

风险权重

---

未来：

```Plain Text
AI Weight
```

---

# 组合回测

输入：

```Plain Text
Portfolio
```

---

执行：

```Plain Text
策略回测

收益聚合

组合收益曲线
```

---

输出：

```Plain Text
Portfolio Backtest Report
```

---

# 组合收益计算

例如：

```Plain Text
策略A收益

20%

权重40%
```

---

贡献：

```Plain Text
8%
```

---

最终：

```Plain Text
Portfolio Return

=
Σ(strategy return × weight)
```

---

# 调仓模型

V1：

```Plain Text
固定权重
```

---

V2：

```Plain Text
周期调仓
```

---

支持：

```Plain Text
月度

季度

年度
```

---

# 调仓方式

---

Rebalance

再平衡

---

Threshold

阈值调仓

---

Signal

信号调仓

---

# Portfolio Analytics

组合分析模块

---

指标：

```Plain Text
累计收益

年化收益

最大回撤

夏普比率

波动率
```

---

新增：

```Plain Text
组合贡献分析
```

---

# 贡献分析

展示：

```Plain Text
哪个策略贡献收益最多
```

---

例如：

```Plain Text
成长策略

贡献：

45%
```

---

# 风险分析

分析：

```Plain Text
策略相关性

风险集中度
```

---

展示：

```Plain Text
相关性矩阵
```

---

# Portfolio Publish

支持：

```Plain Text
公开

私有

仅链接访问
```

---

# Community集成

发布后：

自动生成：

```Plain Text
Content
```

---

进入：

```Plain Text
Feed
```

---

# Ranking集成

组合自动进入：

```Plain Text
Portfolio Ranking
```

---

支持：

```Plain Text
收益榜

夏普榜

综合榜
```

---

# Portfolio详情页

展示：

```Plain Text
组合简介

策略构成

权重

收益曲线

风险指标

评论
```

---

# 收藏组合

支持：

```Plain Text
收藏
```

---

用于：

```Plain Text
推荐系统
```

---

# 关注组合

V2支持：

```Plain Text
Follow Portfolio
```

---

# 组合复制

V2支持：

```Plain Text
Clone Portfolio
```

---

用户：

```Plain Text
复制组合

修改权重
```

---

# Portfolio模板

官方提供：

```Plain Text
价值模板

成长模板

红利模板

ETF模板
```

---

# AI组合（V2）

用户输入：

```Plain Text
稳健型
```

---

AI生成：

```Plain Text
组合配置
```

---

# 订阅机制（V3）

支持：

```Plain Text
组合订阅
```

---

作者：

```Plain Text
收费
```

---

用户：

```Plain Text
订阅
```

---

# 跟投机制（V3）

支持：

```Plain Text
Mirror Portfolio
```

---

自动同步：

```Plain Text
调仓动作
```

---

# Portfolio事件

发布：

```Plain Text
PortfolioCreated
```

---

更新：

```Plain Text
PortfolioUpdated
```

---

删除：

```Plain Text
PortfolioDeleted
```

---

公开：

```Plain Text
PortfolioPublished
```

---

回测完成：

```Plain Text
PortfolioBacktestCompleted
```

---

# API能力

---

创建组合

```Plain Text
POST /portfolios
```

---

更新组合

```Plain Text
PUT /portfolios/{id}
```

---

获取组合

```Plain Text
GET /portfolios/{id}
```

---

回测组合

```Plain Text
POST /portfolios/{id}/backtests
```

---

发布组合

```Plain Text
POST /portfolios/{id}/publish
```

---

获取分析

```Plain Text
GET /portfolios/{id}/analytics
```

---

# 权限模型

Free：

```Plain Text
最多3个组合
```

---

Pro：

```Plain Text
最多50个组合
```

---

Master：

```Plain Text
无限组合
```

---

# MVP范围

必须实现：

✅ Portfolio

✅ Portfolio Item

✅ 权重配置

✅ Portfolio Backtest

✅ Analytics

✅ Publish

✅ Community集成

---

暂缓：

❌ AI组合

❌ 组合订阅

❌ 跟投

❌ 实盘同步

---

# 成功指标

用户维度：

```Plain Text
创建组合用户占比 > 30%
```

---

内容维度：

```Plain Text
Portfolio内容占比 > 20%
```

---

社区维度：

```Plain Text
Portfolio互动率 > Strategy
```




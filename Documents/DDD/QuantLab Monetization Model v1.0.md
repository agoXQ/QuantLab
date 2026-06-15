# QuantLab Monetization Model v1\.0

Version: 1\.0

Status: Business Architecture Baseline

Owner: Product Team

Related Documents:

- QuantLab Platform PRD

- Billing Service PRD

- QuantLab Domain Model v1\.0

- Portfolio Service PRD

- Strategy Service PRD

---

# 文档目标

定义 QuantLab 的商业模式。

回答以下问题：

- 平台如何赚钱？

- 用户为什么愿意付费？

- 创作者为什么愿意入驻？

- 平台如何与创作者分成？

- 收费能力如何与产品能力绑定？

- 商业化路线如何逐步演进？

本文件是：

Billing Service

Subscription Service

Creator Economy

Marketplace

的商业基础。

---

# 商业模式设计原则

## 原则一

工具收费

用户为能力付费。

例如：

- AI能力

- 高级回测

- 私有策略

---

## 原则二

内容收费

用户为优质策略付费。

例如：

- 价值策略

- 红利策略

- 行业策略

---

## 原则三

结果收费

用户为研究成果付费。

例如：

- 优质组合

- 研究报告

- 因子库

---

## 原则四

平台抽成

平台不直接创造内容。

平台创造市场。

---

# 收入结构设计

目标收入结构：

```Plain Text
SaaS Subscription      40%
Strategy Marketplace   30%
Portfolio Marketplace  20%
Research Marketplace   10%
```

---

长期目标：

```Plain Text
Creator Economy 收入
超过
SaaS收入
```

---

# 第一阶段商业模式（MVP）

阶段：

0\~1000用户

---

收入来源：

仅保留：

- Pro会员

- Master会员

---

暂不开放：

- 策略订阅

- 组合订阅

- 创作者收益

---

原因：

先验证工具价值。

---

# 第二阶段商业模式

阶段：

1000\~10000用户

---

新增：

Strategy Subscription

---

用户：

订阅策略

---

作者：

获得收益

---

平台：

获得抽成

---

# 第三阶段商业模式

阶段：

10000\~50000用户

---

新增：

Portfolio Subscription

---

用户：

订阅组合

---

平台：

获得组合市场收入

---

# 第四阶段商业模式

阶段：

50000\+用户

---

新增：

Research Marketplace

---

支持：

研究报告

行业专题

策略研究

量化课程

---

# SaaS会员体系

## Free

价格：

0元

---

能力：

基础策略

基础回测

社区浏览

排行榜浏览

---

限制：

策略数量限制

回测次数限制

无AI能力

---

## Pro

建议价格：

9\.9 USD/月

或

68 RMB/月

---

能力：

高级回测

AI策略解释

AI回测分析

私有策略

私有组合

高级排行榜

---

目标用户：

个人投资者

---

## Master

建议价格：

29\.9 USD/月

或

198 RMB/月

---

能力：

全部Pro能力

API Access

高级AI

创作者认证

收费策略发布

收费组合发布

---

目标用户：

专业投资者

量化研究员

KOL

---

## Enterprise

建议价格：

按合同

---

能力：

团队空间

专属资源

私有部署

SSO

企业管理

---

# 创作者经济体系

创作者等级：

LEVEL\_1

LEVEL\_2

LEVEL\_3

VERIFIED

ELITE

---

升级依据：

粉丝数

订阅数

收入

活跃度

---

# 策略市场

Strategy Marketplace

---

作者发布：

收费策略

---

收费模式：

一次性购买

月订阅

年订阅

---

价格建议：

基础策略：

3\~10 USD/月

---

专业策略：

10\~30 USD/月

---

高级策略：

30\~100 USD/月

---

# 组合市场

Portfolio Marketplace

---

收费模式：

月订阅

年订阅

---

价格建议：

10\~50 USD/月

---

高端组合：

100 USD/月以上

---

# 研究报告市场

支持：

行业研究

量化报告

策略研究

专题分析

---

收费模式：

单次购买

会员免费

高级会员专享

---

# 平台抽成模型

默认：

```Plain Text
创作者 80%
平台    20%
```

---

# 阶梯抽成模型

普通作者：

```Plain Text
作者 80%
平台 20%
```

---

认证作者：

```Plain Text
作者 85%
平台 15%
```

---

Elite作者：

```Plain Text
作者 90%
平台 10%
```

---

# 推荐奖励体系

用户邀请用户：

奖励：

- 免费会员天数

- 优惠券

---

创作者邀请创作者：

奖励：

收入分成奖励

---

# 续费模型

自动续费：

默认开启

---

续费提醒：

到期前：

7天

3天

1天

---

# 增值服务

未来支持：

---

高级AI额度

---

额外回测额度

---

高级数据包

---

实时行情

---

机构数据

---

因子数据

---

# API商业化

Master以上开放。

---

计费方式：

按调用量

---

示例：

10000次/月

免费

---

超过：

按量收费

---

# 企业版商业化

收费模式：

年度合同

---

包含：

用户数授权

API授权

团队空间

专属支持

私有部署

---

# Creator Economy KPI

目标：

---

创作者数量

---

付费策略数量

---

付费组合数量

---

订阅用户数量

---

创作者收入

---

平台抽成收入

---

# SaaS KPI

核心指标：

MRR

Monthly Recurring Revenue

---

ARR

Annual Recurring Revenue

---

付费率

---

续费率

---

流失率

---

ARPU

---

# Marketplace KPI

核心指标：

GMV

---

订阅总金额

---

平台抽成金额

---

创作者总收入

---

平均客单价

---

# 平台收入演进

阶段一：

```Plain Text
90% 会员收入
10% 其他收入
```

---

阶段二：

```Plain Text
60% 会员收入
40% 策略市场收入
```

---

阶段三：

```Plain Text
40% 会员收入
60% Marketplace收入
```

---

长期目标：

```Plain Text
20% SaaS
80% Creator Economy
```

---

# 为什么这个模型适合QuantLab

因为 QuantLab 的核心不是：

量化工具

而是：

量化知识

量化策略

量化研究成果

---

最终形成：

```Plain Text
投资者
   ↓

购买策略
   ↓

购买组合
   ↓

购买研究
   ↓

订阅作者
```

同时：

```Plain Text
策略作者
   ↓

发布内容
   ↓

获得收入
```

平台：

```Plain Text
提供基础设施
+
流量
+
交易市场
```

---

# 最终商业飞轮

Creator

↓

发布策略

↓

产生收益

↓

吸引更多Creator

↓

产生更多优质内容

↓

吸引更多Investor

↓

产生更多订阅

↓

平台获得更多收入

↓

投入更多资源

↓

强化平台生态

形成：

QuantLab Marketplace


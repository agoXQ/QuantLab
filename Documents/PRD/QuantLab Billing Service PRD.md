# QuantLab Billing Service PRD

Version: 1\.0

Module: Billing Service

Priority: P1（商业化核心服务）

Status: Draft

Owner: Platform Product Team

Dependencies:

- User Service

- Strategy Service

- Portfolio Service

- Community Service

- Notification Service

- Ranking Service

- AI Service

---

# 产品定位

Billing Service 是 QuantLab 的商业化中枢。

负责：

- 会员管理

- 策略订阅

- 组合订阅

- 创作者收益

- 平台抽成

- 订单管理

- 支付管理

- 退款管理

- 结算管理

不负责：

- 用户管理

- 策略管理

- 组合管理

- 通知发送

Billing Service 本质定位：

SaaS Subscription Platform

- 

Creator Economy Platform

---

# 产品目标

V1：

实现会员订阅体系

V2：

实现策略订阅市场

V3：

实现创作者经济生态

---

# 用户角色

## 普通用户（Investor）

能力：

- 购买会员

- 订阅策略

- 订阅组合

- 查看订单

---

## 策略作者（Creator）

能力：

- 发布收费策略

- 发布收费组合

- 查看收入

- 提现

---

## 平台管理员（Admin）

能力：

- 管理订单

- 管理退款

- 管理结算

- 管理抽成规则

---

# 核心业务模型

Billing Domain 包含：

Membership

Subscription

Order

Payment

Revenue Share

Settlement

Coupon

Invoice

---

# 商业模式

## SaaS模式

会员等级：

FREE

PRO

MASTER

ENTERPRISE

---

PRO会员

开放：

- AI策略生成

- AI组合生成

- 高级回测

- 私有策略

按月付费

按年付费

---

MASTER会员

开放：

- API Access

- 私有组合

- 高级AI

- 创作者能力

---

ENTERPRISE

开放：

- 团队协作

- 企业空间

- 私有部署

---

# Creator Economy模式

策略作者可发布：

免费策略

收费策略

---

收费模式：

一次性购买

月度订阅

年度订阅

---

组合作者可发布：

免费组合

收费组合

---

研究员可发布：

研究报告

行业报告

专题分析

---

# 平台抽成模型

默认：

创作者收益：80%

平台收益：20%

---

后续支持：

等级抽成

认证作者抽成

活动抽成

---

# 功能模块

Billing Service 包含：

Membership Module

Subscription Module

Order Module

Payment Module

Revenue Module

Settlement Module

Coupon Module

Invoice Module

---

# Membership Module

职责：

管理会员生命周期

---

功能：

购买会员

升级会员

续费会员

取消自动续费

会员到期

会员恢复

---

支持：

月付

季付

年付

---

会员状态：

ACTIVE

EXPIRED

CANCELLED

GRACE\_PERIOD

---

# Subscription Module

职责：

管理内容订阅

---

支持订阅对象：

Strategy

Portfolio

Creator

Research Report

---

订阅周期：

MONTHLY

QUARTERLY

YEARLY

LIFETIME

---

订阅状态：

ACTIVE

EXPIRED

CANCELLED

PAUSED

---

# Order Module

职责：

统一订单管理

---

订单来源：

Membership

Strategy Subscription

Portfolio Subscription

Report Subscription

Recharge

---

订单状态：

PENDING

PAID

FAILED

CANCELLED

REFUNDED

---

# Payment Module

职责：

支付管理

---

支付渠道：

Stripe

PayPal

Alipay

WeChat Pay

PayPay

Bank Transfer

---

支付状态：

INIT

PROCESSING

SUCCESS

FAILED

REFUNDED

---

# Revenue Module

职责：

收益分账

---

计算：

订单支付成功

↓

计算平台抽成

↓

计算作者收益

↓

生成分账记录

---

收益状态：

PENDING

CONFIRMED

SETTLED

---

# Settlement Module

职责：

作者提现

---

流程：

作者申请提现

↓

风控审核

↓

财务打款

↓

提现完成

---

状态：

PENDING

APPROVED

REJECTED

PAID

---

# Coupon Module

支持：

折扣券

活动券

邀请码

会员兑换码

---

优惠类型：

FIXED

PERCENTAGE

---

# Invoice Module

支持：

电子发票

企业发票

会员发票

订阅发票

---

# 订单生命周期

创建订单

↓

待支付

↓

支付成功

↓

生成订阅

↓

生成收益

↓

结算

↓

完成

---

# 退款流程

用户申请退款

↓

平台审核

↓

原路退回

↓

取消订阅

↓

回收权限

↓

完成

---

# Creator中心

作者可以查看：

累计收入

本月收入

待结算收入

已结算收入

订阅人数

付费用户数

续费率

---

# 用户中心

用户可以查看：

我的会员

我的订阅

我的订单

我的发票

我的退款

---

# 权限模型

FREE：

不可发布收费内容

---

PRO：

可发布收费策略

---

MASTER：

可发布收费策略

可发布收费组合

可申请创作者认证

---

ENTERPRISE：

机构发布能力

---

# 风控设计

检测：

异常支付

刷单

自买自卖

恶意退款

提现欺诈

---

处罚：

冻结收益

限制提现

封禁账号

---

# 数据统计

会员收入

订阅收入

创作者收入

平台收入

ARPU

MRR

ARR

付费率

续费率

流失率

---

# KPI

V1目标：

1000注册用户

100付费用户

付费率 \> 10%

会员收入 \> 1000 USD/月

---

V2目标：

10000用户

1000付费用户

100名创作者

---

# MVP范围

必须实现：

- Membership

- Order

- Payment

- Strategy Subscription

- Revenue Share

---

暂缓：

- Coupon

- Invoice

- Enterprise Billing

- Tax Management

- Multi\-Currency Settlement

---

# Billing Service 与平台关系

User Service

↓

Billing Service

↓

Membership

Strategy Service

↓

Billing Service

↓

Strategy Subscription

Portfolio Service

↓

Billing Service

↓

Portfolio Subscription

Community Service

↓

Billing Service

↓

Creator Economy

Ranking Service

↓

Billing Service

↓

Top Creators

Notification Service

↓

Billing Service

↓

Payment Notification

---

# 长期演进路线

V1：

会员体系

V2：

策略订阅市场

V3：

组合订阅市场

V4：

研究报告市场

V5：

Creator Economy

V6：

Quant Marketplace

最终目标：

让 QuantLab 从

量化工具平台

演进为

量化内容交易平台


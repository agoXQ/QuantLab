# QuantLab MVP Roadmap v1\.0

Version: 1\.0

Status: Product Execution Plan

目标：

```Plain Text
6个月内上线可用版本

支持真实用户使用

支持公式选股

支持回测

支持策略分享

验证产品价值
```

---

# 一、产品愿景

打造：

```Plain Text
面向个人投资者的 AI 原生量化研究平台
```

用户可以：

```Plain Text
编写策略
↓
筛选股票
↓
回测验证
↓
构建组合
↓
分享策略
↓
进入排行榜
↓
获得关注
```

未来：

```Plain Text
策略市场
策略订阅
AI研究员
```

---

# 二、MVP核心目标

必须验证三个问题：

---

## 验证1

用户是否愿意写策略

例如：

```Plain Text
ROE > 15
PE < 20
MA5 > MA20
```

---

## 验证2

用户是否愿意回测

例如：

```Plain Text
近5年收益率

最大回撤

夏普比率
```

---

## 验证3

用户是否愿意分享策略

例如：

```Plain Text
公开策略

排行榜

社区讨论
```

---

如果这三个成立：

```Plain Text
产品成立
```

否则：

```Plain Text
停止扩张
重新调整方向
```

---

# 三、MVP范围

## 必须开发（P0）

### User Service

必须：

```Plain Text
注册

登录

JWT认证

个人主页
```

---

### Strategy Service

必须：

```Plain Text
创建策略

编辑策略

版本管理

发布策略
```

---

### Formula Engine

必须：

```Plain Text
DSL解析

指标计算

选股执行
```

---

### Market Data Service

必须：

```Plain Text
股票基础信息

日线行情

复权处理
```

---

### Backtest Engine

必须：

```Plain Text
策略回测

收益率

回撤

夏普
```

---

### Ranking Service

必须：

```Plain Text
收益率排行

胜率排行
```

---

### Portfolio Service（简化版）

必须：

```Plain Text
策略组合

组合收益统计
```

---

# 四、暂缓开发（P2）

---

## Community Service

第一版：

```Plain Text
不要开发论坛
```

只保留：

```Plain Text
策略评论
```

即可。

---

## Notification Service

第一版：

```Plain Text
站内通知
```

即可。

不做：

```Plain Text
短信

邮件

微信
```

---

## AI Service

第一版不要做 Agent。

仅做：

```Plain Text
AI解释策略
AI分析回测
```

---

不做：

```Plain Text
AI自动生成策略

AI自动调参

AI研究员
```

---

# 五、第一版服务架构

不要直接10个微服务。

---

推荐：

## Phase 1

Monolith First

```Plain Text
quantlab-app

├ user
├ strategy
├ formula
├ market
├ backtest
├ ranking
├ portfolio
```

一个仓库。

一个数据库。

---

原因：

```Plain Text
开发快

维护简单

调试简单
```

---

# 六、数据库规划

MVP：

```Plain Text
MySQL
```

即可。

---

不要：

```Plain Text
Kafka

ClickHouse

Qdrant
```

---

先做：

```Plain Text
MySQL

Redis
```

---

# 七、行情数据规划

推荐：

## A股

使用：

Tushare

或者：

AkShare

---

原因：

```Plain Text
开发快

成本低

覆盖足够
```

---

不要第一版自建采集系统。

---

# 八、公式系统 MVP

支持：

## 基础字段

```Plain Text
OPEN
HIGH
LOW
CLOSE
VOL
AMOUNT
```

---

## 财务字段

```Plain Text
PE
PB
ROE
ROA
```

---

## 技术指标

```Plain Text
MA
EMA
MACD
RSI
KDJ
```

---

## 逻辑

```Plain Text
AND
OR
NOT
```

---

即可。

---

不要支持：

```Plain Text
自定义函数

脚本插件

Python策略
```

---

# 九、回测 MVP

支持：

```Plain Text
日线回测
```

---

支持：

```Plain Text
调仓周期

手续费

滑点
```

---

支持：

```Plain Text
收益率

最大回撤

夏普
```

---

不要：

```Plain Text
分钟级回测

Tick级回测

多账户回测
```

---

# 十、排行榜 MVP

支持：

```Plain Text
年化收益率

累计收益率

胜率

最大回撤
```

---

周期：

```Plain Text
近1月

近3月

近1年

全部
```

---

即可。

---

# 十一、前端 MVP

推荐：

```Plain Text
Next.js
```

---

页面：

## 首页

```Plain Text
热门策略

热门组合

排行榜
```

---

## 策略页

```Plain Text
策略详情

回测结果

评论
```

---

## 编辑器

```Plain Text
策略DSL编辑器
```

---

## 个人中心

```Plain Text
我的策略

我的回测

我的组合
```

---

# 十二、团队规模

最小团队：

---

## 后端

```Plain Text
2人
```

---

## 前端

```Plain Text
1人
```

---

## 产品

```Plain Text
1人
```

---

## 测试

```Plain Text
兼职
```

---

总计：

```Plain Text
3~4人
```

---

# 十三、研发周期

## 第一阶段

基础框架

```Plain Text
2周
```

---

## 第二阶段

策略系统

```Plain Text
4周
```

---

## 第三阶段

公式引擎

```Plain Text
4周
```

---

## 第四阶段

行情系统

```Plain Text
2周
```

---

## 第五阶段

回测系统

```Plain Text
4周
```

---

## 第六阶段

排行榜

```Plain Text
2周
```

---

## 第七阶段

前端整合

```Plain Text
4周
```

---

总计：

```Plain Text
18~22周
```

约：

```Plain Text
4~5个月
```

---

# 十四、Phase 2（产品验证成功）

当达到：

```Plain Text
1000注册用户

100DAU

1000次回测/天
```

启动第二阶段。

---

新增：

```Plain Text
Community Service

Notification Service

AI Service
```

---

架构升级：

```Plain Text
单体
↓
模块化单体
↓
微服务
```

---

# 十五、Phase 3（商业化）

新增：

```Plain Text
Subscription Service

Billing Service
```

---

支持：

```Plain Text
策略订阅

组合订阅

创作者收益
```

---

形成：

```Plain Text
Creator Economy
```

---

# 十六、最终落地建议（非常关键）

基于我们已经设计的全部内容，我建议你的实际开发顺序不是：

```Plain Text
User
↓
Strategy
↓
Formula
↓
Market
↓
Backtest
↓
Ranking
↓
Portfolio
↓
Community
↓
AI
```

而是：

```Plain Text
第一优先级
Formula Engine
Backtest Engine
Market Data
```

因为：

```Plain Text
公式引擎
+
行情数据
+
回测系统
```

才是 QuantLab 的核心壁垒。

用户系统、社区、AI 都可以后补。

如果我是 CTO，我会要求团队前 8 周只做三件事：

```Plain Text
公式DSL

选股执行器

回测引擎
```

做到能够完成下面这个闭环：

```Plain Text
输入策略
↓
选股
↓
回测
↓
输出收益率
```

只要这个闭环跑通，QuantLab 就已经具备最小可用价值（MVP）。

然后再逐步叠加：

```Plain Text
用户
→ 排行榜
→ 组合
→ 社区
→ AI
→ 订阅市场
```




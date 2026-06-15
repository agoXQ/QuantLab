# QuantLab Ranking Service TD

Version: 1\.0

Module: Ranking Service

Priority: P0

Status: Draft

Owner: Architecture Team

---

# 服务定位

负责：

```Plain Text
策略评分

排行榜计算

作者评分

曝光推荐

反作弊评分

排行榜快照
```

不负责：

```Plain Text
回测执行

用户管理

策略管理
```

---

# 服务边界

拥有：

```Plain Text
Ranking

RankingSnapshot

Score

TrustScore

AuthorScore
```

只读：

```Plain Text
BacktestReport

Strategy

User
```

---

# 核心目标

建立：

```Plain Text
收益
+
风险
+
稳定性
+
可信度
=
综合评分
```

而不是：

```Plain Text
收益率排序
```

---

# 排行榜体系

V1支持：

```Plain Text
收益榜

胜率榜

夏普榜

回撤榜

综合榜

作者榜
```

---

V2支持：

```Plain Text
行业榜

风格榜

AI策略榜

新人榜
```

---

# 技术架构

```Plain Text
ranking-service

├── calculator
├── scorer
├── anti-cheat
├── snapshot
├── recommendation
├── cache
├── event
└── api
```

---

# 目录结构

```Plain Text
ranking-service
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── scorer
│   │
│   ├── calculator
│   │
│   ├── anti_cheat
│   │
│   ├── snapshot
│   │
│   ├── recommendation
│   │
│   ├── cache
│   │
│   └── event
│
└── deploy
```

---

# 排名计算流程

```Plain Text
BacktestFinished
        ↓
Score Calculator
        ↓
Trust Calculator
        ↓
Ranking Builder
        ↓
Ranking Snapshot
        ↓
Expose
```

---

# DDD聚合设计

---

## Ranking Aggregate

聚合根：

```Plain Text
Ranking
```

---

包含：

```Plain Text
RankingItem
```

---

## Score Aggregate

聚合根：

```Plain Text
Score
```

---

## Trust Aggregate

聚合根：

```Plain Text
TrustScore
```

---

## Snapshot Aggregate

聚合根：

```Plain Text
RankingSnapshot
```

---

# 数据库设计

数据库：

```Plain Text
quantlab_ranking
```

---

## strategy\_score

核心表

```Plain Text
CREATE TABLE strategy_score
(
    strategy_id BIGINT PRIMARY KEY,

    total_return DECIMAL(10,4),

    annual_return DECIMAL(10,4),

    sharpe_ratio DECIMAL(10,4),

    max_drawdown DECIMAL(10,4),

    win_rate DECIMAL(10,4),

    volatility DECIMAL(10,4),

    score DECIMAL(10,4),

    trust_score DECIMAL(10,4),

    updated_at DATETIME
);
```

---

## ranking\_snapshot

```Plain Text
CREATE TABLE ranking_snapshot
(
    id BIGINT PRIMARY KEY,

    ranking_type VARCHAR(64),

    ranking_period VARCHAR(32),

    snapshot_time DATETIME
);
```

---

## ranking\_item

```Plain Text
CREATE TABLE ranking_item
(
    snapshot_id BIGINT,

    rank_no INT,

    strategy_id BIGINT,

    score DECIMAL(10,4),

    PRIMARY KEY(snapshot_id, rank_no)
);
```

---

## author\_score

```Plain Text
CREATE TABLE author_score
(
    author_id BIGINT PRIMARY KEY,

    strategy_count INT,

    avg_return DECIMAL(10,4),

    avg_sharpe DECIMAL(10,4),

    follower_count BIGINT,

    author_score DECIMAL(10,4)
);
```

---

## trust\_score\_log

```Plain Text
CREATE TABLE trust_score_log
(
    id BIGINT PRIMARY KEY,

    strategy_id BIGINT,

    old_score DECIMAL(10,4),

    new_score DECIMAL(10,4),

    reason VARCHAR(256),

    created_at DATETIME
);
```

---

# Score Engine

输入：

```Plain Text
BacktestReport
```

输出：

```Plain Text
Score
```

---

# 综合评分模型

```Plain Text
Score =
  40% 收益率
+ 20% Sharpe
+ 15% 胜率
+ 15% 稳定性
+ 10% 社区热度
```

公式：

```Plain Text
Score =
  0.40 × ReturnScore
+ 0.20 × SharpeScore
+ 0.15 × WinRateScore
+ 0.15 × StabilityScore
+ 0.10 × CommunityScore
```

---

# 收益评分

归一化：

```Plain Text
0~100
```

---

避免：

```Plain Text
3000%

直接碾压其它策略
```

---

# 风险评分

考虑：

```Plain Text
波动率

最大回撤

卡玛比率
```

---

# 稳定性评分

观察：

```Plain Text
月收益连续性

季度收益连续性
```

---

避免：

```Plain Text
一次暴涨
长期失效
```

---

# Trust Score

核心能力。

用于：

```Plain Text
反作弊

反过拟合

反刷榜
```

---

# Trust Score组成

```Plain Text
回测长度

样本数量

交易次数

稳定性

用户反馈
```

---

# 回测长度评分

例如：

```Plain Text
3个月
=
30分

1年
=
70分

5年
=
100分
```

---

# 样本评分

```Plain Text
10次交易
=
低可信

500次交易
=
高可信
```

---

# 防过拟合检测

检测：

```Plain Text
收益异常高

交易极少

集中持仓
```

---

触发：

```Plain Text
TrustScore下降
```

---

# Anti\-Cheat Engine

检测：

```Plain Text
参数作弊

未来函数

幸存者偏差

数据窥探
```

---

# 未来函数

来自：

```Plain Text
Formula Engine
```

---

直接：

```Plain Text
TrustScore = 0
```

---

# 排行榜类型

---

收益榜

```Plain Text
RETURN
```

---

夏普榜

```Plain Text
SHARPE
```

---

回撤榜

```Plain Text
DRAWDOWN
```

---

综合榜

```Plain Text
OVERALL
```

---

作者榜

```Plain Text
AUTHOR
```

---

# 时间维度

支持：

```Plain Text
DAILY

WEEKLY

MONTHLY

YEARLY

ALL_TIME
```

---

# 排行榜快照

每日生成：

```Plain Text
00:05
```

---

形成：

```Plain Text
不可修改历史记录
```

---

# API设计

统一前缀：

```Plain Text
/api/v1/rankings
```

---

## 获取排行榜

```Plain Text
GET /{type}
```

参数：

```Plain Text
period

page

size
```

---

## 获取策略排名

```Plain Text
GET /strategy/{id}
```

---

## 获取作者排名

```Plain Text
GET /author/{id}
```

---

## 获取历史排名

```Plain Text
GET /history
```

---

## 获取排行榜快照

```Plain Text
GET /snapshots
```

---

# 事件设计

Topic：

```Plain Text
ranking-events
```

---

## RankingCalculated

```Plain Text
{
  "ranking":"OVERALL"
}
```

---

## StrategyEnteredRanking

```Plain Text
{
  "strategy_id":1001,

  "rank":8
}
```

---

消费者：

```Plain Text
Community Service

Notification Service
```

---

## StrategyDroppedRanking

```Plain Text
{
  "strategy_id":1001
}
```

---

## TrustScoreChanged

```Plain Text
{
  "strategy_id":1001,

  "score":60
}
```

---

# Redis缓存设计

综合榜：

```Plain Text
ranking:overall
```

---

收益榜：

```Plain Text
ranking:return
```

---

作者榜：

```Plain Text
ranking:author
```

---

TTL：

```Plain Text
5min
```

---

## Redis Sorted Set设计

推荐：

```Plain Text
ZSET
```

---

例如：

```Plain Text
ranking:return
```

score：

```Plain Text
收益率
```

member：

```Plain Text
strategy_id
```

---

# 计算调度

实时：

```Plain Text
BacktestFinished
```

增量更新。

---

定时：

```Plain Text
每日全量重建
```

---

# 推荐引擎

排行榜只是曝光的一部分。

增加：

```Plain Text
猜你喜欢
```

---

输入：

```Plain Text
用户收藏

浏览

关注
```

---

输出：

```Plain Text
推荐策略
```

---

# 部署架构

```Plain Text
Backtest Events

↓

Ranking Service

↓

PostgreSQL

↓

Redis

↓

Kafka
```

---

# K8S部署

```Plain Text
ranking-api

3 Pods

↓

ranking-worker

5 Pods
```

---

# 高可用

PostgreSQL：

```Plain Text
1 Primary

2 Replica
```

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

# 权限设计

公开：

```Plain Text
排行榜浏览
```

---

会员：

```Plain Text
高级排行榜

历史排名分析
```

---

管理员：

```Plain Text
排行榜重建
```

---

# 性能指标

获取排行榜：

```Plain Text
P95 < 20ms
```

---

获取排名：

```Plain Text
P95 < 30ms
```

---

榜单重建：

```Plain Text
100万策略

< 10分钟
```

---

QPS：

```Plain Text
20000+
```

---

# 可观测性

Metrics：

```Plain Text
ranking_calculated_total

ranking_query_total

trust_score_changed_total
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

99.95%
```

---

```Plain Text
RTO

15分钟
```

---

```Plain Text
RPO

5分钟
```

---

# MVP范围

必须实现：

✅ 收益榜

✅ 夏普榜

✅ 综合榜

✅ 作者榜

✅ Trust Score

✅ 排行榜快照

✅ Redis ZSET

✅ Kafka事件

---

暂缓：

❌ AI推荐

❌ 协同过滤

❌ 深度学习推荐

❌ 社交关系推荐

---

# 后续演进

当策略规模达到 10 万+、作者 1 万+ 时，可拆分为：

```Plain Text
ranking-platform
│
├── score-service
├── trust-service
├── ranking-service
└── recommendation-service
```

核心链路：

```Plain Text
Backtest → Score → Trust → Ranking → Recommendation → Community
```




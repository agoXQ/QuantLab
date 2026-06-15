# QuantLab Community Service TD

Version: 2\.0

Module: Community Service

Priority: P1

Status: Draft

Owner: Architecture Team

Dependencies:

```Plain Text
User Service
Strategy Service
Portfolio Service
Ranking Service
Notification Service
```

---

# 1\. 服务定位

Community Service 是 QuantLab 的：

```Plain Text
Research Community Platform
```

负责：

```Plain Text
内容发布

内容聚合

Feed流

点赞

评论

收藏

关注

内容推荐

作者主页
```

不负责：

```Plain Text
策略管理

组合管理

回测执行

排行榜计算
```

---

# 2\. 核心设计原则

## Principle 1

内容统一抽象

所有内容统一为：

```Plain Text
Content
```

---

## Principle 2

内容与领域对象解耦

Community 不保存：

```Plain Text
Strategy

Portfolio
```

完整数据。

只保存引用。

---

## Principle 3

社交关系独立

Follow 关系数据归属 User Service。

Community Service 基于关注关系驱动 Feed 与通知。

---

## Principle 4

Feed优先

Community核心不是帖子。

而是：

```Plain Text
Feed
```

---

# 3\. 服务边界

拥有：

```Plain Text
Content

Comment

Like

Favorite

Feed
```

引用（数据归属 User Service）：

```Plain Text
Follow
UserProfile
```

引用：

```Plain Text
Strategy

Portfolio

ResearchArticle
```

---

# 4\. 内容模型

统一抽象：

```Plain Text
type Content struct {
    ID
    Type
    ObjectID
    AuthorID
}
```

---

# 5\. Content Type

```Plain Text
STRATEGY

PORTFOLIO

ARTICLE
```

未来预留：

```Plain Text
NOTE

VIDEO

LIVE
```

---

# 6\. 架构设计

```Plain Text
community-service

├── content
├── feed
├── social
├── interaction
├── recommendation
├── moderation
├── cache
├── event
└── api
```

---

# 7\. 目录结构

```Plain Text
community-service
│
├── cmd
│
├── internal
│
│   ├── api
│   │
│   ├── content
│   │
│   ├── feed
│   │
│   ├── social
│   │
│   ├── interaction
│   │
│   ├── recommendation
│   │
│   ├── moderation
│   │
│   ├── cache
│   │
│   └── event
│
└── deploy
```

---

# 8\. DDD聚合设计

## Content Aggregate

聚合根：

```Plain Text
type Content struct {
    ID
    Type
    ObjectID
    AuthorID
    Title
    Summary
    Visibility
}
```

---

## Comment Aggregate

```Plain Text
type Comment struct {
    ID
    ContentID
    UserID
    ParentID
    Body
}
```

---

## Favorite Aggregate

```Plain Text
type Favorite struct {
    UserID
    ContentID
}
```

---

## Feed Aggregate

```Plain Text
type FeedItem struct {
    UserID
    ContentID
    Score
}
```

---

# 9\. 数据库设计

数据库：

```Plain Text
quantlab_community
```

---

# 10\.  content

核心内容表

```Plain Text
CREATE TABLE content
(
    id BIGINT PRIMARY KEY,

    content_type VARCHAR(32),

    object_id BIGINT,

    author_id BIGINT,

    title VARCHAR(256),

    summary TEXT,

    visibility VARCHAR(32),

    status VARCHAR(32),

    created_at DATETIME,

    updated_at DATETIME
);
```

---

# 11\. content\_stat

统计信息

```Plain Text
CREATE TABLE content_stat
(
    content_id BIGINT PRIMARY KEY,

    like_count BIGINT,

    favorite_count BIGINT,

    comment_count BIGINT,

    view_count BIGINT
);
```

---

# 12\. comment

```Plain Text
CREATE TABLE comment
(
    id BIGINT PRIMARY KEY,

    content_id BIGINT,

    user_id BIGINT,

    parent_id BIGINT,

    content TEXT,

    status VARCHAR(32),

    created_at DATETIME
);
```

---

# 13\. favorite\_record

```Plain Text
CREATE TABLE favorite_record
(
    user_id BIGINT,

    content_id BIGINT,

    created_at DATETIME,

    PRIMARY KEY(user_id,content_id)
);
```

---

# 15\. like\_record

```Plain Text
CREATE TABLE like_record
(
    user_id BIGINT,

    target_type VARCHAR(32),

    target_id BIGINT,

    created_at DATETIME,

    PRIMARY KEY(user_id,target_type,target_id)
);
```

---

target\_type

```Plain Text
CONTENT

COMMENT
```

---

# 15\. feed\_item

预计算Feed

```Plain Text
CREATE TABLE feed_item
(
    user_id BIGINT,

    content_id BIGINT,

    score DECIMAL(10,4),

    created_at DATETIME,

    PRIMARY KEY(user_id, content_id)
);
```

---

# 18\. 内容发布流程

```Plain Text
Strategy Publish
        ↓
Content Created
        ↓
Feed Fanout
        ↓
Feed Ready
```

---

# 19\. Content创建流程

Strategy：

```Plain Text
Strategy Service
        ↓
Publish
        ↓
Community Content
```

---

Portfolio：

```Plain Text
Portfolio Service
        ↓
Publish
        ↓
Community Content
```

---

Article：

```Plain Text
Community Native Content
```

---

# 20\. Feed Engine

负责：

```Plain Text
首页动态流
```

---

输入：

```Plain Text
关注作者

排行榜

热门内容

推荐内容
```

---

输出：

```Plain Text
Feed Items
```

---

# 21\. Feed排序模型

V1

```Plain Text
Time Score
```

---

V2

```Plain Text
Hot Score
```

---

V3

```Plain Text
Personalized Score
```

---

# 22\. Hot Score

```Plain Text
点赞 × 2

+
评论 × 5

+
收藏 × 8

+
浏览 × 0.1

+
排行榜加权
```

---

# 23\. 排行榜融合

Ranking Service提供：

```Plain Text
Top Strategy

Top Portfolio
```

---

Community融合：

```Plain Text
热门内容流
```

---

# 24\. 推荐系统

V1：

```Plain Text
关注关系

热门内容

排行榜内容
```

---

V2：

```Plain Text
浏览行为

点赞行为

收藏行为

关注行为
```

---

# 25\. 内容审核

模块：

```Plain Text
moderation
```

---

支持：

```Plain Text
敏感词

广告

垃圾信息
```

---

状态：

```Plain Text
PENDING

APPROVED

REJECTED
```

---

# 26\. API设计

统一前缀：

```Plain Text
/api/v1/community
```

---

# 27\. 创建内容

```Plain Text
POST /contents
```

Request：

```Plain Text
{
  "content_type":"STRATEGY",
  "object_id":1001,
  "title":"双均线策略",
  "summary":"年化收益28%"
}
```

---

# 28\. 获取内容详情

```Plain Text
GET /contents/{id}
```

---

# 29\. 评论内容

```Plain Text
POST /contents/{id}/comments
```

---

# 30\. 点赞内容

```Plain Text
POST /contents/{id}/likes
```

---

# 31\. 收藏内容

```Plain Text
POST /contents/{id}/favorites
```

---

# 31\. 关注用户

```Plain Text
POST /follows
```

注意：关注关系数据写入 User Service，Community Service 转发请求或消费 UserFollowed 事件。

---

# 33\. 获取Feed

```Plain Text
GET /feed
```

参数：

```Plain Text
cursor
pageSize
```

---

# 34\. 获取用户主页

```Plain Text
GET /profiles/{userId}
```

---

# 35\. 获取用户内容

```Plain Text
GET /profiles/{userId}/contents
```

---

# 36\. 事件设计

Topic：

```Plain Text
community-events
```

---

# 36\. ContentCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "ContentCreated",
  "aggregate_type": "CONTENT",
  "aggregate_id": "1001",
  "producer": "community-service",
  "payload": {
    "content_id": 1001,
    "content_type": "STRATEGY",
    "object_id": 5001
  }
}
```

---

# 37\. ContentLiked

```JSON
{
  "event_id": "uuid",
  "event_type": "ContentLiked",
  "aggregate_type": "CONTENT",
  "aggregate_id": "1001",
  "payload": {
    "content_id": 1001,
    "user_id": 2001
  }
}
```

---

# 38\. ContentFavorited

```JSON
{
  "event_id": "uuid",
  "event_type": "ContentFavorited",
  "aggregate_type": "CONTENT",
  "aggregate_id": "1001",
  "payload": {
    "content_id": 1001,
    "user_id": 2001
  }
}
```

---

# 39\. CommentCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "CommentCreated",
  "aggregate_type": "COMMENT",
  "aggregate_id": "1001",
  "payload": {
    "comment_id": 1001,
    "content_id": 2001
  }
}
```

---

# 40\. ContentTrending

```JSON
{
  "event_id": "uuid",
  "event_type": "ContentTrending",
  "aggregate_type": "CONTENT",
  "aggregate_id": "1001",
  "payload": {
    "content_id": 1001,
    "hot_score": 95.2
  }
}
```

---

# 43\. 事件消费者

Notification Service：

```Plain Text
点赞通知

评论通知

关注通知
```

---

AI Service：

```Plain Text
用户兴趣建模

推荐建模
```

---

Ranking Service：

```Plain Text
热度融合
```

---

# 44\. Redis缓存设计

内容详情：

```Plain Text
content:{id}
```

TTL：

```Plain Text
30min
```

---

用户主页：

```Plain Text
profile:{userId}
```

TTL：

```Plain Text
1h
```

---

Feed：

```Plain Text
feed:{userId}
```

TTL：

```Plain Text
5min
```

---

热门内容：

```Plain Text
community:hot
```

TTL：

```Plain Text
1min
```

---

# 45\. Feed架构

采用：

```Plain Text
Push + Pull Hybrid
```

---

普通用户：

```Plain Text
Push Feed
```

---

大V用户：

```Plain Text
Pull Feed
```

---

阈值：

```Plain Text
粉丝 > 10000
```

自动切换。

---

# 46\. 权限设计

游客：

```Plain Text
浏览公开内容
```

---

注册用户：

```Plain Text
点赞

评论

收藏

关注
```

---

Pro：

```Plain Text
高级推荐

深度分析
```

---

管理员：

```Plain Text
审核

删除

封禁
```

---

# 47\. 部署架构

```Plain Text
Community API
      ↓
Feed Worker
      ↓
PostgreSQL
      ↓
Redis
      ↓
Kafka
```

---

# 48\. 性能指标

Feed查询：

```Plain Text
P95 < 50ms
```

---

内容详情：

```Plain Text
P95 < 30ms
```

---

评论：

```Plain Text
P95 < 50ms
```

---

QPS：

```Plain Text
10000+
```

---

# 49\. SLA

```Plain Text
Availability

99.95%
```

```Plain Text
RTO

15min
```

```Plain Text
RPO

5min
```

---

# 50\. MVP范围

必须实现：

✅ Content统一模型

✅ Strategy展示

✅ Portfolio展示

✅ Research Article

✅ 评论

✅ 点赞

✅ 收藏

✅ 关注

✅ Feed流

✅ 内容审核

---

暂缓：

❌ 私信

❌ 群组

❌ 直播

❌ 跟投

❌ 策略交易市场

---

# 后续演进

V3 社区将演变为完整的研究社区平台：

```Plain Text
Research Community
│
├── Strategy
├── Portfolio
├── Research Article
├── Ranking
├── Feed
├── Recommendation
└── Subscription
```




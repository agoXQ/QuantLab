# QuantLab User Service TD

Version: 1\.0

Module: User Service

Status: Draft

Owner: Architecture Team

Last Update: 2026

---

# 服务定位

User Service 是 QuantLab 的身份中心（Identity Center）。

负责：

- 用户管理

- 登录认证

- 权限管理

- 会员管理

- 创作者体系

- 关注关系

- 荣誉体系

属于：

```Plain Text
Infrastructure Domain
```

---

# 服务边界

拥有的数据：

```Plain Text
User
Profile
Membership
Role
Permission
Follow
Honor
```

禁止拥有：

```Plain Text
Strategy
Backtest
Ranking
Comment
Article
```

这些数据属于其它服务。

---

# 技术架构

```Plain Text
user-service

├── api
├── application
├── domain
├── infrastructure
├── event
├── scheduler
├── cache
├── security
└── test
```

---

# 目录结构

```Plain Text
user-service
│
├── cmd
│   └── server
│
├── internal
│
│   ├── api
│   │
│   ├── application
│   │   ├── command
│   │   ├── query
│   │   └── service
│   │
│   ├── domain
│   │   ├── user
│   │   ├── profile
│   │   ├── membership
│   │   ├── role
│   │   ├── follow
│   │   └── honor
│   │
│   ├── infrastructure
│   │   ├── postgresql
│   │   ├── redis
│   │   ├── kafka
│   │   └── jwt
│   │
│   ├── event
│   │
│   ├── security
│   │
│   └── scheduler
│
└── deploy
```

---

# DDD聚合设计

## User Aggregate

聚合根：

```Plain Text
User
```

包含：

```Plain Text
User
Profile
Membership
```

---

## Follow Aggregate

聚合根：

```Plain Text
FollowRelation
```

注意：关注关系数据由 User Service 管理。Community Service 基于本模块的关注关系驱动 Feed 与通知。

---

---

## Honor Aggregate

聚合根：

```Plain Text
Honor
```

---

# 数据库设计

数据库：

```SQL
quantlab_user
```

---

## user表

```SQL
CREATE TABLE users
(
    id BIGINT PRIMARY KEY,

    username VARCHAR(64) UNIQUE,

    email VARCHAR(128) UNIQUE,

    phone VARCHAR(32) UNIQUE,

    password_hash VARCHAR(255),

    status TINYINT,

    creator_status TINYINT,

    verified_status TINYINT,

    created_at DATETIME,

    updated_at DATETIME,

    deleted_at DATETIME NULL
);
```

---

## user\_profile表

```SQL
CREATE TABLE user_profile
(
    user_id BIGINT PRIMARY KEY,

    nickname VARCHAR(64),

    avatar VARCHAR(255),

    bio TEXT,

    location VARCHAR(128),

    website VARCHAR(255),

    experience_years INT,

    preferred_market VARCHAR(64)
);
```

---

## membership表

```SQL
CREATE TABLE membership
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT,

    tier VARCHAR(32),

    start_time DATETIME,

    expire_time DATETIME,

    status TINYINT
);
```

---

## role表

```SQL
CREATE TABLE roles
(
    id BIGINT PRIMARY KEY,

    role_code VARCHAR(64),

    role_name VARCHAR(128)
);
```

---

## permission表

```SQL
CREATE TABLE permissions
(
    id BIGINT PRIMARY KEY,

    permission_code VARCHAR(128),

    permission_name VARCHAR(128)
);
```

---

## user\_role表

```SQL
CREATE TABLE user_role
(
    user_id BIGINT,

    role_id BIGINT,

    PRIMARY KEY(user_id, role_id)
);
```

---

## role\_permission表

```SQL
CREATE TABLE role_permission
(
    role_id BIGINT,

    permission_id BIGINT,

    PRIMARY KEY(role_id, permission_id)
);
```

---

## follow\_relation表

```SQL
CREATE TABLE follow_relation
(
    id BIGINT PRIMARY KEY,

    follower_id BIGINT,

    following_id BIGINT,

    created_at DATETIME,

    UNIQUE(follower_id, following_id)
);
```

---

## honor表

```SQL
CREATE TABLE honor
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT,

    honor_type VARCHAR(64),

    honor_name VARCHAR(128),

    granted_at DATETIME
);
```

---

## login\_log表

```SQL
CREATE TABLE login_log
(
    id BIGINT PRIMARY KEY,

    user_id BIGINT,

    ip VARCHAR(64),

    user_agent TEXT,

    login_time DATETIME
);
```

---

# API设计

统一前缀：

```HTTP
/api/v1/users
```

---

## 注册

```HTTP
POST /register
```

Request

```JSON
{
  "email":"test@test.com",
  "password":"xxxx"
}
```

Response

```JSON
{
  "user_id":10001
}
```

---

## 登录

```HTTP
POST /login
```

Response

```JSON
{
  "access_token":"",
  "refresh_token":""
}
```

---

## 获取用户信息

```HTTP
GET /me
```

---

## 修改资料

```HTTP
PUT /profile
```

---

## 获取用户主页

```HTTP
GET /users/{id}
```

---

## 关注用户

```HTTP
POST /users/{id}/follow
```

---

## 取消关注

```HTTP
DELETE /users/{id}/follow
```

---

## 粉丝列表

```HTTP
GET /users/{id}/followers
```

---

## 关注列表

```HTTP
GET /users/{id}/following
```

---

## 会员升级

```HTTP
POST /membership/upgrade
```

---

# JWT设计

Access Token

```Plain Text
15分钟
```

---

Refresh Token

```Plain Text
30天
```

---

JWT Payload

```JSON
{
  "uid":10001,
  "roles":["USER"],
  "tier":"PRO"
}
```

---

# 事件设计

Topic:

```Plain Text
user-events
```

---

## UserCreated

```JSON
{
  "event_id": "uuid",
  "event_type": "UserCreated",
  "event_version": "1.0",
  "aggregate_type": "USER",
  "aggregate_id": "10001",
  "occurred_at": "2026-01-01T00:00:00Z",
  "producer": "user-service",
  "trace_id": "trace_xxx",
  "payload": {
    "user_id": 10001
  }
}
```

---

## UserProfileUpdated

```JSON
{
  "event_id": "uuid",
  "event_type": "UserProfileUpdated",
  "aggregate_type": "USER",
  "aggregate_id": "10001",
  "payload": {
    "user_id": 10001
  }
}
```

---

## MembershipUpgraded

```JSON
{
  "event_id": "uuid",
  "event_type": "MembershipUpgraded",
  "aggregate_type": "USER",
  "aggregate_id": "10001",
  "payload": {
    "user_id": 10001,
    "tier": "PRO"
  }
}
```

---

## UserFollowed

```JSON
{
  "event_id": "uuid",
  "event_type": "UserFollowed",
  "aggregate_type": "USER",
  "aggregate_id": "1",
  "payload": {
    "follower_id": 1,
    "following_id": 2
  }
}
```

---

## UserVerified

```JSON
{
  "event_id": "uuid",
  "event_type": "UserVerified",
  "aggregate_type": "USER",
  "aggregate_id": "10001",
  "payload": {
    "user_id": 10001
  }
}
```

---

# Redis缓存设计

Redis Key:

用户信息

```Plain Text
user:profile:{uid}
```

TTL

```Plain Text
1小时
```

---

权限缓存

```Plain Text
user:permission:{uid}
```

TTL

```Plain Text
30分钟
```

---

粉丝数

```Plain Text
user:followers:{uid}
```

---

关注数

```Plain Text
user:following:{uid}
```

---

# 缓存更新策略

Cache Aside Pattern

```Plain Text
DB
↓
Delete Cache
↓
Read Through
```

---

# 权限模型

RBAC

---

Role

```Plain Text
GUEST

USER

CREATOR

MASTER

Pro

ADMIN

SUPER_ADMIN
```

---

Permission

```Plain Text
strategy:create

strategy:publish

backtest:create

community:comment

admin:user
```

---

# 鉴权流程

```Plain Text
Request

↓

JWT验证

↓

获取Role

↓

获取Permission

↓

访问资源
```

---

# 安全设计

密码：

```Plain Text
Argon2
```

---

禁止：

```Plain Text
MD5

SHA1
```

---

登录限流：

```Plain Text
5次失败

锁定15分钟
```

---

# 风控设计

检测：

```Plain Text
异常IP

异地登录

批量注册

撞库攻击
```

---

# 部署架构

```Plain Text
Nginx

↓

API Gateway

↓

user-service

↓

PostgreSQL Primary

↓

PostgreSQL Replica

↓

Redis

↓

Kafka
```

---

# 高可用设计

User Service

```Plain Text
3 Pods
```

K8S部署

---

PostgreSQL

```Plain Text
1 Primary

2 Replica
```

---

Redis

```Plain Text
Sentinel
```

---

Kafka

```Plain Text
3 Broker
```

---

# 性能指标

登录接口

```Plain Text
P95 < 100ms
```

---

获取用户信息

```Plain Text
P95 < 50ms
```

---

关注用户

```Plain Text
P95 < 80ms
```

---

QPS目标

```Plain Text
登录：
1000 QPS

查询：
5000 QPS

关注：
2000 QPS
```

---

# 可观测性

Metrics

```Plain Text
login_success_total

login_fail_total

user_created_total

follow_created_total
```

---

Trace

```Plain Text
OpenTelemetry
```

---

Log

```Plain Text
JSON Structured Log
```

---

# SLA

可用性：

```Plain Text
99.95%
```

---

RTO

```Plain Text
30分钟
```

---

RPO

```Plain Text
5分钟
```

---

# MVP实现范围

必须实现：

√ 注册

√ 登录

√ JWT

√ 用户资料

√ RBAC

√ 会员体系

√ 关注体系

√ Kafka事件

√ Redis缓存

---

暂缓：

× OAuth

× 企业组织

× SSO

× 多租户

× 2FA

---

# 后续演进

V2

```Plain Text
OAuth2

Google Login

GitHub Login
```

---

V3

```Plain Text
Organization

Team

Enterprise Workspace
```

---

V4

```Plain Text
SSO

IAM

Multi Tenant
```




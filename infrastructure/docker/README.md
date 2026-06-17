# QuantLab Infrastructure

本地开发环境的基础设施，通过 Docker Compose 一键启动。

## 服务清单

| 服务 | 镜像 | 端口 | 说明 |
|---|---|---|---|
| PostgreSQL + TimescaleDB | timescale/timescaledb:2-pg16 | 5432 | 主数据库 + 时序数据 |
| etcd | bitnami/etcd:3.5 | 2379 / 2380 | 服务注册与发现 |
| Redis | redis:7-alpine | 6379 | 缓存 / 会话 / 排行榜 |
| Kafka | apache/kafka:4.0.0 | 9092 | 事件总线（KRaft 模式） |
| Kafka UI | provectuslabs/kafka-ui | 8080 | 消息队列管理界面 |
| MinIO | minio/minio | 9000 / 9001 | S3 兼容对象存储 |

## 快速开始

```bash
# 启动所有服务
make up

# 查看状态
make status

# 查看日志
make logs svc=kafka
```

## 连接信息

### PostgreSQL

```yaml
Host: localhost:5432
User: quantlab
Password: quantlab_dev
Database: quantlab_platform
```

### etcd

```yaml
Endpoint: http://localhost:2379
```

### Redis

```yaml
Host: localhost:6379
Password: quantlab_dev
DB: 0
```

### Kafka

```yaml
Bootstrap: localhost:9092
```

### MinIO

```yaml
API: http://localhost:9000
Console: http://localhost:9001
User: quantlab
Password: quantlab_dev
```

## 数据库

每个微服务拥有独立数据库（One Database Per Service）：

- `quantlab_user` — 用户服务
- `quantlab_strategy` — 策略服务
- `quantlab_portfolio` — 组合服务
- `quantlab_billing` — 计费服务
- `quantlab_community` — 社区服务
- `quantlab_ranking` — 排行榜服务
- `quantlab_notification` — 通知服务
- `quantlab_ai` — AI 服务
- `quantlab_backtest` — 回测引擎
- `quantlab_formula` — 公式引擎
- `quantlab_market_data` — 行情数据服务

## 常用命令

```bash
# 启动
make up

# 停止
make down

# 重启指定服务
make restart svc=postgresql

# 连接 PostgreSQL
make psql

# 连接 Redis
make redis-cli

# 查看 etcd 健康状态
make etcd-health

# 查看 etcd 成员列表
make etcd-members

# 查看 Kafka 主题
make kafka-topics

# 创建 Kafka 主题
make create-topic topic=user-events partitions=3

# 清理数据卷
make clean
```

# ============================================================
# QuantLab Infrastructure — Makefile
# ============================================================

.PHONY: up down restart status logs psql redis-cli kafka-topics clean

# 启动所有基础设施
up:
	docker compose up -d

# 停止所有基础设施
down:
	docker compose down

# 重启指定服务（如：make restart svc=postgresql）
restart:
	docker compose restart $(svc)

# 查看运行状态
status:
	docker compose ps

# 查看日志（如：make logs svc=kafka）
logs:
	docker compose logs -f $(svc)

# 连接 PostgreSQL
psql:
	docker exec -it quantlab-postgresql psql -U quantlab

# 连接 Redis
redis-cli:
	docker exec -it quantlab-redis redis-cli -a quantlab_dev

# 查看 Kafka 主题列表
kafka-topics:
	docker exec -it quantlab-kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

# 创建 Kafka 主题（如：make create-topic topic=user-events partitions=3）
create-topic:
	docker exec quantlab-kafka kafka-topics.sh --bootstrap-server localhost:9092 \
		--create --topic $(topic) --partitions $(partitions) --replication-factor 1

# 查看 MinIO 控制台地址
minio-console:
	@echo "MinIO Console: http://localhost:9001"

# 清理所有数据卷
clean:
	docker compose down -v

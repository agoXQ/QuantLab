package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisCache RedisCacheConfig `json:",optional"`
	Postgres   PostgresConfig   `json:",optional"`
	Kafka      KafkaConfig      `json:",optional"`
	HttpPort   int              `json:",default=8081"`
}

type RedisCacheConfig struct {
	Host string `json:",default=localhost"`
	Port int    `json:",default=6379"`
	Pass string `json:",optional"`
	DB   int    `json:",default=0"`
	TTL  int    `json:",default=86400"`
}

type PostgresConfig struct {
	DSN string `json:",optional"`
}

type KafkaConfig struct {
	Brokers []string `json:",optional"`
}

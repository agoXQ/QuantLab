package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	RedisCache RedisCacheConfig `json:",optional"`
	Postgres   PostgresConfig   `json:",optional"`
	MarketData MarketDataConfig `json:",optional"`
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

// MarketDataConfig configures the in-process Market Data adapter consumed by
// the Formula Engine evaluator.
//
// When DSN is set, the service wires a RepositoryDataPort that reads from
// the Market Data tables directly (Roadmap Phase 1 monolith mode). When DSN
// is empty, the service falls back to an in-memory port suitable for tests
// and local exploration.
type MarketDataConfig struct {
	DSN          string `json:",optional"`
	Adjustment   string `json:",default=pre"`
	MaxOpenConns int    `json:",default=10"`
	MaxIdleConns int    `json:",default=2"`
}

type KafkaConfig struct {
	Brokers []string `json:",optional"`
}

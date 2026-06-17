package config

import "github.com/zeromicro/go-zero/zrpc"

// Config groups runtime configuration for the Market Data service.
type Config struct {
	zrpc.RpcServerConf

	HttpPort    int              `json:",default=8083"`
	MetricsPort int              `json:",default=9092"`

	Postgres   PostgresConfig   `json:",optional"`
	RedisCache RedisCacheConfig `json:",optional"`
	Kafka      KafkaConfig      `json:",optional"`
	Provider   ProviderConfig   `json:",optional"`
}

// PostgresConfig describes the PostgreSQL/TimescaleDB connection.
type PostgresConfig struct {
	DSN             string `json:",optional"`
	MaxOpenConns    int    `json:",default=20"`
	MaxIdleConns    int    `json:",default=5"`
	ConnMaxLifetime int    `json:",default=1800"` // seconds
	AutoMigrate     bool   `json:",default=true"`
}

// RedisCacheConfig describes the Redis connection used as a query cache.
type RedisCacheConfig struct {
	Host string `json:",default=localhost"`
	Port int    `json:",default=6379"`
	Pass string `json:",optional"`
	DB   int    `json:",default=0"`
	TTL  int    `json:",default=3600"`
}

// KafkaConfig describes the Kafka brokers used to publish market events.
type KafkaConfig struct {
	Brokers []string `json:",optional"`
}

// ProviderConfig selects and configures an upstream data provider.
type ProviderConfig struct {
	// Driver selects the provider implementation: "tushare" (default) or "fake"
	// for offline development.
	Driver string `json:",default=tushare"`

	// Endpoint overrides the upstream host (mainly for tests).
	Endpoint string `json:",optional"`

	// TokenEnv names the environment variable that holds the Tushare token.
	TokenEnv string `json:",default=TUSHARE_TOKEN"`
}

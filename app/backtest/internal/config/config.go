package config

import "github.com/zeromicro/go-zero/zrpc"

// Config is the Backtest Engine service configuration.
//
// All external dependencies are optional: when a section is omitted the
// service falls back to the in-memory adapter, which keeps the binary
// usable for local exploration / smoke tests without provisioning Postgres
// or Kafka. The only required block in production is MarketData / Formula.
type Config struct {
	zrpc.RpcServerConf

	HttpPort   int              `json:",default=8082"`
	Postgres   PostgresConfig   `json:",optional"`
	RedisCache RedisCacheConfig `json:",optional"`
	Kafka      KafkaConfig      `json:",optional"`
	MarketData MarketDataConfig `json:",optional"`
	Formula    FormulaConfig    `json:",optional"`
	Engine     EngineConfig     `json:",optional"`
	Queue      QueueConfig      `json:",optional"`
}

// PostgresConfig configures the Postgres connection used to persist jobs,
// orders, trades, snapshots, and reports.
type PostgresConfig struct {
	DSN          string `json:",optional"`
	MaxOpenConns int    `json:",default=10"`
	MaxIdleConns int    `json:",default=2"`
	// AutoMigrate triggers EnsureSchema at boot when true. Production
	// deployments should keep this false and apply migrations through a
	// dedicated tool, but for the MVP it lets a fresh database come up
	// without manual SQL.
	AutoMigrate  bool   `json:",default=true"`
}

// RedisCacheConfig configures the Redis client used for short-lived caches
// (job status, report payloads). Reserved for the upcoming caching layer.
type RedisCacheConfig struct {
	Host string `json:",default=localhost"`
	Port int    `json:",default=6379"`
	Pass string `json:",optional"`
	DB   int    `json:",default=0"`
	TTL  int    `json:",default=86400"`
}

// KafkaConfig configures the event publisher. When Brokers is empty the
// service uses the no-op publisher so MVP runs locally.
type KafkaConfig struct {
	Brokers []string `json:",optional"`
}

// MarketDataConfig points the engine at the in-process Market Data service
// or its DSN. The MVP shares the Postgres database with the Formula Engine,
// so DSN is the same Market Data DSN the Formula service consumes.
type MarketDataConfig struct {
	DSN          string `json:",optional"`
	Adjustment   string `json:",default=pre"`
	MaxOpenConns int    `json:",default=10"`
	MaxIdleConns int    `json:",default=2"`
}

// FormulaConfig configures how the engine talks to the Formula service.
// MVP runs Formula in-process; future deployments swap Mode to "rpc" and
// fill Endpoint with the gRPC target.
type FormulaConfig struct {
	Mode     string `json:",default=inprocess"`
	Endpoint string `json:",optional"`
}

// EngineConfig captures backtest-engine knobs surfaced through configuration
// rather than per-job.
type EngineConfig struct {
	AnnualisationDays int     `json:",default=252"`
	RiskFreeRate      float64 `json:",default=0"`
}

// QueueConfig controls the in-process async backtest pipeline. The MVP
// ships an in-memory channel-backed queue; durable adapters (Kafka,
// Redis Streams) will land alongside it without changing this surface.
type QueueConfig struct {
	// Enabled toggles the async submit path. When false the service
	// still accepts inline ?run=true / ?wait=true synchronous calls but
	// POST :id/run rejects with ErrQueueUnavailable so callers know to
	// flip the flag (or move to a deployment with the worker stack).
	Enabled bool `json:",default=true"`
	// Workers is the number of goroutines draining the queue. Each
	// worker runs one backtest at a time; tune this against the host's
	// CPU and the typical job runtime.
	Workers int `json:",default=2"`
	// Buffer caps the number of queued jobs in memory; Submit blocks
	// when the queue is saturated. Production deployments will replace
	// this with a durable backend.
	Buffer int `json:",default=256"`
	// JobTimeoutSeconds caps the runtime of a single job. Zero disables
	// the timeout; that is the safer default for the MVP because
	// backtests can legitimately take several minutes.
	JobTimeoutSeconds int `json:",default=0"`
}

package config

import "github.com/zeromicro/go-zero/zrpc"

// Config is the Strategy Service configuration.
//
// All external dependencies are optional: when a section is omitted the
// service falls back to the in-memory adapter, which keeps the binary
// usable for local exploration / smoke tests without provisioning
// Postgres or Kafka.
type Config struct {
	zrpc.RpcServerConf

	HttpPort     int                `json:",default=8083"`
	Postgres     PostgresConfig     `json:",optional"`
	Kafka        KafkaConfig        `json:",optional"`
	BacktestSync BacktestSyncConfig `json:",optional"`
}

// PostgresConfig configures the Postgres connection used to persist
// strategies, versions, and forks.
type PostgresConfig struct {
	DSN          string `json:",optional"`
	MaxOpenConns int    `json:",default=10"`
	MaxIdleConns int    `json:",default=2"`
	// AutoMigrate triggers EnsureSchema at boot when true. Production
	// deployments should keep this false and rely on a dedicated
	// migration tool, but the MVP lets a fresh database come up
	// without manual SQL.
	AutoMigrate bool `json:",default=true"`
}

// KafkaConfig configures the event publisher. When Brokers is empty the
// service uses the no-op publisher so MVP runs locally.
type KafkaConfig struct {
	Brokers []string `json:",optional"`
}


// BacktestSyncConfig configures the Backtest events consumer. When
// Enabled is true the service subscribes to backtest-events and flips
// strategies into BACKTESTED on every BacktestFinished. Disabled by
// default so a vanilla MVP boot does not require Kafka.
type BacktestSyncConfig struct {
	Enabled bool     `json:",default=false"`
	Brokers []string `json:",optional"`
	Topic   string   `json:",default=backtest-events"`
	GroupID string   `json:",default=strategy-backtest-sync"`
}

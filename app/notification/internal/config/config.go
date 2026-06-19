// Package config defines the Notification Service configuration. The
// shape mirrors the User Service so operators can copy / paste the
// patterns that already exist for Postgres, Kafka, and cross-service
// consumers.
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config is the Notification Service configuration.
type Config struct {
	zrpc.RpcServerConf

	HttpPort     int             `json:",default=8089"`
	Postgres     PostgresConfig  `json:",optional"`
	UserSync     EventSyncConfig `json:",optional"`
	StrategySync EventSyncConfig `json:",optional"`
	BacktestSync EventSyncConfig `json:",optional"`
}

// PostgresConfig configures the Postgres connection used to persist
// notifications, preferences, and subscriptions.
type PostgresConfig struct {
	DSN          string `json:",optional"`
	MaxOpenConns int    `json:",default=10"`
	MaxIdleConns int    `json:",default=2"`
	// AutoMigrate triggers EnsureSchema at boot when true. Production
	// deployments should rely on the platform migration tool.
	AutoMigrate bool `json:",default=true"`
}

// EventSyncConfig configures a cross-service Kafka consumer.
type EventSyncConfig struct {
	Enabled bool     `json:",default=false"`
	Brokers []string `json:",optional"`
	Topic   string   `json:",optional"`
	GroupID string   `json:",optional"`
}

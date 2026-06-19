package config

import "github.com/zeromicro/go-zero/zrpc"

// Config is the User Service configuration.
//
// All external dependencies are optional: when a section is omitted the
// service falls back to the in-memory adapter, which keeps the binary
// usable for local exploration / smoke tests without provisioning
// Postgres or Kafka.
type Config struct {
	zrpc.RpcServerConf

	HttpPort     int                `json:",default=8087"`
	Postgres     PostgresConfig     `json:",optional"`
	Kafka        KafkaConfig        `json:",optional"`
	Token        TokenConfig        `json:",optional"`
	Password     PasswordConfig     `json:",optional"`
	StrategySync EventSyncConfig    `json:",optional"`
	BacktestSync EventSyncConfig    `json:",optional"`
}

// EventSyncConfig configures a cross-service Kafka consumer. When
// Enabled is true the service subscribes to the topic and updates the
// activity counters on every relevant event. Disabled by default so a
// vanilla MVP boot does not require Kafka.
type EventSyncConfig struct {
	Enabled bool     `json:",default=false"`
	Brokers []string `json:",optional"`
	Topic   string   `json:",optional"`
	GroupID string   `json:",optional"`
}

// PostgresConfig configures the Postgres connection used to persist
// users and follow rows.
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

// KafkaConfig configures the event publisher. Empty Brokers selects
// the no-op publisher.
type KafkaConfig struct {
	Brokers []string `json:",optional"`
}

// TokenConfig configures the JWT issuer. Production deployments
// should populate Keys (one row per id+secret) so secret rotation is
// just an additive config change. Secret remains for backwards
// compatibility; when set without Keys it is wrapped in a single-key
// KeySet at boot.
type TokenConfig struct {
	Secret            string             `json:",optional"`
	ActiveKeyID       string             `json:",optional"`
	Keys              []SigningKeyConfig `json:",optional"`
	Issuer            string             `json:",default=quantlab.user"`
	AccessTTLSeconds  int                `json:",default=1800"`
	RefreshTTLSeconds int                `json:",default=1209600"`
}

// SigningKeyConfig is one row of the rotation-aware key set.
type SigningKeyConfig struct {
	ID     string `json:",optional"`
	Secret string `json:""`
}

// PasswordConfig configures the bcrypt hasher.
type PasswordConfig struct {
	BcryptCost int `json:",default=10"`
}

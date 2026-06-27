// Package config holds the API Gateway configuration. The gateway is
// the single REST entry point for the frontend; it dials every backend
// service over gRPC and owns the shared JWT verifier so auth is
// resolved once and propagated downstream via gRPC metadata.
package config

// Config is the API Gateway configuration.
type Config struct {
	// HttpPort is the public REST port the frontend connects to.
	HttpPort int `json:",default=8000"`

	// Token mirrors the User Service signing key set so the gateway
	// can verify access tokens issued by the User Service without an
	// extra round-trip. The two configs must stay in sync.
	Token TokenConfig `json:",optional"`

	// Per-service gRPC endpoints. Each entry is a list of host:port
	// targets; the gateway dials them directly (no Etcd in the MVP).
	// Missing / empty endpoints leave the corresponding gRPC client
	// nil so the gateway boots even when a service is not running yet.
	User         ServiceConfig `json:",optional"`
	Strategy     ServiceConfig `json:",optional"`
	Formula      ServiceConfig `json:",optional"`
	// FormulaHTTP is the formula service's own HTTP address (separate
	// from its gRPC port). The gateway proxies /formulas/evaluate here
	// because Evaluate is not in the gRPC proto — it lives only on the
	// formula service's HTTP surface.
	FormulaHTTP  string        `json:",optional"`
	Backtest     ServiceConfig `json:",optional"`
	Market       ServiceConfig `json:",optional"`
	Ranking      ServiceConfig `json:",optional"`
	Portfolio    ServiceConfig `json:",optional"`
	Community    ServiceConfig `json:",optional"`
	AI           ServiceConfig `json:",optional"`
	Billing      ServiceConfig `json:",optional"`
	Notification ServiceConfig `json:",optional"`
}

// ServiceConfig configures one downstream gRPC client.
type ServiceConfig struct {
	Endpoints []string `json:",optional"`
	Timeout   int64    `json:",default=2000"` // milliseconds
	NonBlock  bool     `json:",default=true"`
}

// TokenConfig mirrors app/user/internal/config.TokenConfig so the
// gateway verifies tokens with the same keys the User Service signs
// them with.
type TokenConfig struct {
	ActiveKeyID string             `json:",optional"`
	Keys        []SigningKeyConfig `json:",optional"`
	Issuer      string             `json:",default=quantlab.user"`
}

// SigningKeyConfig is one row of the rotation-aware key set.
type SigningKeyConfig struct {
	ID     string `json:",optional"`
	Secret string `json:""`
}

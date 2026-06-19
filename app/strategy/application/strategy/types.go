package strategy

import (
	"github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// CreateRequest is the payload for the Create use case.
type CreateRequest struct {
	AuthorID    int64
	Title       string
	Description string
	Category    string
	Tags        []string
	Visibility  valueobject.Visibility
}

// CreateResult bundles the aggregate after persistence.
type CreateResult struct {
	Strategy *strategy.Strategy
}

// UpdateRequest is the payload for metadata edits. Pointer fields keep
// "leave alone" distinct from "set empty"; this matters most for tags.
type UpdateRequest struct {
	StrategyID  int64
	CallerID    int64
	Title       *string
	Description *string
	Category    *string
	Tags        *[]string
	Visibility  *valueobject.Visibility
}

// CreateVersionRequest carries a fresh snapshot of the strategy body.
// AutoNumber is always true in MVP: the application layer derives the
// next semver-ish number; explicit numbering arrives with the editor's
// "save as" feature.
type CreateVersionRequest struct {
	StrategyID    int64
	CallerID      int64
	FormulaText   string
	BuyRule       string
	SellRule      string
	RiskRule      string
	PositionRule  string
	RebalanceRule string
	ChangeLog     string
}

// CreateVersionResult bundles the new version + the updated strategy.
type CreateVersionResult struct {
	Strategy *strategy.Strategy
	Version  *version.StrategyVersion
}

// PublishRequest publishes a specific version. VersionID is optional;
// when zero the application layer publishes the strategy's current
// version, which is what the MVP UI does today.
type PublishRequest struct {
	StrategyID int64
	CallerID   int64
	VersionID  int64
}

// ArchiveRequest takes a strategy off the public surface.
type ArchiveRequest struct {
	StrategyID int64
	CallerID   int64
}

// ForkRequest copies a public strategy into the caller's namespace.
// Title may override the source title; empty falls back to the source.
type ForkRequest struct {
	SourceStrategyID int64
	CallerID         int64
	Title            string
}

// ForkResult bundles the new strategy + the seeded version.
type ForkResult struct {
	Strategy *strategy.Strategy
	Version  *version.StrategyVersion
}

// ListQuery filters the list / search use cases.
type ListQuery struct {
	AuthorID   int64
	Status     valueobject.LifecycleStatus
	Visibility valueobject.Visibility
	Category   string
	Tag        string
	Keyword    string
	Sort       string
	Limit      int
	Offset     int
}

// MarkBacktestedRequest is the contract the Backtest Engine consumer
// uses to flip the strategy into Backtested. We keep it as a separate
// use case so future event-driven wiring (Kafka consumer in the
// strategy service) calls the same code path the in-process tests do.
type MarkBacktestedRequest struct {
	StrategyID int64
}

// Package strategysync defines the contract Backtest uses to react to
// Strategy lifecycle events. The package keeps the schema decoupled
// from the Strategy service's domain types: the consumer decodes a
// generic envelope, fans the payload into one of the handler methods,
// and the handler decides what to do (e.g. trigger a baseline
// backtest). Splitting it this way means the consumer adapter never
// reaches into the Backtest application service directly, and the
// handler stays unit-testable without spinning up Kafka.
package strategysync

import (
	"context"
	"time"
)

// EventType lists the canonical Strategy events Backtest cares about.
// We re-declare the strings here (instead of importing the strategy
// service's domain package) so Backtest does not pick up a build-time
// dependency on a sibling service. The two services agree on the
// envelope through the platform Event Specification, not through code.
type EventType string

const (
	EventStrategyPublished      EventType = "StrategyPublished"
	EventStrategyVersionCreated EventType = "StrategyVersionCreated"
)

// Envelope mirrors the common Event shape every QuantLab service emits.
// Only the fields Backtest needs are decoded; unknown fields are
// preserved by json.RawMessage on Payload so tests can assert on the
// raw bytes without re-marshalling.
type Envelope struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Producer      string    `json:"producer"`
	// Payload is the inner event-specific object. The consumer decodes
	// it into the right struct based on EventType before calling the
	// handler.
	Payload map[string]any `json:"payload"`
}

// PublishedPayload is the inner shape of StrategyPublished. The Strategy
// service writes strategy_id / version_id; AuthorID is included so
// downstream consumers can attribute the job without an extra lookup.
type PublishedPayload struct {
	StrategyID int64 `json:"strategy_id"`
	VersionID  int64 `json:"version_id"`
	AuthorID   int64 `json:"author_id,omitempty"`
}

// VersionCreatedPayload is the inner shape of StrategyVersionCreated.
type VersionCreatedPayload struct {
	StrategyID int64  `json:"strategy_id"`
	VersionID  int64  `json:"version_id"`
	VersionNo  string `json:"version_no,omitempty"`
}

// Handler reacts to decoded Strategy events. The consumer adapter calls
// exactly one method per event; methods return an error so the consumer
// can decide whether to ack or retry. Implementations should be safe to
// call concurrently.
type Handler interface {
	OnPublished(ctx context.Context, env Envelope, payload PublishedPayload) error
	OnVersionCreated(ctx context.Context, env Envelope, payload VersionCreatedPayload) error
}

// StrategySnapshot is the minimal projection Backtest needs to build a
// baseline run. The Resolver below returns one of these so the handler
// stays independent of the Strategy service's protobuf types.
type StrategySnapshot struct {
	StrategyID  int64
	VersionID   int64
	VersionNo   string
	AuthorID    int64
	Title       string
	FormulaText string
}

// Resolver loads a StrategySnapshot for a given (strategy, version)
// pair. The default implementation in infrastructure wraps the
// strategy service's gRPC client; an in-memory implementation is
// shipped alongside for tests.
type Resolver interface {
	Resolve(ctx context.Context, strategyID, versionID int64) (*StrategySnapshot, error)
}

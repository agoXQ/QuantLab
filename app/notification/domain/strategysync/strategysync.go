// Package strategysync defines the contract the Notification service
// uses to react to Strategy lifecycle events. The shape mirrors how
// User consumes Strategy events: a small Envelope + payload pair per
// event type, plus a Handler interface the consumer dispatches to.
//
// Re-declaring the JSON-level event names instead of importing the
// Strategy service's domain types keeps Notification free of build-
// time coupling. The platform Event Specification is the integration
// contract; a wire-compat test pins the two ends together.
package strategysync

import (
	"context"
	"time"
)

// EventType lists the canonical Strategy events Notification cares
// about. Created stays in for symmetry with the User service; the
// MVP fan-out only acts on Published / Forked.
type EventType string

const (
	EventStrategyCreated   EventType = "StrategyCreated"
	EventStrategyPublished EventType = "StrategyPublished"
	EventStrategyForked    EventType = "StrategyForked"
)

// Envelope mirrors the common Event shape every QuantLab service emits.
type Envelope struct {
	EventID       string         `json:"event_id"`
	EventType     EventType      `json:"event_type"`
	EventVersion  string         `json:"event_version"`
	OccurredAt    time.Time      `json:"occurred_at"`
	AggregateType string         `json:"aggregate_type"`
	AggregateID   string         `json:"aggregate_id"`
	Producer      string         `json:"producer"`
	Payload       map[string]any `json:"payload"`
}

// CreatedPayload mirrors strategy/domain/event.StrategyCreatedPayload.
type CreatedPayload struct {
	StrategyID int64  `json:"strategy_id"`
	AuthorID   int64  `json:"author_id,omitempty"`
	Title      string `json:"title,omitempty"`
}

// PublishedPayload mirrors strategy/domain/event.StrategyPublishedPayload.
type PublishedPayload struct {
	StrategyID int64 `json:"strategy_id"`
	VersionID  int64 `json:"version_id"`
	AuthorID   int64 `json:"author_id,omitempty"`
}

// ForkedPayload mirrors strategy/domain/event.StrategyForkedPayload.
type ForkedPayload struct {
	SourceStrategyID int64 `json:"source_strategy_id"`
	TargetStrategyID int64 `json:"target_strategy_id"`
	CreatorID        int64 `json:"creator_id,omitempty"`
}

// Handler reacts to decoded Strategy events.
type Handler interface {
	OnCreated(ctx context.Context, env Envelope, payload CreatedPayload) error
	OnPublished(ctx context.Context, env Envelope, payload PublishedPayload) error
	OnForked(ctx context.Context, env Envelope, payload ForkedPayload) error
}

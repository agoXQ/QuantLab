// Package strategysync defines the contract the User service uses to
// react to Strategy lifecycle events. The shape mirrors how Backtest
// consumes Strategy events: a small Envelope + payload pair per event
// type, plus a Handler interface the consumer dispatches to.
//
// Re-declaring the JSON-level event names instead of importing the
// Strategy service's domain types keeps the User service free of
// build-time coupling. The platform Event Specification is the
// integration contract; a wire-compat test pins the two ends together.
package strategysync

import (
	"context"
	"time"
)

// EventType lists the canonical Strategy events User cares about.
type EventType string

const (
	EventStrategyCreated  EventType = "StrategyCreated"
	EventStrategyArchived EventType = "StrategyArchived"
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

// CreatedPayload is the inner shape of StrategyCreated; the User
// service only needs the author id to bump the strategy counter.
type CreatedPayload struct {
	StrategyID int64  `json:"strategy_id"`
	AuthorID   int64  `json:"author_id,omitempty"`
	Title      string `json:"title,omitempty"`
}

// ArchivedPayload is the inner shape of StrategyArchived; we use it
// to decrement the counter so the profile view stays accurate.
type ArchivedPayload struct {
	StrategyID int64 `json:"strategy_id"`
	AuthorID   int64 `json:"author_id,omitempty"`
}

// Handler reacts to decoded Strategy events.
type Handler interface {
	OnCreated(ctx context.Context, env Envelope, payload CreatedPayload) error
	OnArchived(ctx context.Context, env Envelope, payload ArchivedPayload) error
}

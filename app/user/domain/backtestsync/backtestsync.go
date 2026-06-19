// Package backtestsync defines the contract the User service uses to
// react to Backtest lifecycle events. Counters track only finished
// runs; failed / cancelled jobs are intentionally ignored to keep the
// "backtest_count" semantically meaningful for profile pages.
package backtestsync

import (
	"context"
	"time"
)

// EventType lists the canonical Backtest events User reacts to.
type EventType string

const (
	EventBacktestFinished EventType = "BacktestFinished"
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

// FinishedPayload is the inner shape of BacktestFinished. The
// canonical Backtest service writes more fields; the User service
// only needs the user id to bump the activity counter.
type FinishedPayload struct {
	JobID      int64 `json:"job_id"`
	UserID     int64 `json:"user_id,omitempty"`
	StrategyID int64 `json:"strategy_id,omitempty"`
}

// Handler reacts to decoded Backtest events.
type Handler interface {
	OnFinished(ctx context.Context, env Envelope, payload FinishedPayload) error
}

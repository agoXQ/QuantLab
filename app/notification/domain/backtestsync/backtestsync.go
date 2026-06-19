// Package backtestsync defines the contract the Notification service
// uses to react to Backtest lifecycle events. The shape mirrors how
// User consumes Backtest events: a small Envelope + payload pair per
// event type, plus a Handler interface the consumer dispatches to.
package backtestsync

import (
	"context"
	"time"
)

// EventType lists the canonical Backtest events Notification cares
// about. The MVP only acts on Finished / Failed; the rest are stubbed
// for future expansion.
type EventType string

const (
	EventBacktestFinished EventType = "BacktestFinished"
	EventBacktestFailed   EventType = "BacktestFailed"
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

// FinishedPayload mirrors backtest/domain/event.BacktestFinishedPayload.
type FinishedPayload struct {
	JobID        int64   `json:"job_id"`
	UserID       int64   `json:"user_id,omitempty"`
	StrategyID   int64   `json:"strategy_id,omitempty"`
	TotalReturn  float64 `json:"total_return"`
	AnnualReturn float64 `json:"annual_return"`
	Sharpe       float64 `json:"sharpe_ratio"`
	MaxDrawdown  float64 `json:"max_drawdown"`
}

// FailedPayload mirrors backtest/domain/event.BacktestFailedPayload.
type FailedPayload struct {
	JobID  int64  `json:"job_id"`
	UserID int64  `json:"user_id,omitempty"`
	Reason string `json:"reason"`
}

// Handler reacts to decoded Backtest events.
type Handler interface {
	OnFinished(ctx context.Context, env Envelope, payload FinishedPayload) error
	OnFailed(ctx context.Context, env Envelope, payload FailedPayload) error
}

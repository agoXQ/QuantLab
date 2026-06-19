// Package backtestsync defines the contract the Strategy service uses
// to react to Backtest lifecycle events. The shape mirrors how
// Backtest consumes Strategy events: a small Envelope + payload pair
// per event type, plus a Handler interface the consumer dispatches to.
//
// The package re-declares the JSON-level event names instead of
// importing the Backtest service's domain types so the two services
// stay free of build-time coupling. The platform Event Specification
// is the integration contract; a wire-compat test pins the two ends
// together.
package backtestsync

import (
	"context"
	"time"
)

// EventType lists the canonical Backtest events Strategy reacts to.
type EventType string

const (
	EventBacktestFinished EventType = "BacktestFinished"
)

// Envelope mirrors the common Event shape every QuantLab service emits.
type Envelope struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Producer      string    `json:"producer"`
	Payload       map[string]any `json:"payload"`
}

// FinishedPayload is the inner shape of BacktestFinished. Backtest
// reports more fields (returns, drawdown), but Strategy only needs the
// strategy id to advance the lifecycle. The remaining fields stay
// available through Envelope.Payload for any future consumer that
// wants them.
type FinishedPayload struct {
	JobID        int64   `json:"job_id"`
	StrategyID   int64   `json:"strategy_id,omitempty"`
	TotalReturn  float64 `json:"total_return,omitempty"`
	AnnualReturn float64 `json:"annual_return,omitempty"`
	Sharpe       float64 `json:"sharpe_ratio,omitempty"`
	MaxDrawdown  float64 `json:"max_drawdown,omitempty"`
}

// Handler reacts to decoded Backtest events. Implementations should be
// safe to call concurrently. Methods return an error so the consumer
// can decide whether to ack or retry.
type Handler interface {
	OnFinished(ctx context.Context, env Envelope, payload FinishedPayload) error
}

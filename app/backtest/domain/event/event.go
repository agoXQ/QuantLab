// Package event defines domain events emitted by the Backtest Engine.
package event

import (
	"context"
	"time"
)

// EventType lists the canonical Backtest events.
type EventType string

const (
	EventBacktestCreated  EventType = "BacktestCreated"
	EventBacktestStarted  EventType = "BacktestStarted"
	EventBacktestFinished EventType = "BacktestFinished"
	EventBacktestFailed   EventType = "BacktestFailed"
)

const (
	AggregateTypeBacktest = "BACKTEST_JOB"
	ProducerBacktest      = "backtest-engine"
	TopicBacktestEvents   = "backtest-events"
	EventVersionV1        = "1.0"
)

// Event is the canonical envelope used by every Backtest event.
//
// Field naming aligns with the platform Event Specification so the same
// consumers (Ranking / Notification / AI) can decode all upstream services.
type Event struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Producer      string    `json:"producer"`
	Payload       any       `json:"payload"`
}

// BacktestCreatedPayload is the payload for BacktestCreated.
type BacktestCreatedPayload struct {
	JobID      int64  `json:"job_id"`
	UserID     int64  `json:"user_id,omitempty"`
	StrategyID int64  `json:"strategy_id,omitempty"`
	Formula    string `json:"formula"`
}

// BacktestStartedPayload is the payload for BacktestStarted.
type BacktestStartedPayload struct {
	JobID int64 `json:"job_id"`
}

// BacktestFinishedPayload is the payload for BacktestFinished.
type BacktestFinishedPayload struct {
	JobID        int64   `json:"job_id"`
	StrategyID   int64   `json:"strategy_id,omitempty"`
	TotalReturn  float64 `json:"total_return"`
	AnnualReturn float64 `json:"annual_return"`
	Sharpe       float64 `json:"sharpe_ratio"`
	MaxDrawdown  float64 `json:"max_drawdown"`
}

// BacktestFailedPayload is the payload for BacktestFailed.
type BacktestFailedPayload struct {
	JobID  int64  `json:"job_id"`
	Reason string `json:"reason"`
}

// Publisher publishes Backtest events.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

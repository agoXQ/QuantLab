package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appBacktestSync "github.com/agoXQ/QuantLab/app/user/application/backtestsync"
	appStrategySync "github.com/agoXQ/QuantLab/app/user/application/strategysync"
	infraBacktestSync "github.com/agoXQ/QuantLab/app/user/infrastructure/backtestsync"
	infraStrategySync "github.com/agoXQ/QuantLab/app/user/infrastructure/strategysync"
)

// TestStrategySync_DispatchEnvelope drives a minted strategy envelope
// through the consumer dispatcher and confirms the user's strategy
// counter ticks up.
func TestStrategySync_DispatchEnvelope(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	id := registerHelper(t, fx, "alice", "alice@example.com")

	handler := appStrategySync.NewCounterHandler(fx.svc)
	env := mintStrategyEnvelope("StrategyCreated", id, time.Now())
	if err := infraStrategySync.Dispatch(ctx, handler, env); err != nil {
		t.Fatalf("dispatch created: %v", err)
	}
	prof, err := fx.svc.GetProfile(ctx, id)
	if err != nil || prof.StrategyCount != 1 {
		t.Fatalf("after StrategyCreated expected count=1 err=%v prof=%+v", err, prof)
	}

	// Replay the same envelope: counter must stay at 1 (idempotent).
	if err := infraStrategySync.Dispatch(ctx, handler, env); err != nil {
		t.Fatalf("replay: %v", err)
	}
	prof, _ = fx.svc.GetProfile(ctx, id)
	if prof.StrategyCount != 1 {
		t.Fatalf("expected idempotent count=1, got %d", prof.StrategyCount)
	}

	archived := mintStrategyEnvelope("StrategyArchived", id, time.Now())
	if err := infraStrategySync.Dispatch(ctx, handler, archived); err != nil {
		t.Fatalf("dispatch archived: %v", err)
	}
	prof, _ = fx.svc.GetProfile(ctx, id)
	if prof.StrategyCount != 0 {
		t.Fatalf("after StrategyArchived expected count=0, got %d", prof.StrategyCount)
	}
}

// TestBacktestSync_DispatchEnvelope mirrors the strategy test for the
// backtest counter.
func TestBacktestSync_DispatchEnvelope(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	id := registerHelper(t, fx, "alice", "alice@example.com")

	handler := appBacktestSync.NewCounterHandler(fx.svc)
	env := mintBacktestEnvelope(101, id)
	if err := infraBacktestSync.Dispatch(ctx, handler, env); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	prof, err := fx.svc.GetProfile(ctx, id)
	if err != nil || prof.BacktestCount != 1 {
		t.Fatalf("expected count=1 err=%v prof=%+v", err, prof)
	}

	// Idempotent on job id.
	if err := infraBacktestSync.Dispatch(ctx, handler, env); err != nil {
		t.Fatalf("replay: %v", err)
	}
	prof, _ = fx.svc.GetProfile(ctx, id)
	if prof.BacktestCount != 1 {
		t.Fatalf("expected idempotent count=1, got %d", prof.BacktestCount)
	}
}

func mintStrategyEnvelope(eventType string, authorID int64, occurred time.Time) []byte {
	type envelope struct {
		EventID       string         `json:"event_id"`
		EventType     string         `json:"event_type"`
		EventVersion  string         `json:"event_version"`
		OccurredAt    time.Time      `json:"occurred_at"`
		AggregateType string         `json:"aggregate_type"`
		AggregateID   string         `json:"aggregate_id"`
		Producer      string         `json:"producer"`
		Payload       map[string]any `json:"payload"`
	}
	id := eventType + "-" + occurred.Format(time.RFC3339Nano)
	env := envelope{
		EventID:       id,
		EventType:     eventType,
		EventVersion:  "1.0",
		OccurredAt:    occurred,
		AggregateType: "STRATEGY",
		AggregateID:   "1",
		Producer:      "strategy-service",
		Payload: map[string]any{
			"strategy_id": 1,
			"author_id":   authorID,
			"title":       "Test",
		},
	}
	buf, _ := json.Marshal(env)
	return buf
}

func mintBacktestEnvelope(jobID, userID int64) []byte {
	type envelope struct {
		EventID       string         `json:"event_id"`
		EventType     string         `json:"event_type"`
		EventVersion  string         `json:"event_version"`
		OccurredAt    time.Time      `json:"occurred_at"`
		AggregateType string         `json:"aggregate_type"`
		AggregateID   string         `json:"aggregate_id"`
		Producer      string         `json:"producer"`
		Payload       map[string]any `json:"payload"`
	}
	env := envelope{
		EventID:       "bt-finished",
		EventType:     "BacktestFinished",
		EventVersion:  "1.0",
		OccurredAt:    time.Now(),
		AggregateType: "BACKTEST_JOB",
		AggregateID:   "101",
		Producer:      "backtest-engine",
		Payload: map[string]any{
			"job_id":  jobID,
			"user_id": userID,
		},
	}
	buf, _ := json.Marshal(env)
	return buf
}

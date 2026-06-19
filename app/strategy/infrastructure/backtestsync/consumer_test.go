package backtestsync

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	btEvent "github.com/agoXQ/QuantLab/app/backtest/domain/event"

	domsync "github.com/agoXQ/QuantLab/app/strategy/domain/backtestsync"
)

// recordingHandler captures every call so we can assert on routing.
type recordingHandler struct {
	finished []domsync.FinishedPayload
}

func (r *recordingHandler) OnFinished(_ context.Context, _ domsync.Envelope, p domsync.FinishedPayload) error {
	r.finished = append(r.finished, p)
	return nil
}

func TestDispatch_RoutesFinished(t *testing.T) {
	handler := &recordingHandler{}
	env := domsync.Envelope{
		EventID:   "evt-1",
		EventType: domsync.EventBacktestFinished,
		Payload: map[string]any{
			"job_id":      float64(101),
			"strategy_id": float64(7),
		},
	}
	raw, _ := json.Marshal(env)
	if err := Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(handler.finished) != 1 || handler.finished[0].JobID != 101 || handler.finished[0].StrategyID != 7 {
		t.Fatalf("routing mismatch: %+v", handler.finished)
	}
}

func TestDispatch_IgnoresUnknownEventType(t *testing.T) {
	handler := &recordingHandler{}
	other, _ := json.Marshal(domsync.Envelope{EventType: "BacktestStarted", Payload: map[string]any{"job_id": 1}})
	if err := Dispatch(context.Background(), handler, other); err != nil {
		t.Fatalf("dispatch ignored type: %v", err)
	}
	if len(handler.finished) != 0 {
		t.Fatalf("expected no routing for unknown type, got %d", len(handler.finished))
	}
}

func TestDispatch_RejectsMalformedJSON(t *testing.T) {
	handler := &recordingHandler{}
	if err := Dispatch(context.Background(), handler, []byte("not json")); err == nil {
		t.Fatal("expected decode error")
	}
}

// TestEnvelopeRoundtripFromBacktestService guards the wire compatibility
// between the Backtest service's emitted envelope and the Strategy
// consumer's decoder. If either side drifts (renamed JSON tag, type
// change), this test fails immediately.
func TestEnvelopeRoundtripFromBacktestService(t *testing.T) {
	handler := &recordingHandler{}
	finished := btEvent.Event{
		EventID:       "evt-fin-1",
		EventType:     btEvent.EventBacktestFinished,
		EventVersion:  btEvent.EventVersionV1,
		OccurredAt:    time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC),
		AggregateType: btEvent.AggregateTypeBacktest,
		AggregateID:   "101",
		Producer:      btEvent.ProducerBacktest,
		Payload: btEvent.BacktestFinishedPayload{
			JobID:        101,
			StrategyID:   7,
			TotalReturn:  0.42,
			AnnualReturn: 0.18,
			Sharpe:       1.5,
			MaxDrawdown:  -0.12,
		},
	}
	raw, err := json.Marshal(finished)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(handler.finished) != 1 {
		t.Fatalf("expected 1 routed event, got %d", len(handler.finished))
	}
	got := handler.finished[0]
	if got.JobID != 101 || got.StrategyID != 7 {
		t.Fatalf("ids roundtrip mismatch: %+v", got)
	}
	if got.TotalReturn == 0 || got.Sharpe == 0 {
		t.Fatalf("metrics dropped during roundtrip: %+v", got)
	}
}

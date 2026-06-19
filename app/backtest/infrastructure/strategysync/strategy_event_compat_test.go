package strategysync

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	stratEvent "github.com/agoXQ/QuantLab/app/strategy/domain/event"

	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
)

// roundtripHandler counts how many events come through.
type roundtripHandler struct {
	publishedCount int
	versionCount   int
	lastPub        domsync.PublishedPayload
	lastVer        domsync.VersionCreatedPayload
}

func (r *roundtripHandler) OnPublished(_ context.Context, _ domsync.Envelope, p domsync.PublishedPayload) error {
	r.publishedCount++
	r.lastPub = p
	return nil
}

func (r *roundtripHandler) OnVersionCreated(_ context.Context, _ domsync.Envelope, p domsync.VersionCreatedPayload) error {
	r.versionCount++
	r.lastVer = p
	return nil
}

// TestEnvelopeRoundtripFromStrategyService guards the wire compatibility
// between the Strategy service's emitted envelope and the Backtest
// consumer's decoder. If either side drifts (renamed JSON tag, type
// change), this test fails immediately.
func TestEnvelopeRoundtripFromStrategyService(t *testing.T) {
	handler := &roundtripHandler{}

	// Build the same envelope shape the Strategy service publishes.
	pubEvent := stratEvent.Event{
		EventID:       "evt-pub-1",
		EventType:     stratEvent.EventStrategyPublished,
		EventVersion:  stratEvent.EventVersionV1,
		OccurredAt:    time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC),
		AggregateType: stratEvent.AggregateTypeStrategy,
		AggregateID:   "7",
		Producer:      stratEvent.ProducerStrategy,
		Payload: stratEvent.StrategyPublishedPayload{
			StrategyID: 7,
			VersionID:  3,
			AuthorID:   42,
		},
	}
	raw, err := json.Marshal(pubEvent)
	if err != nil {
		t.Fatalf("marshal published: %v", err)
	}
	if err := Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch published: %v", err)
	}
	if handler.publishedCount != 1 {
		t.Fatalf("expected 1 published call, got %d", handler.publishedCount)
	}
	if handler.lastPub.StrategyID != 7 || handler.lastPub.VersionID != 3 || handler.lastPub.AuthorID != 42 {
		t.Fatalf("payload roundtrip mismatch: %+v", handler.lastPub)
	}

	verEvent := stratEvent.Event{
		EventID:      "evt-ver-1",
		EventType:    stratEvent.EventStrategyVersionCreated,
		EventVersion: stratEvent.EventVersionV1,
		Payload: stratEvent.StrategyVersionCreatedPayload{
			StrategyID: 7,
			VersionID:  4,
			VersionNo:  "v4",
		},
	}
	raw2, _ := json.Marshal(verEvent)
	if err := Dispatch(context.Background(), handler, raw2); err != nil {
		t.Fatalf("dispatch version-created: %v", err)
	}
	if handler.versionCount != 1 || handler.lastVer.VersionNo != "v4" {
		t.Fatalf("version-created mismatch: count=%d payload=%+v", handler.versionCount, handler.lastVer)
	}
}

package strategysync

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
)

// recordingHandler captures every call so we can assert on routing.
type recordingHandler struct {
	published []domsync.PublishedPayload
	versions  []domsync.VersionCreatedPayload
	failNext  error
}

func (r *recordingHandler) OnPublished(_ context.Context, _ domsync.Envelope, p domsync.PublishedPayload) error {
	r.published = append(r.published, p)
	if err := r.failNext; err != nil {
		r.failNext = nil
		return err
	}
	return nil
}

func (r *recordingHandler) OnVersionCreated(_ context.Context, _ domsync.Envelope, p domsync.VersionCreatedPayload) error {
	r.versions = append(r.versions, p)
	return nil
}

func TestDispatch_RoutesPublishedAndVersionCreated(t *testing.T) {
	handler := &recordingHandler{}
	pubEnv := domsync.Envelope{
		EventID:      "evt-1",
		EventType:    domsync.EventStrategyPublished,
		EventVersion: "1.0",
		OccurredAt:   time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC),
		Payload: map[string]any{
			"strategy_id": float64(7),
			"version_id":  float64(3),
			"author_id":   float64(42),
		},
	}
	pubBytes, _ := json.Marshal(pubEnv)

	verEnv := domsync.Envelope{
		EventID:   "evt-2",
		EventType: domsync.EventStrategyVersionCreated,
		Payload: map[string]any{
			"strategy_id": float64(7),
			"version_id":  float64(4),
			"version_no":  "v4",
		},
	}
	verBytes, _ := json.Marshal(verEnv)

	if err := Dispatch(context.Background(), handler, pubBytes); err != nil {
		t.Fatalf("dispatch published: %v", err)
	}
	if err := Dispatch(context.Background(), handler, verBytes); err != nil {
		t.Fatalf("dispatch version-created: %v", err)
	}
	if len(handler.published) != 1 || handler.published[0].StrategyID != 7 || handler.published[0].VersionID != 3 || handler.published[0].AuthorID != 42 {
		t.Fatalf("published routing mismatch: %+v", handler.published)
	}
	if len(handler.versions) != 1 || handler.versions[0].VersionNo != "v4" {
		t.Fatalf("version-created routing mismatch: %+v", handler.versions)
	}
}

func TestDispatch_IgnoresUnknownEventType(t *testing.T) {
	handler := &recordingHandler{}
	other, _ := json.Marshal(domsync.Envelope{EventType: "StrategyArchived", Payload: map[string]any{"strategy_id": 1}})
	if err := Dispatch(context.Background(), handler, other); err != nil {
		t.Fatalf("dispatch ignored type: %v", err)
	}
	if len(handler.published) != 0 || len(handler.versions) != 0 {
		t.Fatalf("expected no routing for unknown type, got published=%d versions=%d", len(handler.published), len(handler.versions))
	}
}

func TestDispatch_PropagatesHandlerError(t *testing.T) {
	want := errors.New("boom")
	handler := &recordingHandler{failNext: want}
	pubEnv := domsync.Envelope{
		EventType: domsync.EventStrategyPublished,
		Payload: map[string]any{
			"strategy_id": float64(1),
			"version_id":  float64(1),
		},
	}
	raw, _ := json.Marshal(pubEnv)
	if err := Dispatch(context.Background(), handler, raw); !errors.Is(err, want) {
		t.Fatalf("expected handler error to propagate, got %v", err)
	}
}

func TestDispatch_RejectsMalformedJSON(t *testing.T) {
	handler := &recordingHandler{}
	if err := Dispatch(context.Background(), handler, []byte("not json")); err == nil {
		t.Fatal("expected decode error")
	}
}

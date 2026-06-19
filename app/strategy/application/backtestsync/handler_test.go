package backtestsync

import (
	"context"
	"errors"
	"testing"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domsync "github.com/agoXQ/QuantLab/app/strategy/domain/backtestsync"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// fakeStrategyService is the minimal stub the handler needs. We only
// care about MarkBacktested; everything else returns zero values.
type fakeStrategyService struct {
	calls   []int64
	failNext error
}

func (f *fakeStrategyService) MarkBacktested(_ context.Context, req appStrategy.MarkBacktestedRequest) (*domstrategy.Strategy, error) {
	if err := f.failNext; err != nil {
		f.failNext = nil
		return nil, err
	}
	f.calls = append(f.calls, req.StrategyID)
	return &domstrategy.Strategy{ID: req.StrategyID}, nil
}

func (f *fakeStrategyService) Create(context.Context, appStrategy.CreateRequest) (*appStrategy.CreateResult, error) {
	return nil, nil
}
func (f *fakeStrategyService) Update(context.Context, appStrategy.UpdateRequest) (*domstrategy.Strategy, error) {
	return nil, nil
}
func (f *fakeStrategyService) Get(context.Context, int64) (*domstrategy.Strategy, error) {
	return nil, nil
}
func (f *fakeStrategyService) List(context.Context, appStrategy.ListQuery) ([]*domstrategy.Strategy, error) {
	return nil, nil
}
func (f *fakeStrategyService) Delete(context.Context, int64, int64) error { return nil }
func (f *fakeStrategyService) CreateVersion(context.Context, appStrategy.CreateVersionRequest) (*appStrategy.CreateVersionResult, error) {
	return nil, nil
}
func (f *fakeStrategyService) GetVersion(context.Context, int64) (*domversion.StrategyVersion, error) {
	return nil, nil
}
func (f *fakeStrategyService) ListVersions(context.Context, int64, int) ([]*domversion.StrategyVersion, error) {
	return nil, nil
}
func (f *fakeStrategyService) Publish(context.Context, appStrategy.PublishRequest) (*domstrategy.Strategy, error) {
	return nil, nil
}
func (f *fakeStrategyService) Archive(context.Context, appStrategy.ArchiveRequest) (*domstrategy.Strategy, error) {
	return nil, nil
}
func (f *fakeStrategyService) Fork(context.Context, appStrategy.ForkRequest) (*appStrategy.ForkResult, error) {
	return nil, nil
}

// TestOnFinished_HappyPath confirms a BacktestFinished event flips the
// strategy through the application service.
func TestOnFinished_HappyPath(t *testing.T) {
	svc := &fakeStrategyService{}
	h := NewMarkBacktestedHandler(svc)
	if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 1, StrategyID: 7}); err != nil {
		t.Fatalf("OnFinished: %v", err)
	}
	if len(svc.calls) != 1 || svc.calls[0] != 7 {
		t.Fatalf("expected MarkBacktested(7), got %v", svc.calls)
	}
}

// TestOnFinished_DedupesByJob confirms repeated events for the same job
// only flip the strategy once. The dedupe is keyed on JobID because
// the same strategy can be backtested many times with different jobs;
// each job legitimately advances the timeline.
func TestOnFinished_DedupesByJob(t *testing.T) {
	svc := &fakeStrategyService{}
	h := NewMarkBacktestedHandler(svc)
	for i := 0; i < 3; i++ {
		if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 42, StrategyID: 7}); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	if len(svc.calls) != 1 {
		t.Fatalf("expected dedupe to keep one call, got %d", len(svc.calls))
	}
	// A fresh job for the same strategy must still call through.
	if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 43, StrategyID: 7}); err != nil {
		t.Fatalf("new job: %v", err)
	}
	if len(svc.calls) != 2 {
		t.Fatalf("expected new job to bypass dedupe, got %d calls", len(svc.calls))
	}
}

// TestOnFinished_SkipsManualJobs guards that ad-hoc backtests with no
// strategy id do not stall the consumer.
func TestOnFinished_SkipsManualJobs(t *testing.T) {
	svc := &fakeStrategyService{}
	h := NewMarkBacktestedHandler(svc)
	if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 1}); err != nil {
		t.Fatalf("manual job: %v", err)
	}
	if len(svc.calls) != 0 {
		t.Fatalf("expected no calls for manual job, got %d", len(svc.calls))
	}
}

// TestOnFinished_SwallowsExpectedErrors checks that ErrStrategyArchived
// and ErrStrategyNotFound are treated as benign so a stale event does
// not stall the consumer.
func TestOnFinished_SwallowsExpectedErrors(t *testing.T) {
	for _, errExpected := range []error{stratErr.ErrStrategyArchived, stratErr.ErrStrategyNotFound} {
		svc := &fakeStrategyService{failNext: errExpected}
		h := NewMarkBacktestedHandler(svc)
		if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 1, StrategyID: 7}); err != nil {
			t.Fatalf("expected expected-error swallowed, got %v", err)
		}
	}
}

// TestOnFinished_PropagatesUnknownErrors confirms generic failures
// surface so the Kafka adapter can log them.
func TestOnFinished_PropagatesUnknownErrors(t *testing.T) {
	svc := &fakeStrategyService{failNext: errors.New("boom")}
	h := NewMarkBacktestedHandler(svc)
	if err := h.OnFinished(context.Background(), domsync.Envelope{}, domsync.FinishedPayload{JobID: 1, StrategyID: 7}); err == nil {
		t.Fatal("expected error to bubble")
	}
}

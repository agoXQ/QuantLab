package strategysync

import (
	"context"
	"testing"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	domportfolio "github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	domreport "github.com/agoXQ/QuantLab/app/backtest/domain/report"
	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
	domtrade "github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// fakeBacktestService is the minimal stub needed by the BaselineHandler
// tests; we only assert on Create + Submit.
type fakeBacktestService struct {
	created []appBacktest.CreateBacktestRequest
	submits []int64
	nextID  int64
	failSub bool
}

func (f *fakeBacktestService) Create(_ context.Context, req appBacktest.CreateBacktestRequest) (*appBacktest.CreateBacktestResult, error) {
	f.created = append(f.created, req)
	f.nextID++
	job := &backtestjob.BacktestJob{
		ID:             f.nextID,
		UserID:         req.UserID,
		StrategyID:     req.StrategyID,
		VersionID:      req.VersionID,
		Formula:        req.Formula,
		Universe:       req.Universe,
		InitialCapital: req.InitialCapital,
		Range:          req.Range,
		Status:         valueobject.JobStatusCreated,
	}
	return &appBacktest.CreateBacktestResult{Job: job}, nil
}

func (f *fakeBacktestService) Submit(_ context.Context, jobID int64) (*backtestjob.BacktestJob, error) {
	if f.failSub {
		return nil, context.DeadlineExceeded
	}
	f.submits = append(f.submits, jobID)
	return &backtestjob.BacktestJob{ID: jobID, Status: valueobject.JobStatusQueued}, nil
}

// Methods below are unused by the handler; we keep them stubbed so the
// fake satisfies the application Service interface.
func (f *fakeBacktestService) Run(context.Context, int64) (*appBacktest.RunResult, error) {
	return nil, nil
}
func (f *fakeBacktestService) Cancel(context.Context, int64, string) (*backtestjob.BacktestJob, error) {
	return nil, nil
}
func (f *fakeBacktestService) RunQueued(context.Context, int64) error { return nil }
func (f *fakeBacktestService) Get(context.Context, int64) (*backtestjob.BacktestJob, error) {
	return nil, nil
}
func (f *fakeBacktestService) List(context.Context, appBacktest.ListJobsQuery) ([]*backtestjob.BacktestJob, error) {
	return nil, nil
}
func (f *fakeBacktestService) GetReport(context.Context, int64) (*domreport.PerformanceReport, error) {
	return nil, nil
}
func (f *fakeBacktestService) GetTrades(context.Context, int64) ([]*domtrade.Trade, error) {
	return nil, nil
}
func (f *fakeBacktestService) GetSnapshots(context.Context, int64) ([]domportfolio.Snapshot, error) {
	return nil, nil
}
func (f *fakeBacktestService) Reconcile(context.Context) (appBacktest.ReconcileResult, error) {
	return appBacktest.ReconcileResult{}, nil
}

// stubResolver returns a snapshot keyed by (strategy, version) pair.
type stubResolver struct {
	snap *domsync.StrategySnapshot
	err  error
}

func (s *stubResolver) Resolve(_ context.Context, strategyID, versionID int64) (*domsync.StrategySnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.snap != nil {
		// Always return the canned snapshot but with the requested ids
		// so the handler's name builder uses real numbers.
		out := *s.snap
		out.StrategyID = strategyID
		out.VersionID = versionID
		return &out, nil
	}
	return &domsync.StrategySnapshot{StrategyID: strategyID, VersionID: versionID, FormulaText: "ROE > 0", Title: "stub"}, nil
}

// fixedClock returns a stable clock for deterministic ranges.
func fixedClock() time.Time { return time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC) }

func TestBaselineHandler_OnPublished_CreatesAndSubmits(t *testing.T) {
	bt := &fakeBacktestService{}
	handler := NewBaselineHandler(Config{AutoSubmit: true, Universe: []string{"600519"}}, &stubResolver{}, bt, nil, fixedClock)
	env := domsync.Envelope{EventType: domsync.EventStrategyPublished}
	if err := handler.OnPublished(context.Background(), env, domsync.PublishedPayload{StrategyID: 7, VersionID: 3, AuthorID: 42}); err != nil {
		t.Fatalf("OnPublished: %v", err)
	}
	if len(bt.created) != 1 {
		t.Fatalf("expected 1 create, got %d", len(bt.created))
	}
	if bt.created[0].StrategyID != 7 || bt.created[0].VersionID != 3 || bt.created[0].UserID != 42 {
		t.Fatalf("create payload mismatch: %+v", bt.created[0])
	}
	if bt.created[0].Formula != "ROE > 0" {
		t.Fatalf("expected formula ROE > 0, got %q", bt.created[0].Formula)
	}
	if got := bt.created[0].Universe; len(got) != 1 || got[0] != "600519" {
		t.Fatalf("universe override missing: %v", got)
	}
	if len(bt.submits) != 1 || bt.submits[0] != 1 {
		t.Fatalf("expected submit job 1, got %v", bt.submits)
	}
}

func TestBaselineHandler_DedupesSamePublish(t *testing.T) {
	bt := &fakeBacktestService{}
	handler := NewBaselineHandler(Config{AutoSubmit: true}, &stubResolver{}, bt, nil, fixedClock)
	env := domsync.Envelope{EventType: domsync.EventStrategyPublished}
	for i := 0; i < 3; i++ {
		if err := handler.OnPublished(context.Background(), env, domsync.PublishedPayload{StrategyID: 1, VersionID: 1, AuthorID: 1}); err != nil {
			t.Fatalf("OnPublished iter %d: %v", i, err)
		}
	}
	if len(bt.created) != 1 {
		t.Fatalf("expected dedupe to keep one create, got %d", len(bt.created))
	}
}

func TestBaselineHandler_OnVersionCreatedIsNoop(t *testing.T) {
	bt := &fakeBacktestService{}
	handler := NewBaselineHandler(Config{AutoSubmit: true}, &stubResolver{}, bt, nil, fixedClock)
	env := domsync.Envelope{EventType: domsync.EventStrategyVersionCreated}
	if err := handler.OnVersionCreated(context.Background(), env, domsync.VersionCreatedPayload{StrategyID: 1, VersionID: 1}); err != nil {
		t.Fatalf("OnVersionCreated: %v", err)
	}
	if len(bt.created) != 0 {
		t.Fatalf("expected no creates on version-created in MVP, got %d", len(bt.created))
	}
}

func TestBaselineHandler_RejectsEmptyFormula(t *testing.T) {
	bt := &fakeBacktestService{}
	handler := NewBaselineHandler(Config{AutoSubmit: true}, &stubResolver{snap: &domsync.StrategySnapshot{}}, bt, nil, fixedClock)
	env := domsync.Envelope{EventType: domsync.EventStrategyPublished}
	if err := handler.OnPublished(context.Background(), env, domsync.PublishedPayload{StrategyID: 1, VersionID: 1}); err == nil {
		t.Fatal("expected error on empty formula")
	}
}

// noopProvider lets us exercise the calendar branch without spinning up
// the full Market Data stack. TradingDays returns the supplied window
// verbatim so the handler picks it up unchanged.
type noopProvider struct{ days []time.Time }

func (n *noopProvider) LoadBars(context.Context, dommarket.BarsRequest) (map[string][]dommarket.Bar, error) {
	return nil, nil
}
func (n *noopProvider) TradingDays(_ context.Context, _ dommarket.CalendarRequest) ([]time.Time, error) {
	return n.days, nil
}

func TestBaselineHandler_PicksRangeFromCalendar(t *testing.T) {
	bt := &fakeBacktestService{}
	days := []time.Time{
		time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC),
	}
	handler := NewBaselineHandler(Config{AutoSubmit: false}, &stubResolver{}, bt, &noopProvider{days: days}, fixedClock)
	env := domsync.Envelope{EventType: domsync.EventStrategyPublished}
	if err := handler.OnPublished(context.Background(), env, domsync.PublishedPayload{StrategyID: 1, VersionID: 1}); err != nil {
		t.Fatalf("OnPublished: %v", err)
	}
	if len(bt.created) != 1 {
		t.Fatalf("expected 1 create, got %d", len(bt.created))
	}
	rng := bt.created[0].Range
	if !rng.Start.Equal(days[0]) || !rng.End.Equal(days[1]) {
		t.Fatalf("expected calendar range, got %+v", rng)
	}
}

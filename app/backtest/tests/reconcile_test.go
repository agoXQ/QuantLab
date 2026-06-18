package tests

import (
	"context"
	"testing"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// TestReconcile_RequeuesQueuedJobs simulates an ungraceful shutdown: a
// job is left in QUEUED in the repository while the queue itself is
// empty. After Reconcile runs the queue should carry the job and a
// fresh worker should drive it to COMPLETED.
func TestReconcile_RequeuesQueuedJobs(t *testing.T) {
	fx := newFixture(t)
	fx.enableAsync(t, 1)

	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 6)
	bars := linearBars(calendar, 30, 0.2)
	fx.provider.SetBars("000001", bars)
	fx.provider.SetCalendar(calendar)
	fx.formula.SetBars("000001", formulaBars(bars))
	fx.formula.SetFinancials("000001", map[string]float64{"ROE": 30})

	created, err := fx.svc.Create(ctxBg(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100000,
		Range:          valueobject.DateRange{Start: calendar[0], End: calendar[len(calendar)-1]},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Manually leave the row in QUEUED without enqueuing it; that is
	// the post-crash state Reconcile must repair.
	job, _ := fx.svc.Get(ctxBg(), created.Job.ID)
	if err := job.MarkQueued(time.Now()); err != nil {
		t.Fatalf("MarkQueued: %v", err)
	}
	if err := fx.deps.Jobs.Update(ctxBg(), job); err != nil {
		t.Fatalf("update queued: %v", err)
	}

	res, err := fx.svc.Reconcile(ctxBg())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if res.Requeued != 1 {
		t.Fatalf("expected 1 requeue, got %+v", res)
	}

	waitFor(t, 3*time.Second, 20*time.Millisecond, func() bool {
		got, err := fx.svc.Get(ctxBg(), created.Job.ID)
		if err != nil {
			return false
		}
		return got.Status == valueobject.JobStatusCompleted
	}, "requeued job to reach COMPLETED")
}

// TestReconcile_FailsStuckRunning verifies that a job left in RUNNING
// (because the previous worker died mid-replay) is flipped to FAILED
// with a deterministic reason rather than left forever.
func TestReconcile_FailsStuckRunning(t *testing.T) {
	fx := newFixture(t)

	// We do not enableAsync here: Reconcile must work without a queue
	// for stuck-RUNNING repair so a sync-only deployment still recovers.
	created, err := fx.svc.Create(ctxBg(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100000,
		Range: valueobject.DateRange{
			Start: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	job, _ := fx.svc.Get(ctxBg(), created.Job.ID)
	job.Status = valueobject.JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	if err := fx.deps.Jobs.Update(ctxBg(), job); err != nil {
		t.Fatalf("update running: %v", err)
	}

	res, err := fx.svc.Reconcile(ctxBg())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if res.FailedStuck != 1 {
		t.Fatalf("expected 1 failed-stuck, got %+v", res)
	}

	got, _ := fx.svc.Get(ctxBg(), created.Job.ID)
	if got.Status != valueobject.JobStatusFailed {
		t.Fatalf("expected FAILED, got %s", got.Status)
	}
	if got.ErrorMessage == "" {
		t.Fatal("expected error message recorded")
	}
}

// TestProgress_HTTPStatusReportsCompletion runs a job synchronously and
// then fetches /:id/status to confirm progress=1 once the job is done.
// We exercise the HTTP shape end to end so client expectations are
// covered alongside the application service.
func TestProgress_HTTPStatusReportsCompletion(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)
	start, end := seedSimpleUniverse(fx)

	body := map[string]any{
		"formula":         "ROE > 0",
		"universe":        []string{"000001"},
		"initial_capital": 100000,
		"start_date":      start.Format("2006-01-02"),
		"end_date":        end.Format("2006-01-02"),
		"config": map[string]any{
			"rebalance_frequency": "daily",
			"max_position_count":  1,
		},
	}
	w := executeJSON(router, "POST", "/api/v1/backtests?run=true", body)
	if w.Code != 201 {
		t.Fatalf("inline run: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	w = executeJSON(router, "GET", "/api/v1/backtests/1/status", nil)
	if w.Code != 200 {
		t.Fatalf("status: expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !indexOfContains(w.Body.String(), `"progress":1`) {
		t.Fatalf("expected progress=1 in body, got %s", w.Body.String())
	}
}

// indexOfContains is a tiny substring helper kept local to the
// progress / reconcile tests so we do not pollute http_test.go's API.
func indexOfContains(haystack, needle string) bool {
	return indexOf(haystack, needle) >= 0
}

// keep unused imports happy when the test file is compiled in isolation.
var _ = backtestjob.BacktestJob{}
var _ = context.Background

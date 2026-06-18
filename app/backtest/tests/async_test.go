package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// waitFor polls cond up to timeout, sleeping interval between checks.
func waitFor(t *testing.T, timeout, interval time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatalf("timed out waiting for: %s", msg)
}

// TestAsync_SubmitDrivesWorkerToCompletion is the happy path: Submit
// flips the job to QUEUED, the worker picks it up, runs it, and the
// repository ends with COMPLETED + a persisted report.
func TestAsync_SubmitDrivesWorkerToCompletion(t *testing.T) {
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

	job, err := fx.svc.Submit(ctxBg(), created.Job.ID)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if job.Status != valueobject.JobStatusQueued {
		t.Fatalf("expected QUEUED right after submit, got %s", job.Status)
	}

	waitFor(t, 3*time.Second, 20*time.Millisecond, func() bool {
		got, err := fx.svc.Get(ctxBg(), created.Job.ID)
		if err != nil {
			return false
		}
		return got.Status == valueobject.JobStatusCompleted
	}, "job to reach COMPLETED")

	rep, err := fx.svc.GetReport(ctxBg(), created.Job.ID)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if rep == nil {
		t.Fatal("expected non-nil report after async run")
	}
}

// TestAsync_SubmitWithoutQueueErrors verifies the synchronous-only mode
// still works: in a fixture without enableAsync, Submit must return
// ErrQueueUnavailable so callers fall back to Run.
func TestAsync_SubmitWithoutQueueErrors(t *testing.T) {
	fx := newFixture(t)
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
	if _, err := fx.svc.Submit(ctxBg(), created.Job.ID); !errors.Is(err, bterr.ErrQueueUnavailable) {
		t.Fatalf("expected ErrQueueUnavailable, got %v", err)
	}
}

// TestAsync_CancelStopsRunningJob enqueues a job, waits for the worker
// to start, cancels it, and expects the row to land in CANCELLED.
func TestAsync_CancelStopsRunningJob(t *testing.T) {
	fx := newFixture(t)
	fx.enableAsync(t, 1)

	// Long enough universe / calendar that runJob's mid-run status
	// re-read (every 16 days) gets a chance to observe the cancel.
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 100)
	bars := linearBars(calendar, 10, 0.05)
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

	if _, err := fx.svc.Submit(ctxBg(), created.Job.ID); err != nil {
		t.Fatalf("submit: %v", err)
	}

	if _, err := fx.svc.Cancel(ctxBg(), created.Job.ID, "operator"); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	waitFor(t, 3*time.Second, 20*time.Millisecond, func() bool {
		got, err := fx.svc.Get(ctxBg(), created.Job.ID)
		if err != nil {
			return false
		}
		return got.Status == valueobject.JobStatusCancelled
	}, "job to reach CANCELLED")
}

// TestAsync_HTTPRunReturns202 walks the HTTP surface end to end: POST
// :id/run without ?wait=true returns 202 + queued, then poll :id/status
// converges to COMPLETED.
func TestAsync_HTTPRunReturns202(t *testing.T) {
	fx := newFixture(t)
	fx.enableAsync(t, 1)
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
	w := executeJSON(router, "POST", "/api/v1/backtests", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	w = executeJSON(router, "POST", "/api/v1/backtests/1/run", nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("run: expected 202, got %d body=%s", w.Code, w.Body.String())
	}

	waitFor(t, 3*time.Second, 20*time.Millisecond, func() bool {
		w := executeJSON(router, "GET", "/api/v1/backtests/1/status", nil)
		return w.Code == http.StatusOK && containsStatus(w.Body.String(), "COMPLETED")
	}, "/status to report COMPLETED")
}

// containsStatus is a tiny string check because we only need to see the
// status token inside the JSON envelope without spinning up another
// decoder for the test.
func containsStatus(body, want string) bool {
	return len(body) > 0 && (indexOf(body, `"status":"`+want+`"`) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// keep stdlib happy in case the test runs in isolation.
var _ = context.Background

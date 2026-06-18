package backtestjob

import (
	"testing"
	"time"

	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// TestStateMachine guards the Created -> Queued -> Running -> Completed /
// Failed / Cancelled transitions. The async pipeline relies on these
// transitions being enforced by the aggregate, not by ad-hoc checks
// inside the application service.
func TestStateMachine(t *testing.T) {
	now := time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC)

	t.Run("created -> queued", func(t *testing.T) {
		j := &BacktestJob{Status: valueobject.JobStatusCreated}
		if err := j.MarkQueued(now); err != nil {
			t.Fatalf("MarkQueued: %v", err)
		}
		if j.Status != valueobject.JobStatusQueued {
			t.Fatalf("expected QUEUED, got %s", j.Status)
		}
	})

	t.Run("queued -> running -> completed", func(t *testing.T) {
		j := &BacktestJob{Status: valueobject.JobStatusCreated}
		_ = j.MarkQueued(now)
		if err := j.MarkRunning(now); err != nil {
			t.Fatalf("MarkRunning: %v", err)
		}
		if j.StartedAt == nil {
			t.Fatal("expected StartedAt set")
		}
		if err := j.MarkCompleted(now); err != nil {
			t.Fatalf("MarkCompleted: %v", err)
		}
		if j.FinishedAt == nil {
			t.Fatal("expected FinishedAt set")
		}
	})

	t.Run("re-queue clears previous attempt timestamps", func(t *testing.T) {
		ts := now
		j := &BacktestJob{
			Status:       valueobject.JobStatusQueued,
			ErrorMessage: "stale",
			StartedAt:    &ts,
			FinishedAt:   &ts,
		}
		if err := j.MarkQueued(now); err != nil {
			t.Fatalf("MarkQueued (idempotent): %v", err)
		}
		if j.ErrorMessage != "" || j.StartedAt != nil || j.FinishedAt != nil {
			t.Fatalf("expected stale fields cleared, got %+v", j)
		}
	})

	t.Run("cancel running", func(t *testing.T) {
		j := &BacktestJob{Status: valueobject.JobStatusRunning}
		if err := j.MarkCancelled(now, "operator"); err != nil {
			t.Fatalf("MarkCancelled: %v", err)
		}
		if j.Status != valueobject.JobStatusCancelled {
			t.Fatalf("expected CANCELLED, got %s", j.Status)
		}
		if j.ErrorMessage != "operator" {
			t.Fatalf("expected reason persisted, got %q", j.ErrorMessage)
		}
	})

	t.Run("cancel terminal job is rejected", func(t *testing.T) {
		j := &BacktestJob{Status: valueobject.JobStatusCompleted}
		if err := j.MarkCancelled(now, "late"); err != bterr.ErrJobNotCancellable {
			t.Fatalf("expected ErrJobNotCancellable, got %v", err)
		}
	})

	t.Run("re-queue terminal job is rejected", func(t *testing.T) {
		j := &BacktestJob{Status: valueobject.JobStatusFailed}
		if err := j.MarkQueued(now); err != bterr.ErrInvalidStateTransition {
			t.Fatalf("expected ErrInvalidStateTransition, got %v", err)
		}
	})
}

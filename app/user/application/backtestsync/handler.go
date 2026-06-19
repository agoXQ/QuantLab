// Package backtestsync is the application-layer use case that turns a
// decoded Backtest event into a User counter mutation. The MVP only
// counts BacktestFinished -> +1; failed / cancelled jobs intentionally
// stay outside the metric so the profile reflects real activity.
package backtestsync

import (
	"context"
	"fmt"
	"log"
	"sync"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	domsync "github.com/agoXQ/QuantLab/app/user/domain/backtestsync"
)

// CounterHandler implements domsync.Handler by delegating to the
// application service's IncrementBacktestCount hook.
type CounterHandler struct {
	svc appUser.Service

	dedupeMu sync.Mutex
	dedupe   map[int64]struct{}
}

// NewCounterHandler wires the use case. svc must not be nil.
func NewCounterHandler(svc appUser.Service) *CounterHandler {
	return &CounterHandler{svc: svc, dedupe: map[int64]struct{}{}}
}

// OnFinished bumps the user's backtest_count by 1.
func (h *CounterHandler) OnFinished(ctx context.Context, env domsync.Envelope, p domsync.FinishedPayload) error {
	if p.UserID == 0 {
		return nil
	}
	if h.markSeen(p.JobID) {
		return nil
	}
	if err := h.svc.IncrementBacktestCount(ctx, p.UserID, 1); err != nil {
		return fmt.Errorf("backtestsync: bump backtest_count user=%d: %w", p.UserID, err)
	}
	log.Printf("[backtestsync] user=%d backtest_count +1 job=%d", p.UserID, p.JobID)
	return nil
}

func (h *CounterHandler) markSeen(jobID int64) bool {
	if jobID == 0 {
		return false
	}
	h.dedupeMu.Lock()
	defer h.dedupeMu.Unlock()
	if _, ok := h.dedupe[jobID]; ok {
		return true
	}
	h.dedupe[jobID] = struct{}{}
	return false
}

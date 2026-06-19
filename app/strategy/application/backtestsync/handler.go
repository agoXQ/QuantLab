// Package backtestsync is the application-layer use case that turns a
// decoded Backtest event into a Strategy lifecycle transition. The
// MVP only handles BacktestFinished -> MarkBacktested, but the shape
// is identical to backtest's strategysync handler so future events
// (BacktestFailed, BacktestCancelled) slot in without restructuring.
package backtestsync

import (
	"context"
	"fmt"
	"log"
	"sync"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domsync "github.com/agoXQ/QuantLab/app/strategy/domain/backtestsync"
)

// MarkBacktestedHandler implements domsync.Handler by flipping the
// strategy aggregate into BACKTESTED whenever Backtest announces a
// finished job. Idempotent across replays; the underlying use case is
// itself idempotent (BACKTESTED is not a terminal state, and the
// service does not error when called repeatedly).
type MarkBacktestedHandler struct {
	svc appStrategy.Service

	dedupeMu sync.Mutex
	dedupe   map[int64]struct{}
}

// NewMarkBacktestedHandler wires the use case. svc must not be nil.
func NewMarkBacktestedHandler(svc appStrategy.Service) *MarkBacktestedHandler {
	return &MarkBacktestedHandler{
		svc:    svc,
		dedupe: map[int64]struct{}{},
	}
}

// OnFinished is called for every BacktestFinished event. We swallow
// "strategy archived" and "strategy not found" errors so an event for a
// stale aggregate does not stall the consumer; everything else is
// surfaced so the Kafka adapter can log it.
func (h *MarkBacktestedHandler) OnFinished(ctx context.Context, env domsync.Envelope, p domsync.FinishedPayload) error {
	if p.StrategyID == 0 {
		// Manual / ad-hoc backtests carry no strategy id; silently
		// skip them rather than treating it as a malformed payload.
		return nil
	}
	if h.markSeen(p.JobID) {
		return nil
	}
	if _, err := h.svc.MarkBacktested(ctx, appStrategy.MarkBacktestedRequest{StrategyID: p.StrategyID}); err != nil {
		// ErrStrategyArchived is an expected steady state when the
		// strategy was archived between submit and finish; logging it
		// keeps the operator informed without retrying forever.
		if isExpected(err) {
			log.Printf("[backtestsync] skip MarkBacktested strategy=%d: %v", p.StrategyID, err)
			return nil
		}
		return fmt.Errorf("backtestsync: mark backtested strategy=%d: %w", p.StrategyID, err)
	}
	log.Printf("[backtestsync] strategy=%d marked BACKTESTED on job=%d", p.StrategyID, p.JobID)
	return nil
}

// markSeen records the job id and reports whether we have already seen
// it in this process. Backtest only emits one Finished per job, so the
// map size is bounded by the run history; in production we push this
// into Redis behind the same boolean signature.
func (h *MarkBacktestedHandler) markSeen(jobID int64) bool {
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

// isExpected reports whether err is a benign domain refusal we should
// treat as success. Only ErrStrategyArchived / ErrStrategyNotFound
// land here today; the rest bubble up so the consumer can log + retry.
func isExpected(err error) bool {
	if err == nil {
		return true
	}
	switch err {
	case stratErr.ErrStrategyArchived, stratErr.ErrStrategyNotFound:
		return true
	default:
		return false
	}
}

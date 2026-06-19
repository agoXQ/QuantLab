// Package strategysync is the application-layer use case that turns a
// decoded Strategy event into a User counter mutation. The MVP only
// handles StrategyCreated -> +1 and StrategyArchived -> -1 on the
// author's strategy_count; future events (Forked) slot in without
// restructuring.
package strategysync

import (
	"context"
	"fmt"
	"log"
	"sync"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	domsync "github.com/agoXQ/QuantLab/app/user/domain/strategysync"
)

// CounterHandler implements domsync.Handler by delegating to the
// application service's IncrementStrategyCount hook. Idempotent
// across replays via a process-local seen set keyed by event_id.
type CounterHandler struct {
	svc appUser.Service

	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewCounterHandler wires the use case. svc must not be nil.
func NewCounterHandler(svc appUser.Service) *CounterHandler {
	return &CounterHandler{svc: svc, dedupe: map[string]struct{}{}}
}

// OnCreated bumps the author's strategy_count by 1.
func (h *CounterHandler) OnCreated(ctx context.Context, env domsync.Envelope, p domsync.CreatedPayload) error {
	if p.AuthorID == 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	if err := h.svc.IncrementStrategyCount(ctx, p.AuthorID, 1); err != nil {
		return fmt.Errorf("strategysync: bump strategy_count user=%d: %w", p.AuthorID, err)
	}
	log.Printf("[strategysync] user=%d strategy_count +1", p.AuthorID)
	return nil
}

// OnArchived decrements the author's strategy_count by 1; the
// repository clamps at zero so a stray double-event cannot underflow.
func (h *CounterHandler) OnArchived(ctx context.Context, env domsync.Envelope, p domsync.ArchivedPayload) error {
	if p.AuthorID == 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	if err := h.svc.IncrementStrategyCount(ctx, p.AuthorID, -1); err != nil {
		return fmt.Errorf("strategysync: drop strategy_count user=%d: %w", p.AuthorID, err)
	}
	log.Printf("[strategysync] user=%d strategy_count -1", p.AuthorID)
	return nil
}

// markSeen returns true when the event id was already processed.
// Empty event ids are not deduped, mirroring strategy/backtestsync.
func (h *CounterHandler) markSeen(eventID string) bool {
	if eventID == "" {
		return false
	}
	h.dedupeMu.Lock()
	defer h.dedupeMu.Unlock()
	if _, ok := h.dedupe[eventID]; ok {
		return true
	}
	h.dedupe[eventID] = struct{}{}
	return false
}

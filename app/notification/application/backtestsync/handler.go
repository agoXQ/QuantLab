// Package backtestsync is the application-layer use case that turns
// decoded Backtest events into Notification rows. Backtest jobs are
// per-user, so the fan-out is trivial: notify the job's owner.
package backtestsync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	domsync "github.com/agoXQ/QuantLab/app/notification/domain/backtestsync"
	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// Handler implements domsync.Handler.
type Handler struct {
	svc appNotif.Service

	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewHandler wires the use case.
func NewHandler(svc appNotif.Service) *Handler {
	return &Handler{svc: svc, dedupe: map[string]struct{}{}}
}

// OnFinished mints a BACKTEST notification for the job's owner with
// the headline metrics inline.
func (h *Handler) OnFinished(ctx context.Context, env domsync.Envelope, p domsync.FinishedPayload) error {
	if p.UserID <= 0 || p.JobID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	title := "回测完成"
	content := fmt.Sprintf(
		"job %d finished: total_return=%.2f%% sharpe=%.2f max_dd=%.2f%%",
		p.JobID, p.TotalReturn*100, p.Sharpe, p.MaxDrawdown*100,
	)
	if _, err := h.svc.DeliverNotification(ctx, appNotif.CreateNotificationInput{
		UserID:  p.UserID,
		Type:    valueobject.NotificationTypeBacktest,
		Title:   title,
		Content: content,
	}); err != nil {
		if errors.Is(err, notifErr.ErrInAppDisabled) {
			log.Printf("[backtestsync] finished job=%d user=%d in-app disabled", p.JobID, p.UserID)
			return nil
		}
		return fmt.Errorf("backtestsync: create finished notification user=%d: %w", p.UserID, err)
	}
	log.Printf("[backtestsync] finished job=%d user=%d", p.JobID, p.UserID)
	return nil
}

// OnFailed mints a BACKTEST notification with the failure reason. The
// owner can click through to see the run page.
func (h *Handler) OnFailed(ctx context.Context, env domsync.Envelope, p domsync.FailedPayload) error {
	if p.UserID <= 0 || p.JobID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	title := "回测失败"
	content := fmt.Sprintf("job %d failed: %s", p.JobID, p.Reason)
	if _, err := h.svc.DeliverNotification(ctx, appNotif.CreateNotificationInput{
		UserID:  p.UserID,
		Type:    valueobject.NotificationTypeBacktest,
		Title:   title,
		Content: content,
	}); err != nil {
		if errors.Is(err, notifErr.ErrInAppDisabled) {
			log.Printf("[backtestsync] failed job=%d user=%d in-app disabled", p.JobID, p.UserID)
			return nil
		}
		return fmt.Errorf("backtestsync: create failed notification user=%d: %w", p.UserID, err)
	}
	log.Printf("[backtestsync] failed job=%d user=%d", p.JobID, p.UserID)
	return nil
}

func (h *Handler) markSeen(eventID string) bool {
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

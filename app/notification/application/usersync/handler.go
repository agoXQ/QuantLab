// Package usersync is the application-layer use case that turns a
// decoded User event into a notification row. The MVP only handles
// UserFollowed -> create a FOLLOW notification on the followee; the
// rest are stubbed for parity so future expansion is purely additive.
package usersync

import (
	"context"
	"fmt"
	"log"
	"sync"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	domsync "github.com/agoXQ/QuantLab/app/notification/domain/usersync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// Handler implements domsync.Handler by delegating to the application
// service's CreateNotification hook. Idempotent across replays via a
// process-local seen set keyed by event_id.
type Handler struct {
	svc appNotif.Service

	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewHandler wires the use case. svc must not be nil.
func NewHandler(svc appNotif.Service) *Handler {
	return &Handler{svc: svc, dedupe: map[string]struct{}{}}
}

// OnRegistered is currently a no-op; the welcome flow lives elsewhere.
func (h *Handler) OnRegistered(_ context.Context, _ domsync.Envelope, _ domsync.RegisteredPayload) error {
	return nil
}

// OnFollowed creates a FOLLOW notification addressed to the followee.
func (h *Handler) OnFollowed(ctx context.Context, env domsync.Envelope, p domsync.FollowedPayload) error {
	if p.FolloweeID <= 0 || p.FollowerID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	if _, err := h.svc.CreateNotification(ctx, appNotif.CreateNotificationInput{
		UserID:  p.FolloweeID,
		Type:    valueobject.NotificationTypeFollow,
		Title:   "新关注",
		Content: fmt.Sprintf("user %d followed you", p.FollowerID),
	}); err != nil {
		return fmt.Errorf("usersync: create follow notification user=%d: %w", p.FolloweeID, err)
	}
	log.Printf("[usersync] notification follow user=%d follower=%d", p.FolloweeID, p.FollowerID)
	return nil
}

// OnUnfollowed is a no-op for the MVP; notifications are journals and
// keep historical follows visible.
func (h *Handler) OnUnfollowed(_ context.Context, _ domsync.Envelope, _ domsync.UnfollowedPayload) error {
	return nil
}

// markSeen returns true when the event id was already processed.
// Empty event ids are not deduped, mirroring the user-side handlers.
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

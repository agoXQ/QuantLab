// Package usersync is the application-layer use case that turns a
// decoded User event into a notification row plus an author
// subscription. The MVP wires two reactions:
//
//   - UserFollowed -> create a FOLLOW notification on the followee +
//     auto-subscribe the follower to ("author", followee_id) so the
//     strategy fan-out reaches anyone who follows that author.
//   - UserUnfollowed -> drop the matching author subscription so a
//     reverted follow stops producing strategy notifications.
//
// Notifications themselves are journals; nothing rewrites historical
// rows on unfollow.
package usersync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	domsync "github.com/agoXQ/QuantLab/app/notification/domain/usersync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// ObjectTypeAuthor is the subscription discriminator used by the
// implicit "follow == subscribe to author" wiring; strategy fan-out
// reads this in addition to the explicit ("strategy", id) rows.
const ObjectTypeAuthor = "author"

// Handler implements domsync.Handler by delegating to the application
// service plus the subscription repository. Idempotent across replays
// via a process-local seen set keyed by event_id.
type Handler struct {
	svc   appNotif.Service
	subs  domSub.Repository
	clock func() time.Time

	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewHandler wires the use case. svc + subs must not be nil; clock
// defaults to time.Now when zero.
func NewHandler(svc appNotif.Service, subs domSub.Repository, clock func() time.Time) *Handler {
	if clock == nil {
		clock = time.Now
	}
	return &Handler{svc: svc, subs: subs, clock: clock, dedupe: map[string]struct{}{}}
}

// OnRegistered is currently a no-op; the welcome flow lives elsewhere.
func (h *Handler) OnRegistered(_ context.Context, _ domsync.Envelope, _ domsync.RegisteredPayload) error {
	return nil
}

// OnFollowed creates a FOLLOW notification on the followee and
// auto-subscribes the follower to ("author", followee_id).
func (h *Handler) OnFollowed(ctx context.Context, env domsync.Envelope, p domsync.FollowedPayload) error {
	if p.FolloweeID <= 0 || p.FollowerID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	if _, err := h.svc.DeliverNotification(ctx, appNotif.CreateNotificationInput{
		UserID:  p.FolloweeID,
		Type:    valueobject.NotificationTypeFollow,
		Title:   "新关注",
		Content: fmt.Sprintf("user %d followed you", p.FollowerID),
	}); err != nil {
		if errors.Is(err, notifErr.ErrInAppDisabled) {
			log.Printf("[usersync] follow user=%d in-app disabled; skipped", p.FolloweeID)
		} else {
			return fmt.Errorf("usersync: create follow notification user=%d: %w", p.FolloweeID, err)
		}
	}
	if err := h.ensureAuthorSubscription(ctx, p.FollowerID, p.FolloweeID); err != nil {
		log.Printf("[usersync] author subscription follower=%d author=%d: %v", p.FollowerID, p.FolloweeID, err)
	}
	log.Printf("[usersync] notification follow user=%d follower=%d", p.FolloweeID, p.FollowerID)
	return nil
}

// OnUnfollowed drops the matching ("author", followee_id) subscription
// so the follower stops receiving strategy fan-out from the unfollowed
// author. Historical notifications stay in place by design.
func (h *Handler) OnUnfollowed(ctx context.Context, env domsync.Envelope, p domsync.UnfollowedPayload) error {
	if p.FolloweeID <= 0 || p.FollowerID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	if err := h.removeAuthorSubscription(ctx, p.FollowerID, p.FolloweeID); err != nil {
		log.Printf("[usersync] drop author subscription follower=%d author=%d: %v", p.FollowerID, p.FolloweeID, err)
	}
	return nil
}

func (h *Handler) ensureAuthorSubscription(ctx context.Context, follower, author int64) error {
	exists, err := h.subs.ExistsByObject(ctx, follower, ObjectTypeAuthor, author)
	if err != nil {
		return fmt.Errorf("probe author subscription: %w", err)
	}
	if exists {
		return nil
	}
	row := &domSub.Subscription{
		SubscriberID: follower,
		ObjectType:   domSub.NormaliseObjectType(ObjectTypeAuthor),
		ObjectID:     author,
		CreatedAt:    h.clock(),
	}
	if err := h.subs.Create(ctx, row); err != nil {
		// A racing duplicate is fine; treat as already present.
		if errors.Is(err, notifErr.ErrSubscriptionConflict) {
			return nil
		}
		return fmt.Errorf("create author subscription: %w", err)
	}
	return nil
}

func (h *Handler) removeAuthorSubscription(ctx context.Context, follower, author int64) error {
	rows, err := h.subs.List(ctx, domSub.ListFilter{
		SubscriberID: follower,
		ObjectType:   domSub.NormaliseObjectType(ObjectTypeAuthor),
		Limit:        100,
	})
	if err != nil {
		return fmt.Errorf("list author subscriptions: %w", err)
	}
	for _, row := range rows {
		if row.ObjectID != author {
			continue
		}
		if err := h.subs.Delete(ctx, follower, row.ID); err != nil {
			return fmt.Errorf("delete author subscription: %w", err)
		}
		return nil
	}
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

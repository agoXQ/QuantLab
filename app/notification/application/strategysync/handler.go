// Package strategysync is the application-layer use case that turns
// decoded Strategy events into Notification rows. The fan-out shape:
//
//   - StrategyPublished: notify every subscriber on
//     ("strategy", strategy_id) + the author of the source strategy
//     when present.
//   - StrategyForked:   notify the source author so they see attention
//     coming in.
//
// StrategyCreated is a no-op for the MVP; nothing is "subscribed" to
// the strategy yet at create time.
package strategysync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domsync "github.com/agoXQ/QuantLab/app/notification/domain/strategysync"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// ObjectTypeStrategy is the canonical subscription discriminator the
// application uses for strategy fan-out. Stored lowercase so the
// repository's normalisation rule applies.
const ObjectTypeStrategy = "strategy"

// ObjectTypeAuthor is the implicit subscription written by the
// User-followed handler so anyone following the author receives
// strategy fan-out. Kept here too so the strategy handler does not
// need to import the user-sync application package.
const ObjectTypeAuthor = "author"

// Handler implements domsync.Handler by delegating to the Notification
// service's CreateNotification hook plus the subscription repo.
//
// dedupe is keyed on event_id so a Kafka replay does not produce
// duplicate notifications even when the consumer restarts.
type Handler struct {
	svc  appNotif.Service
	subs domSub.Repository

	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewHandler wires the use case. svc + subs must not be nil.
func NewHandler(svc appNotif.Service, subs domSub.Repository) *Handler {
	return &Handler{svc: svc, subs: subs, dedupe: map[string]struct{}{}}
}

// OnCreated is currently a no-op; nothing in the MVP wants to be
// notified at strategy create time.
func (h *Handler) OnCreated(_ context.Context, _ domsync.Envelope, _ domsync.CreatedPayload) error {
	return nil
}

// OnPublished fans the event out to every subscriber on the strategy.
// The author themselves is filtered out so they do not receive a
// notification for their own publish.
func (h *Handler) OnPublished(ctx context.Context, env domsync.Envelope, p domsync.PublishedPayload) error {
	if p.StrategyID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	recipients, err := h.collectRecipients(ctx, p.StrategyID, p.AuthorID)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return nil
	}
	title := "策略已发布"
	content := fmt.Sprintf("strategy %d published a new version", p.StrategyID)
	for _, uid := range recipients {
		if uid == p.AuthorID || uid <= 0 {
			continue
		}
		if _, err := h.svc.DeliverNotification(ctx, appNotif.CreateNotificationInput{
			UserID:  uid,
			Type:    valueobject.NotificationTypeStrategy,
			Title:   title,
			Content: content,
		}); err != nil {
			if errors.Is(err, notifErr.ErrInAppDisabled) {
				continue
			}
			log.Printf("[strategysync] published fan-out user=%d: %v", uid, err)
		}
	}
	log.Printf("[strategysync] published strategy=%d fan-out=%d", p.StrategyID, len(recipients))
	return nil
}

// collectRecipients merges explicit ("strategy", id) subscribers with
// the implicit ("author", author_id) followers and returns the
// deduplicated set. The author is filtered downstream so a self-
// follow does not echo back at publish time.
func (h *Handler) collectRecipients(ctx context.Context, strategyID, authorID int64) ([]int64, error) {
	stratSubs, err := h.subs.ListSubscribers(ctx, ObjectTypeStrategy, strategyID)
	if err != nil {
		return nil, fmt.Errorf("strategysync: list subscribers strategy=%d: %w", strategyID, err)
	}
	var authorSubs []int64
	if authorID > 0 {
		authorSubs, err = h.subs.ListSubscribers(ctx, ObjectTypeAuthor, authorID)
		if err != nil {
			return nil, fmt.Errorf("strategysync: list subscribers author=%d: %w", authorID, err)
		}
	}
	if len(stratSubs) == 0 && len(authorSubs) == 0 {
		return nil, nil
	}
	seen := make(map[int64]struct{}, len(stratSubs)+len(authorSubs))
	out := make([]int64, 0, len(stratSubs)+len(authorSubs))
	for _, uid := range stratSubs {
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	for _, uid := range authorSubs {
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	return out, nil
}

// OnForked notifies the author of the source strategy that someone
// forked their work. The forking creator is intentionally not
// notified.
func (h *Handler) OnForked(ctx context.Context, env domsync.Envelope, p domsync.ForkedPayload) error {
	if p.SourceStrategyID <= 0 {
		return nil
	}
	if h.markSeen(env.EventID) {
		return nil
	}
	// Fan out to subscribers of the source strategy.
	subs, err := h.subs.ListSubscribers(ctx, ObjectTypeStrategy, p.SourceStrategyID)
	if err != nil {
		return fmt.Errorf("strategysync: list subscribers source=%d: %w", p.SourceStrategyID, err)
	}
	title := "策略被复刻"
	content := fmt.Sprintf("strategy %d was forked into %d", p.SourceStrategyID, p.TargetStrategyID)
	for _, uid := range subs {
		if uid == p.CreatorID || uid <= 0 {
			continue
		}
		if _, err := h.svc.DeliverNotification(ctx, appNotif.CreateNotificationInput{
			UserID:  uid,
			Type:    valueobject.NotificationTypeStrategy,
			Title:   title,
			Content: content,
		}); err != nil {
			if errors.Is(err, notifErr.ErrInAppDisabled) {
				continue
			}
			log.Printf("[strategysync] forked fan-out user=%d: %v", uid, err)
		}
	}
	log.Printf("[strategysync] forked source=%d target=%d fan-out=%d", p.SourceStrategyID, p.TargetStrategyID, len(subs))
	return nil
}

// markSeen returns true when the event id was already processed.
// Empty event ids are not deduped, mirroring user/strategysync.
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

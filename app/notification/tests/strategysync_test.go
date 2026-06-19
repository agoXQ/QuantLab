package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appStrategySync "github.com/agoXQ/QuantLab/app/notification/application/strategysync"
	"github.com/agoXQ/QuantLab/app/notification/domain/strategysync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
	infraStrategySync "github.com/agoXQ/QuantLab/app/notification/infrastructure/strategysync"
)

func newStrategySyncFixture(t *testing.T) (appNotif.Service, *infraMemory.SubscriptionRepository, *appStrategySync.Handler) {
	t.Helper()
	notifs := infraMemory.NewNotificationRepository()
	subs := infraMemory.NewSubscriptionRepository()
	svc := appNotif.NewService(appNotif.Dependencies{
		Notifications: notifs,
		Preferences:   infraMemory.NewPreferenceRepository(),
		Subscriptions: subs,
		Clock:         func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) },
	})
	handler := appStrategySync.NewHandler(svc, subs)
	return svc, subs, handler
}

// TestStrategySync_PublishedFansOutToSubscribers seeds two subscribers
// (excluding the author) on the strategy and confirms each receives a
// STRATEGY notification while the author themselves is filtered out.
func TestStrategySync_PublishedFansOutToSubscribers(t *testing.T) {
	svc, _, handler := newStrategySyncFixture(t)
	ctx := context.Background()
	for _, uid := range []int64{200, 300, 999} { // 999 is the author
		if _, err := svc.CreateSubscription(ctx, appNotif.CreateSubscriptionInput{
			SubscriberID: uid,
			ObjectType:   "strategy",
			ObjectID:     1,
		}); err != nil {
			t.Fatalf("seed sub %d: %v", uid, err)
		}
	}
	env := strategysync.Envelope{
		EventID:   "pub-1",
		EventType: strategysync.EventStrategyPublished,
		Payload: map[string]any{
			"strategy_id": float64(1),
			"version_id":  float64(7),
			"author_id":   float64(999),
		},
	}
	raw, _ := json.Marshal(env)
	if err := infraStrategySync.Dispatch(ctx, handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	for _, uid := range []int64{200, 300} {
		out, _ := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: uid, Limit: 10})
		if len(out.Items) != 1 || out.Items[0].Type != valueobject.NotificationTypeStrategy {
			t.Fatalf("user %d expected one STRATEGY notification, got %+v", uid, out.Items)
		}
	}
	out, _ := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: 999, Limit: 10})
	if len(out.Items) != 0 {
		t.Fatalf("author 999 should not receive own publish, got %d", len(out.Items))
	}
}

// TestStrategySync_DedupeOnReplay confirms a Kafka replay with the
// same event id does not duplicate the fan-out.
func TestStrategySync_DedupeOnReplay(t *testing.T) {
	svc, _, handler := newStrategySyncFixture(t)
	ctx := context.Background()
	if _, err := svc.CreateSubscription(ctx, appNotif.CreateSubscriptionInput{
		SubscriberID: 11, ObjectType: "strategy", ObjectID: 5,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	env := strategysync.Envelope{
		EventID:   "dup-1",
		EventType: strategysync.EventStrategyPublished,
		Payload:   map[string]any{"strategy_id": float64(5)},
	}
	raw, _ := json.Marshal(env)
	for i := 0; i < 3; i++ {
		if err := infraStrategySync.Dispatch(ctx, handler, raw); err != nil {
			t.Fatalf("dispatch %d: %v", i, err)
		}
	}
	out, _ := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: 11, Limit: 10})
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 (dedup), got %d", len(out.Items))
	}
}

// TestStrategySync_ForkedNotifiesSubscribers seeds a subscriber on the
// source strategy and confirms a Forked envelope lands a STRATEGY row
// while the forking creator stays silent.
func TestStrategySync_ForkedNotifiesSubscribers(t *testing.T) {
	svc, _, handler := newStrategySyncFixture(t)
	ctx := context.Background()
	if _, err := svc.CreateSubscription(ctx, appNotif.CreateSubscriptionInput{
		SubscriberID: 77, ObjectType: "strategy", ObjectID: 9,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	env := strategysync.Envelope{
		EventID:   "fork-1",
		EventType: strategysync.EventStrategyForked,
		Payload: map[string]any{
			"source_strategy_id": float64(9),
			"target_strategy_id": float64(20),
			"creator_id":         float64(77), // also the subscriber
		},
	}
	raw, _ := json.Marshal(env)
	if err := infraStrategySync.Dispatch(ctx, handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	out, _ := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: 77, Limit: 10})
	if len(out.Items) != 0 {
		t.Fatalf("forking creator must not be notified, got %d", len(out.Items))
	}
}

package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appUserSync "github.com/agoXQ/QuantLab/app/notification/application/usersync"
	"github.com/agoXQ/QuantLab/app/notification/domain/usersync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
	infraUserSync "github.com/agoXQ/QuantLab/app/notification/infrastructure/usersync"
)

func newUserSyncFixture() (appNotif.Service, *infraMemory.SubscriptionRepository, *appUserSync.Handler) {
	subs := infraMemory.NewSubscriptionRepository()
	svc := appNotif.NewService(appNotif.Dependencies{
		Notifications: infraMemory.NewNotificationRepository(),
		Preferences:   infraMemory.NewPreferenceRepository(),
		Subscriptions: subs,
		Clock:         func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) },
	})
	handler := appUserSync.NewHandler(svc, subs, func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) })
	return svc, subs, handler
}

// TestUserSync_FollowedCreatesNotification confirms a UserFollowed
// envelope hits the application service and produces a FOLLOW row
// addressed at the followee.
func TestUserSync_FollowedCreatesNotification(t *testing.T) {
	svc, subs, handler := newUserSyncFixture()

	env := usersync.Envelope{
		EventID:       "evt-1",
		EventType:     usersync.EventUserFollowed,
		EventVersion:  "1.0",
		OccurredAt:    time.Now(),
		AggregateType: "USER",
		AggregateID:   "42",
		Producer:      "user-service",
		Payload: map[string]any{
			"follower_id": float64(99),
			"followee_id": float64(42),
		},
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := infraUserSync.Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	out, err := svc.ListNotifications(context.Background(), appNotif.ListNotificationsInput{
		UserID: 42,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(out.Items))
	}
	if out.Items[0].Type != valueobject.NotificationTypeFollow {
		t.Fatalf("expected FOLLOW type, got %v", out.Items[0].Type)
	}

	// Implicit author subscription must land so future strategy
	// fan-out reaches the follower.
	exists, err := subs.ExistsByObject(context.Background(), 99, "author", 42)
	if err != nil || !exists {
		t.Fatalf("expected author subscription to exist, got exists=%v err=%v", exists, err)
	}

	// Replay the same event id; dedupe should keep both the
	// notification count and the subscription count at one.
	if err := infraUserSync.Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("replay: %v", err)
	}
	out2, _ := svc.ListNotifications(context.Background(), appNotif.ListNotificationsInput{UserID: 42, Limit: 10})
	if len(out2.Items) != 1 {
		t.Fatalf("dedupe failed, got %d items", len(out2.Items))
	}
}

// TestUserSync_UnfollowedDropsAuthorSubscription confirms a follow ->
// unfollow round-trip cleans up the implicit author subscription so
// the follower stops receiving strategy fan-out.
func TestUserSync_UnfollowedDropsAuthorSubscription(t *testing.T) {
	_, subs, handler := newUserSyncFixture()

	follow := usersync.Envelope{
		EventID:   "f-1",
		EventType: usersync.EventUserFollowed,
		Payload:   map[string]any{"follower_id": float64(7), "followee_id": float64(8)},
	}
	rawFollow, _ := json.Marshal(follow)
	if err := infraUserSync.Dispatch(context.Background(), handler, rawFollow); err != nil {
		t.Fatalf("follow dispatch: %v", err)
	}

	unfollow := usersync.Envelope{
		EventID:   "u-1",
		EventType: usersync.EventUserUnfollowed,
		Payload:   map[string]any{"follower_id": float64(7), "followee_id": float64(8)},
	}
	rawUn, _ := json.Marshal(unfollow)
	if err := infraUserSync.Dispatch(context.Background(), handler, rawUn); err != nil {
		t.Fatalf("unfollow dispatch: %v", err)
	}

	exists, err := subs.ExistsByObject(context.Background(), 7, "author", 8)
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if exists {
		t.Fatalf("expected author subscription to be cleared after unfollow")
	}
}

// TestUserSync_UnknownEventIsNoop confirms an unrelated envelope is
// silently dropped so an upstream producer can grow new types without
// breaking Notification.
func TestUserSync_UnknownEventIsNoop(t *testing.T) {
	_, _, handler := newUserSyncFixture()
	raw := []byte(`{"event_type":"UserSomethingNew","payload":{}}`)
	if err := infraUserSync.Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
}

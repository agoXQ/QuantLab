package tests

import (
	"context"
	"testing"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
)

func newServiceFixture() appNotif.Service {
	return appNotif.NewService(appNotif.Dependencies{
		Notifications: infraMemory.NewNotificationRepository(),
		Preferences:   infraMemory.NewPreferenceRepository(),
		Subscriptions: infraMemory.NewSubscriptionRepository(),
		Clock:         func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) },
	})
}

func TestService_CreateAndList(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{
			UserID:  101,
			Type:    valueobject.NotificationTypeFollow,
			Title:   "新关注",
			Content: "user 99 followed you",
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}
	out, err := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: 101, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(out.Items))
	}
	if out.NextCursor != "" {
		t.Fatalf("expected empty cursor, got %q", out.NextCursor)
	}
	count, err := svc.GetUnreadCount(ctx, 101)
	if err != nil || count != 3 {
		t.Fatalf("unread = %d err=%v", count, err)
	}
}

func TestService_MarkReadAndAllRead(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	first, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{
		UserID: 7, Type: valueobject.NotificationTypeSystem, Title: "hello",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{
		UserID: 7, Type: valueobject.NotificationTypeStrategy, Title: "publish",
	}); err != nil {
		t.Fatalf("create2: %v", err)
	}
	if err := svc.MarkRead(ctx, 7, first.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if err := svc.MarkRead(ctx, 7, first.ID); err != nil {
		t.Fatalf("mark read idempotent: %v", err)
	}
	count, _ := svc.GetUnreadCount(ctx, 7)
	if count != 1 {
		t.Fatalf("expected 1 unread, got %d", count)
	}
	affected, err := svc.MarkAllRead(ctx, 7)
	if err != nil {
		t.Fatalf("mark all: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected 1 affected, got %d", affected)
	}
}

func TestService_DeleteHidesFromList(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	n, _ := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{
		UserID: 5, Type: valueobject.NotificationTypeMention, Title: "@you",
	})
	if err := svc.DeleteNotification(ctx, 5, n.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	out, _ := svc.ListNotifications(ctx, appNotif.ListNotificationsInput{UserID: 5, Limit: 10})
	if len(out.Items) != 0 {
		t.Fatalf("expected 0 visible items, got %d", len(out.Items))
	}
}

func TestService_PreferencesDefaultsAndUpdate(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	pref, err := svc.GetPreferences(ctx, 11)
	if err != nil {
		t.Fatalf("get defaults: %v", err)
	}
	if !pref.InAppEnabled || pref.EmailEnabled {
		t.Fatalf("unexpected defaults: %+v", pref)
	}
	updated, err := svc.UpdatePreferences(ctx, appNotif.UpdatePreferencesInput{
		UserID: 11, InAppEnabled: true, EmailEnabled: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !updated.EmailEnabled {
		t.Fatalf("update did not stick: %+v", updated)
	}
	roundtrip, err := svc.GetPreferences(ctx, 11)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if !roundtrip.EmailEnabled {
		t.Fatalf("read after upsert did not surface change: %+v", roundtrip)
	}
}

func TestService_SubscriptionLifecycle(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	sub, err := svc.CreateSubscription(ctx, appNotif.CreateSubscriptionInput{
		SubscriberID: 21, ObjectType: "STRATEGY", ObjectID: 99,
	})
	if err != nil {
		t.Fatalf("create sub: %v", err)
	}
	if sub.ObjectType != "strategy" {
		t.Fatalf("object type not normalised: %q", sub.ObjectType)
	}
	if _, err := svc.CreateSubscription(ctx, appNotif.CreateSubscriptionInput{
		SubscriberID: 21, ObjectType: "strategy", ObjectID: 99,
	}); err != notifErr.ErrSubscriptionConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	out, err := svc.ListSubscriptions(ctx, appNotif.ListSubscriptionsInput{SubscriberID: 21, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out.Items))
	}
	if err := svc.CancelSubscription(ctx, 21, sub.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	out2, _ := svc.ListSubscriptions(ctx, appNotif.ListSubscriptionsInput{SubscriberID: 21, Limit: 10})
	if len(out2.Items) != 0 {
		t.Fatalf("expected empty, got %d", len(out2.Items))
	}
}

func TestService_ValidationErrors(t *testing.T) {
	svc := newServiceFixture()
	ctx := context.Background()
	if _, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{UserID: 0, Type: valueobject.NotificationTypeFollow, Title: "hi"}); err != notifErr.ErrInvalidUserID {
		t.Fatalf("expected invalid user id, got %v", err)
	}
	if _, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{UserID: 1, Type: valueobject.NotificationTypeUnspecified, Title: "hi"}); err != notifErr.ErrInvalidType {
		t.Fatalf("expected invalid type, got %v", err)
	}
	if _, err := svc.CreateNotification(ctx, appNotif.CreateNotificationInput{UserID: 1, Type: valueobject.NotificationTypeFollow, Title: ""}); err != notifErr.ErrInvalidNotification {
		t.Fatalf("expected invalid title, got %v", err)
	}
}

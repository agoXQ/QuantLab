package tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appBacktestSync "github.com/agoXQ/QuantLab/app/notification/application/backtestsync"
	"github.com/agoXQ/QuantLab/app/notification/domain/backtestsync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
	infraBacktestSync "github.com/agoXQ/QuantLab/app/notification/infrastructure/backtestsync"
)

func newBacktestSyncFixture() (appNotif.Service, *appBacktestSync.Handler) {
	svc := appNotif.NewService(appNotif.Dependencies{
		Notifications: infraMemory.NewNotificationRepository(),
		Preferences:   infraMemory.NewPreferenceRepository(),
		Subscriptions: infraMemory.NewSubscriptionRepository(),
		Clock:         func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) },
	})
	return svc, appBacktestSync.NewHandler(svc)
}

func TestBacktestSync_FinishedNotifiesOwner(t *testing.T) {
	svc, handler := newBacktestSyncFixture()
	env := backtestsync.Envelope{
		EventID:   "bt-1",
		EventType: backtestsync.EventBacktestFinished,
		Payload: map[string]any{
			"job_id":        float64(42),
			"user_id":       float64(7),
			"total_return":  0.235,
			"sharpe_ratio":  1.4,
			"max_drawdown":  0.082,
		},
	}
	raw, _ := json.Marshal(env)
	if err := infraBacktestSync.Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	out, _ := svc.ListNotifications(context.Background(), appNotif.ListNotificationsInput{UserID: 7, Limit: 10})
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(out.Items))
	}
	if out.Items[0].Type != valueobject.NotificationTypeBacktest {
		t.Fatalf("expected BACKTEST type, got %v", out.Items[0].Type)
	}
	if !strings.Contains(out.Items[0].Content, "total_return=23.50") {
		t.Fatalf("formatted content missing metric: %q", out.Items[0].Content)
	}
}

func TestBacktestSync_FailedNotifiesOwner(t *testing.T) {
	svc, handler := newBacktestSyncFixture()
	env := backtestsync.Envelope{
		EventID:   "bt-fail-1",
		EventType: backtestsync.EventBacktestFailed,
		Payload: map[string]any{
			"job_id":  float64(99),
			"user_id": float64(3),
			"reason":  "ohlc gap",
		},
	}
	raw, _ := json.Marshal(env)
	if err := infraBacktestSync.Dispatch(context.Background(), handler, raw); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	out, _ := svc.ListNotifications(context.Background(), appNotif.ListNotificationsInput{UserID: 3, Limit: 10})
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(out.Items))
	}
	if !strings.Contains(out.Items[0].Content, "ohlc gap") {
		t.Fatalf("expected reason in content: %q", out.Items[0].Content)
	}
}

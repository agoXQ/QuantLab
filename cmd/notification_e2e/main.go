// Command notification_e2e is the in-process smoke harness for the
// Notification fan-out chain. It boots an in-memory Notification
// service, the matching User / Strategy / Backtest application
// services, and then drives a realistic scenario:
//
//  1. user A registers, user B registers and follows A.
//  2. A creates a strategy and publishes a new version.
//  3. The fan-out reaches B (via the implicit "author" subscription).
//  4. A backtest finishes for B; the BACKTEST notification lands.
//  5. The harness reads the list / unread-count / mark-read APIs and
//     asserts the platform-level behaviour end to end.
//
// Failure prints a one-liner and exits non-zero so a CI step can pick
// up regressions; success exits 0 with a short summary.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appBacktestSync "github.com/agoXQ/QuantLab/app/notification/application/backtestsync"
	appStrategySync "github.com/agoXQ/QuantLab/app/notification/application/strategysync"
	appUserSync "github.com/agoXQ/QuantLab/app/notification/application/usersync"
	domBacktestSync "github.com/agoXQ/QuantLab/app/notification/domain/backtestsync"
	domStrategySync "github.com/agoXQ/QuantLab/app/notification/domain/strategysync"
	domUserSync "github.com/agoXQ/QuantLab/app/notification/domain/usersync"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	infraNotifMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "notification_e2e: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	notifs := infraNotifMemory.NewNotificationRepository()
	prefs := infraNotifMemory.NewPreferenceRepository()
	subs := infraNotifMemory.NewSubscriptionRepository()
	svc := appNotif.NewService(appNotif.Dependencies{
		Notifications: notifs,
		Preferences:   prefs,
		Subscriptions: subs,
		Clock:         clock,
	})

	userHandler := appUserSync.NewHandler(svc, subs, clock)
	stratHandler := appStrategySync.NewHandler(svc, subs)
	btHandler := appBacktestSync.NewHandler(svc)

	const (
		alice = int64(1001) // strategy author
		bob   = int64(2002) // follower
	)

	// 1. Alice publishes; nothing happens yet because Bob has not
	// followed her.
	step("alice publishes (no audience)")
	if err := stratHandler.OnPublished(ctx, domStrategySync.Envelope{
		EventID:   "pub-cold-1",
		EventType: domStrategySync.EventStrategyPublished,
	}, domStrategySync.PublishedPayload{
		StrategyID: 100, AuthorID: alice,
	}); err != nil {
		return fmt.Errorf("warmup publish: %w", err)
	}
	if got := unread(t(svc, bob)); got != 0 {
		return fmt.Errorf("expected 0 notifications for bob before follow, got %d", got)
	}

	// 2. Bob follows Alice; the FOLLOW notification lands on Alice
	// and the implicit ("author", alice) subscription seeds Bob into
	// the strategy fan-out path.
	step("bob follows alice")
	if err := userHandler.OnFollowed(ctx, domUserSync.Envelope{
		EventID:   "fol-1",
		EventType: domUserSync.EventUserFollowed,
	}, domUserSync.FollowedPayload{FollowerID: bob, FolloweeID: alice}); err != nil {
		return fmt.Errorf("follow dispatch: %w", err)
	}
	if got := unread(t(svc, alice)); got != 1 {
		return fmt.Errorf("alice expected 1 follow notification, got %d", got)
	}
	if got := unread(t(svc, bob)); got != 0 {
		return fmt.Errorf("bob expected 0 notifications after follow, got %d", got)
	}

	// 3. Alice publishes another version; Bob is reached via the
	// implicit author subscription.
	step("alice publishes (bob is following)")
	if err := stratHandler.OnPublished(ctx, domStrategySync.Envelope{
		EventID:   "pub-warm-1",
		EventType: domStrategySync.EventStrategyPublished,
	}, domStrategySync.PublishedPayload{
		StrategyID: 100, AuthorID: alice,
	}); err != nil {
		return fmt.Errorf("warm publish: %w", err)
	}
	if got := unread(t(svc, bob)); got != 1 {
		return fmt.Errorf("bob expected 1 strategy notification, got %d", got)
	}

	// 4. Bob runs a backtest; finished event addresses Bob directly.
	step("backtest finished for bob")
	if err := btHandler.OnFinished(ctx, domBacktestSync.Envelope{
		EventID:   "bt-1",
		EventType: domBacktestSync.EventBacktestFinished,
	}, domBacktestSync.FinishedPayload{
		JobID: 7, UserID: bob, StrategyID: 100,
		TotalReturn: 0.18, Sharpe: 1.4, MaxDrawdown: 0.12,
	}); err != nil {
		return fmt.Errorf("backtest finished: %w", err)
	}
	bobItems := list(svc, bob)
	if len(bobItems) != 2 {
		return fmt.Errorf("bob expected 2 notifications total, got %d", len(bobItems))
	}
	if !containsType(bobItems, valueobject.NotificationTypeBacktest) {
		return fmt.Errorf("bob is missing BACKTEST notification: %s", typesOf(bobItems))
	}

	// 5. Bob disables in-app preference; a second publish must skip.
	step("bob disables in-app, alice publishes again")
	if _, err := svc.UpdatePreferences(ctx, appNotif.UpdatePreferencesInput{
		UserID: bob, InAppEnabled: false,
	}); err != nil {
		return fmt.Errorf("toggle off: %w", err)
	}
	if err := stratHandler.OnPublished(ctx, domStrategySync.Envelope{
		EventID:   "pub-warm-2",
		EventType: domStrategySync.EventStrategyPublished,
	}, domStrategySync.PublishedPayload{
		StrategyID: 100, AuthorID: alice,
	}); err != nil {
		return fmt.Errorf("publish 2: %w", err)
	}
	if got := unread(t(svc, bob)); got != 2 {
		return fmt.Errorf("bob expected 2 unread (preference-off blocked the new one), got %d", got)
	}

	// 6. Mark-read flips the unread count to 0.
	step("bob marks all read")
	if _, err := svc.MarkAllRead(ctx, bob); err != nil {
		return fmt.Errorf("mark all: %w", err)
	}
	if got := unread(t(svc, bob)); got != 0 {
		return fmt.Errorf("bob expected 0 unread after mark-all, got %d", got)
	}

	// 7. Bob unfollows; the implicit subscription drops, so a third
	// publish does not produce a row even after re-enabling in-app.
	step("bob unfollows + re-enables, then alice publishes")
	if err := userHandler.OnUnfollowed(ctx, domUserSync.Envelope{
		EventID:   "unfol-1",
		EventType: domUserSync.EventUserUnfollowed,
	}, domUserSync.UnfollowedPayload{FollowerID: bob, FolloweeID: alice}); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	if _, err := svc.UpdatePreferences(ctx, appNotif.UpdatePreferencesInput{
		UserID: bob, InAppEnabled: true,
	}); err != nil {
		return fmt.Errorf("toggle on: %w", err)
	}
	if err := stratHandler.OnPublished(ctx, domStrategySync.Envelope{
		EventID:   "pub-cold-2",
		EventType: domStrategySync.EventStrategyPublished,
	}, domStrategySync.PublishedPayload{
		StrategyID: 100, AuthorID: alice,
	}); err != nil {
		return fmt.Errorf("publish 3: %w", err)
	}
	if got := unread(t(svc, bob)); got != 0 {
		return fmt.Errorf("bob expected 0 unread after unfollow, got %d", got)
	}

	fmt.Println()
	fmt.Println("notification_e2e: all checks passed")
	fmt.Printf("  alice notifications: %d\n", len(list(svc, alice)))
	fmt.Printf("  bob   notifications: %d (final unread=%d)\n", len(list(svc, bob)), unread(t(svc, bob)))
	return nil
}

func step(label string) {
	log.Printf("[step] %s", label)
}

func t(svc appNotif.Service, uid int64) int64 {
	n, err := svc.GetUnreadCount(context.Background(), uid)
	if err != nil {
		log.Fatalf("unread count user=%d: %v", uid, err)
	}
	return n
}

func unread(v int64) int64 { return v }

func list(svc appNotif.Service, uid int64) []*notificationView {
	out, err := svc.ListNotifications(context.Background(), appNotif.ListNotificationsInput{
		UserID: uid, Limit: 50,
	})
	if err != nil {
		log.Fatalf("list user=%d: %v", uid, err)
	}
	views := make([]*notificationView, 0, len(out.Items))
	for _, item := range out.Items {
		views = append(views, &notificationView{Type: item.Type, Title: item.Title})
	}
	return views
}

type notificationView struct {
	Type  valueobject.NotificationType
	Title string
}

func containsType(items []*notificationView, t valueobject.NotificationType) bool {
	for _, it := range items {
		if it.Type == t {
			return true
		}
	}
	return false
}

func typesOf(items []*notificationView) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, it.Type.String())
	}
	return strings.Join(parts, ",")
}

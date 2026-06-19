// Package notification is the application layer for the Notification
// Service. It exposes a single Service interface backed by repository
// + clock dependencies so the gRPC / HTTP / Kafka adapters share one
// orchestration surface.
package notification

import (
	"time"

	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// CreateNotificationInput is the orchestration shape used by the
// fan-out path; both the in-process consumer and the gRPC adapter
// translate their inputs into this struct.
type CreateNotificationInput struct {
	UserID  int64
	Type    valueobject.NotificationType
	Title   string
	Content string
}

// ListNotificationsInput is the query shape used by the list endpoint.
// The Cursor is opaque from the caller's perspective; the repository
// MVP decodes it as an offset to keep the implementation simple.
type ListNotificationsInput struct {
	UserID    int64
	Cursor    string
	Limit     int
	Statuses  []valueobject.NotificationStatus
	Types     []valueobject.NotificationType
}

// ListNotificationsOutput pairs the page with the next cursor.
type ListNotificationsOutput struct {
	Items      []*domNotif.Notification
	NextCursor string
}

// UpdatePreferencesInput captures the set of channel toggles the
// caller wants to apply.
type UpdatePreferencesInput struct {
	UserID         int64
	InAppEnabled   bool
	EmailEnabled   bool
	WebhookEnabled bool
	PushEnabled    bool
}

// CreateSubscriptionInput is the orchestration shape used by
// CreateSubscription.
type CreateSubscriptionInput struct {
	SubscriberID int64
	ObjectType   string
	ObjectID     int64
}

// ListSubscriptionsInput is the query shape used by ListSubscriptions.
type ListSubscriptionsInput struct {
	SubscriberID int64
	ObjectType   string
	Cursor       string
	Limit        int
}

// ListSubscriptionsOutput pairs the page with the next cursor.
type ListSubscriptionsOutput struct {
	Items      []*domSub.Subscription
	NextCursor string
}

// Dependencies bundles the dependencies used by the service.
type Dependencies struct {
	Notifications domNotif.Repository
	Preferences   domPref.Repository
	Subscriptions domSub.Repository
	Clock         func() time.Time
}

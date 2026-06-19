// Package usersync defines the contract the Notification service uses
// to react to User lifecycle events. The shape mirrors how User
// consumes Strategy / Backtest events: a small Envelope + payload
// pair per event type, plus a Handler interface the consumer
// dispatches to.
//
// Re-declaring the JSON-level event names instead of importing the
// User service's domain types keeps the Notification service free of
// build-time coupling. The platform Event Specification is the
// integration contract; a wire-compat test pins the two ends together.
package usersync

import (
	"context"
	"time"
)

// EventType lists the canonical User events Notification cares about.
type EventType string

const (
	EventUserRegistered EventType = "UserRegistered"
	EventUserFollowed   EventType = "UserFollowed"
	EventUserUnfollowed EventType = "UserUnfollowed"
)

// Envelope mirrors the common Event shape every QuantLab service emits.
type Envelope struct {
	EventID       string         `json:"event_id"`
	EventType     EventType      `json:"event_type"`
	EventVersion  string         `json:"event_version"`
	OccurredAt    time.Time      `json:"occurred_at"`
	AggregateType string         `json:"aggregate_type"`
	AggregateID   string         `json:"aggregate_id"`
	Producer      string         `json:"producer"`
	Payload       map[string]any `json:"payload"`
}

// RegisteredPayload announces a new account. We use it to seed a
// default preferences row so the first preferences GET returns a
// real document.
type RegisteredPayload struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// FollowedPayload notifies that follower_id followed followee_id.
// Notification turns it into a FOLLOW row addressed at the followee.
type FollowedPayload struct {
	FollowerID int64 `json:"follower_id"`
	FolloweeID int64 `json:"followee_id"`
}

// UnfollowedPayload is the inverse of FollowedPayload. The MVP records
// the event for parity but does not delete the original notification:
// notifications are journals, not subscriptions.
type UnfollowedPayload struct {
	FollowerID int64 `json:"follower_id"`
	FolloweeID int64 `json:"followee_id"`
}

// Handler reacts to decoded User events.
type Handler interface {
	OnRegistered(ctx context.Context, env Envelope, payload RegisteredPayload) error
	OnFollowed(ctx context.Context, env Envelope, payload FollowedPayload) error
	OnUnfollowed(ctx context.Context, env Envelope, payload UnfollowedPayload) error
}

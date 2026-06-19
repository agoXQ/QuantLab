// Package event defines domain events emitted by the User Service.
//
// The envelope mirrors the Strategy / Backtest / Market Data shape so
// downstream consumers (Notification, Ranking, Community) can decode
// every platform event with one schema.
package event

import (
	"context"
	"time"
)

// EventType lists the canonical User events.
type EventType string

const (
	EventUserRegistered EventType = "UserRegistered"
	EventUserUpdated    EventType = "UserUpdated"
	EventUserFollowed   EventType = "UserFollowed"
	EventUserUnfollowed EventType = "UserUnfollowed"
)

const (
	AggregateTypeUser = "USER"
	ProducerUser      = "user-service"
	TopicUserEvents   = "user-events"
	EventVersionV1    = "1.0"
)

// Event is the canonical envelope used by every User event.
type Event struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Producer      string    `json:"producer"`
	Payload       any       `json:"payload"`
}

// UserRegisteredPayload announces a new account. Notification subscribes
// to it for the welcome flow; Ranking seeds creator-tier defaults.
type UserRegisteredPayload struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// UserUpdatedPayload covers profile metadata edits (avatar / bio /
// nickname / location).
type UserUpdatedPayload struct {
	UserID int64 `json:"user_id"`
}

// UserFollowedPayload notifies that follower_id followed followee_id.
type UserFollowedPayload struct {
	FollowerID int64 `json:"follower_id"`
	FolloweeID int64 `json:"followee_id"`
}

// UserUnfollowedPayload is the inverse of UserFollowedPayload.
type UserUnfollowedPayload struct {
	FollowerID int64 `json:"follower_id"`
	FolloweeID int64 `json:"followee_id"`
}

// Publisher publishes User events.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

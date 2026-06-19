// Package subscription defines the NotificationSubscription aggregate.
// A subscription is the user-side opt-in that says "tell me when this
// object changes". The platform fan-out service later joins
// subscriptions with producer events to mint notifications, but the
// MVP only stores the rows.
package subscription

import (
	"context"
	"strings"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
)

// Subscription is the aggregate root.
type Subscription struct {
	ID           int64     `json:"id"`
	SubscriberID int64     `json:"subscriber_id"`
	ObjectType   string    `json:"object_type"`
	ObjectID     int64     `json:"object_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// MaxObjectTypeLength caps the discriminator length so the index stays
// compact.
const MaxObjectTypeLength = 64

// NormaliseObjectType lowercases / trims the object type so callers
// cannot create duplicate rows that differ only in casing.
func NormaliseObjectType(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Validate runs structural checks on the aggregate.
func (s *Subscription) Validate() error {
	if s == nil {
		return notifErr.ErrInvalidObjectType
	}
	if s.SubscriberID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	if s.ObjectID <= 0 {
		return notifErr.ErrInvalidObjectID
	}
	t := strings.TrimSpace(s.ObjectType)
	if t == "" || len(t) > MaxObjectTypeLength {
		return notifErr.ErrInvalidObjectType
	}
	return nil
}

// ListFilter scopes a list query.
type ListFilter struct {
	SubscriberID int64
	ObjectType   string
	Limit        int
	Offset       int
}

// Repository persists Subscription rows.
type Repository interface {
	Create(ctx context.Context, s *Subscription) error
	Delete(ctx context.Context, subscriberID, id int64) error
	Get(ctx context.Context, subscriberID, id int64) (*Subscription, error)
	List(ctx context.Context, filter ListFilter) ([]*Subscription, error)
	Count(ctx context.Context, filter ListFilter) (int64, error)
	ExistsByObject(ctx context.Context, subscriberID int64, objectType string, objectID int64) (bool, error)
	// ListSubscribers returns the user ids subscribed to the given
	// (object_type, object_id) pair so the fan-out path can mint one
	// notification per follower without paging twice.
	ListSubscribers(ctx context.Context, objectType string, objectID int64) ([]int64, error)
}

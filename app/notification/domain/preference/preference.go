// Package preference defines the NotificationPreference aggregate. The
// row is a per-user record describing which delivery channels the
// owner opted into; absence of a row means "platform defaults".
package preference

import (
	"context"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
)

// Preference is the aggregate root.
type Preference struct {
	UserID         int64     `json:"user_id"`
	InAppEnabled   bool      `json:"in_app_enabled"`
	EmailEnabled   bool      `json:"email_enabled"`
	WebhookEnabled bool      `json:"webhook_enabled"`
	PushEnabled    bool      `json:"push_enabled"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Defaults returns the platform-wide defaults applied to a brand-new
// user; in-app notifications are on, the rest are off until the owner
// opts in.
func Defaults(userID int64, now time.Time) Preference {
	return Preference{
		UserID:       userID,
		InAppEnabled: true,
		UpdatedAt:    now,
	}
}

// Validate runs structural checks on the aggregate.
func (p *Preference) Validate() error {
	if p == nil {
		return notifErr.ErrPreferenceNotFound
	}
	if p.UserID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	return nil
}

// Repository persists Preference rows.
type Repository interface {
	Get(ctx context.Context, userID int64) (*Preference, error)
	Upsert(ctx context.Context, p *Preference) error
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
)

// PreferenceRepository persists NotificationPreference rows.
type PreferenceRepository struct {
	db *sql.DB
}

// NewPreferenceRepository wires the repository.
func NewPreferenceRepository(db *sql.DB) *PreferenceRepository {
	return &PreferenceRepository{db: db}
}

// Get returns the user's preference row.
func (r *PreferenceRepository) Get(ctx context.Context, userID int64) (*domPref.Preference, error) {
	const stmt = `
		SELECT user_id, in_app_enabled, email_enabled, webhook_enabled, push_enabled, updated_at
		FROM notification_preference
		WHERE user_id = $1`
	var p domPref.Preference
	err := r.db.QueryRowContext(ctx, stmt, userID).Scan(
		&p.UserID,
		&p.InAppEnabled, &p.EmailEnabled, &p.WebhookEnabled, &p.PushEnabled,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notifErr.ErrPreferenceNotFound
		}
		return nil, fmt.Errorf("preference repository: get: %w", err)
	}
	return &p, nil
}

// Upsert writes the row, replacing any prior value.
func (r *PreferenceRepository) Upsert(ctx context.Context, p *domPref.Preference) error {
	if p == nil {
		return notifErr.ErrPreferenceNotFound
	}
	const stmt = `
		INSERT INTO notification_preference (
			user_id, in_app_enabled, email_enabled, webhook_enabled, push_enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			in_app_enabled  = EXCLUDED.in_app_enabled,
			email_enabled   = EXCLUDED.email_enabled,
			webhook_enabled = EXCLUDED.webhook_enabled,
			push_enabled    = EXCLUDED.push_enabled,
			updated_at      = EXCLUDED.updated_at`
	if _, err := r.db.ExecContext(ctx, stmt,
		p.UserID, p.InAppEnabled, p.EmailEnabled, p.WebhookEnabled, p.PushEnabled, p.UpdatedAt,
	); err != nil {
		return fmt.Errorf("preference repository: upsert: %w", err)
	}
	return nil
}

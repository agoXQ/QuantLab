package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
)

// SubscriptionRepository persists NotificationSubscription rows.
type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository wires the repository.
func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create inserts a row, mapping the unique-violation onto the
// SubscriptionConflict domain error.
func (r *SubscriptionRepository) Create(ctx context.Context, s *domSub.Subscription) error {
	if s == nil {
		return notifErr.ErrInvalidObjectType
	}
	const stmt = `
		INSERT INTO notification_subscription (subscriber_id, object_type, object_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	if err := r.db.QueryRowContext(ctx, stmt,
		s.SubscriberID, s.ObjectType, s.ObjectID, s.CreatedAt,
	).Scan(&s.ID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique") {
			return notifErr.ErrSubscriptionConflict
		}
		return fmt.Errorf("subscription repository: create: %w", err)
	}
	return nil
}

// Delete removes the row scoped by subscriber.
func (r *SubscriptionRepository) Delete(ctx context.Context, subscriberID, id int64) error {
	const stmt = `DELETE FROM notification_subscription WHERE id = $1 AND subscriber_id = $2`
	res, err := r.db.ExecContext(ctx, stmt, id, subscriberID)
	if err != nil {
		return fmt.Errorf("subscription repository: delete: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return notifErr.ErrSubscriptionNotFound
	}
	return nil
}

// Get fetches one row.
func (r *SubscriptionRepository) Get(ctx context.Context, subscriberID, id int64) (*domSub.Subscription, error) {
	const stmt = `
		SELECT id, subscriber_id, object_type, object_id, created_at
		FROM notification_subscription
		WHERE id = $1 AND subscriber_id = $2`
	row := r.db.QueryRowContext(ctx, stmt, id, subscriberID)
	var s domSub.Subscription
	if err := row.Scan(&s.ID, &s.SubscriberID, &s.ObjectType, &s.ObjectID, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notifErr.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("subscription repository: get: %w", err)
	}
	return &s, nil
}

// List filters by subscriber + optional object type and applies paging.
func (r *SubscriptionRepository) List(ctx context.Context, filter domSub.ListFilter) ([]*domSub.Subscription, error) {
	clauses := []string{"subscriber_id = $1"}
	args := []any{filter.SubscriberID}
	idx := 2
	if t := domSub.NormaliseObjectType(filter.ObjectType); t != "" {
		clauses = append(clauses, fmt.Sprintf("object_type = $%d", idx))
		args = append(args, t)
		idx++
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	stmt := fmt.Sprintf(`
		SELECT id, subscriber_id, object_type, object_id, created_at
		FROM notification_subscription
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(clauses, " AND "), idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("subscription repository: list: %w", err)
	}
	defer rows.Close()
	var out []*domSub.Subscription
	for rows.Next() {
		var s domSub.Subscription
		if err := rows.Scan(&s.ID, &s.SubscriberID, &s.ObjectType, &s.ObjectID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("subscription repository: scan: %w", err)
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

// Count returns the number of rows matching the filter.
func (r *SubscriptionRepository) Count(ctx context.Context, filter domSub.ListFilter) (int64, error) {
	clauses := []string{"subscriber_id = $1"}
	args := []any{filter.SubscriberID}
	idx := 2
	if t := domSub.NormaliseObjectType(filter.ObjectType); t != "" {
		clauses = append(clauses, fmt.Sprintf("object_type = $%d", idx))
		args = append(args, t)
		idx++
	}
	stmt := fmt.Sprintf(`SELECT COUNT(*) FROM notification_subscription WHERE %s`, strings.Join(clauses, " AND "))
	var n int64
	if err := r.db.QueryRowContext(ctx, stmt, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("subscription repository: count: %w", err)
	}
	return n, nil
}

// ExistsByObject probes the unique triple.
func (r *SubscriptionRepository) ExistsByObject(ctx context.Context, subscriberID int64, objectType string, objectID int64) (bool, error) {
	const stmt = `
		SELECT 1
		FROM notification_subscription
		WHERE subscriber_id = $1 AND object_type = $2 AND object_id = $3
		LIMIT 1`
	var dummy int
	err := r.db.QueryRowContext(ctx, stmt, subscriberID, domSub.NormaliseObjectType(objectType), objectID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("subscription repository: exists: %w", err)
	}
	return true, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// NotificationRepository persists Notification aggregates in Postgres.
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository wires the repository against an open *sql.DB.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a row, assigning the generated id back.
func (r *NotificationRepository) Create(ctx context.Context, n *domNotif.Notification) error {
	if n == nil {
		return notifErr.ErrInvalidNotification
	}
	const stmt = `
		INSERT INTO notification (user_id, type, title, content, status, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	if err := r.db.QueryRowContext(ctx, stmt,
		n.UserID, int32(n.Type), n.Title, n.Content,
		int32(n.Status), n.ReadAt, n.CreatedAt,
	).Scan(&n.ID); err != nil {
		return fmt.Errorf("notification repository: create: %w", err)
	}
	return nil
}

// Get fetches one row by user + id.
func (r *NotificationRepository) Get(ctx context.Context, userID, id int64) (*domNotif.Notification, error) {
	const stmt = `
		SELECT id, user_id, type, title, content, status, read_at, created_at
		FROM notification
		WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRowContext(ctx, stmt, id, userID)
	return scanNotification(row)
}

// List returns the rows matching the filter.
func (r *NotificationRepository) List(ctx context.Context, filter domNotif.ListFilter) ([]*domNotif.Notification, error) {
	clauses := []string{"user_id = $1"}
	args := []any{filter.UserID}
	idx := 2
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, 0, len(filter.Statuses))
		for _, s := range filter.Statuses {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, int32(s))
			idx++
		}
		clauses = append(clauses, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	} else {
		clauses = append(clauses, fmt.Sprintf("status <> %d", int32(valueobject.NotificationStatusDeleted)))
	}
	if len(filter.Types) > 0 {
		placeholders := make([]string, 0, len(filter.Types))
		for _, t := range filter.Types {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, int32(t))
			idx++
		}
		clauses = append(clauses, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
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
		SELECT id, user_id, type, title, content, status, read_at, created_at
		FROM notification
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(clauses, " AND "), idx, idx+1,
	)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("notification repository: list: %w", err)
	}
	defer rows.Close()
	var out []*domNotif.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// CountUnread returns the user's unread row count.
func (r *NotificationRepository) CountUnread(ctx context.Context, userID int64) (int64, error) {
	const stmt = `SELECT COUNT(*) FROM notification WHERE user_id = $1 AND status = $2`
	var n int64
	if err := r.db.QueryRowContext(ctx, stmt, userID, int32(valueobject.NotificationStatusUnread)).Scan(&n); err != nil {
		return 0, fmt.Errorf("notification repository: count unread: %w", err)
	}
	return n, nil
}

// MarkRead transitions the row to READ.
func (r *NotificationRepository) MarkRead(ctx context.Context, userID, id int64, now time.Time) error {
	const stmt = `
		UPDATE notification
		SET status = $1, read_at = $2
		WHERE id = $3 AND user_id = $4 AND status <> $5`
	res, err := r.db.ExecContext(ctx, stmt,
		int32(valueobject.NotificationStatusRead), now, id, userID,
		int32(valueobject.NotificationStatusDeleted),
	)
	if err != nil {
		return fmt.Errorf("notification repository: mark read: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		// Either the row is gone or already deleted; differentiate so
		// callers can react appropriately.
		var status int32
		err := r.db.QueryRowContext(ctx, `SELECT status FROM notification WHERE id = $1 AND user_id = $2`, id, userID).Scan(&status)
		if errors.Is(err, sql.ErrNoRows) {
			return notifErr.ErrNotificationNotFound
		}
		if err != nil {
			return fmt.Errorf("notification repository: mark read recheck: %w", err)
		}
		if valueobject.NotificationStatus(status) == valueobject.NotificationStatusDeleted {
			return notifErr.ErrAlreadyDeleted
		}
		// Already READ; idempotent success.
		return nil
	}
	return nil
}

// MarkAllRead transitions every UNREAD row of the user to READ.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID int64, now time.Time) (int64, error) {
	const stmt = `
		UPDATE notification
		SET status = $1, read_at = $2
		WHERE user_id = $3 AND status = $4`
	res, err := r.db.ExecContext(ctx, stmt,
		int32(valueobject.NotificationStatusRead), now, userID,
		int32(valueobject.NotificationStatusUnread),
	)
	if err != nil {
		return 0, fmt.Errorf("notification repository: mark all read: %w", err)
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// Delete soft-deletes the row.
func (r *NotificationRepository) Delete(ctx context.Context, userID, id int64) error {
	const stmt = `
		UPDATE notification
		SET status = $1
		WHERE id = $2 AND user_id = $3`
	res, err := r.db.ExecContext(ctx, stmt, int32(valueobject.NotificationStatusDeleted), id, userID)
	if err != nil {
		return fmt.Errorf("notification repository: delete: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return notifErr.ErrNotificationNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNotification(row rowScanner) (*domNotif.Notification, error) {
	var (
		n        domNotif.Notification
		typeRaw  int32
		statRaw  int32
		readAt   sql.NullTime
	)
	if err := row.Scan(&n.ID, &n.UserID, &typeRaw, &n.Title, &n.Content, &statRaw, &readAt, &n.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notifErr.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("notification repository: scan: %w", err)
	}
	n.Type = valueobject.NotificationType(typeRaw)
	n.Status = valueobject.NotificationStatus(statRaw)
	if readAt.Valid {
		t := readAt.Time
		n.ReadAt = &t
	}
	return &n, nil
}

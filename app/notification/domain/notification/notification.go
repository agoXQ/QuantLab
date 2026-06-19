// Package notification defines the Notification aggregate.
//
// A Notification is an immutable platform message addressed to a
// specific user. Mutations are limited to the read / deleted
// transitions; the body itself is set at creation time and never
// rewritten so audit trails stay simple.
package notification

import (
	"context"
	"strings"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// MaxTitleLength caps the title at the schema limit so a renderer can
// rely on a single line of text.
const MaxTitleLength = 256

// MaxContentLength caps the body so a misbehaving producer cannot fill
// the table with multi-megabyte rows.
const MaxContentLength = 4096

// Notification is the aggregate root.
type Notification struct {
	ID        int64                          `json:"id"`
	UserID    int64                          `json:"user_id"`
	Type      valueobject.NotificationType   `json:"type"`
	Title     string                         `json:"title"`
	Content   string                         `json:"content"`
	Status    valueobject.NotificationStatus `json:"status"`
	ReadAt    *time.Time                     `json:"read_at,omitempty"`
	CreatedAt time.Time                      `json:"created_at"`
}

// Validate runs structural checks. The repository refuses rows the
// rest of the system cannot trust.
func (n *Notification) Validate() error {
	if n == nil {
		return notifErr.ErrInvalidNotification
	}
	if n.UserID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	if !n.Type.IsValid() {
		return notifErr.ErrInvalidType
	}
	if !n.Status.IsValid() {
		return notifErr.ErrInvalidStatus
	}
	if strings.TrimSpace(n.Title) == "" || len(n.Title) > MaxTitleLength {
		return notifErr.ErrInvalidNotification
	}
	if len(n.Content) > MaxContentLength {
		return notifErr.ErrInvalidNotification
	}
	return nil
}

// MarkRead transitions an UNREAD row to READ. Replays on an already
// read row are a no-op so an idempotent consumer never observes a
// state-machine refusal.
func (n *Notification) MarkRead(now time.Time) {
	if n == nil || n.Status == valueobject.NotificationStatusRead {
		return
	}
	if n.Status == valueobject.NotificationStatusDeleted {
		return
	}
	n.Status = valueobject.NotificationStatusRead
	t := now
	n.ReadAt = &t
}

// MarkDeleted transitions a row to DELETED. Idempotent on an already
// deleted row for the same reason as MarkRead.
func (n *Notification) MarkDeleted() {
	if n == nil {
		return
	}
	n.Status = valueobject.NotificationStatusDeleted
}

// ListFilter is the query shape the application service passes to
// Repository.List so a single SQL builder can serve every list view.
type ListFilter struct {
	UserID   int64
	Statuses []valueobject.NotificationStatus
	Types    []valueobject.NotificationType
	Limit    int
	Offset   int
}

// Repository persists Notification rows. Implementations must respect
// per-user isolation: every read / write is keyed on the caller's
// user id so the same Postgres table can serve every tenant safely.
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	Get(ctx context.Context, userID, id int64) (*Notification, error)
	List(ctx context.Context, filter ListFilter) ([]*Notification, error)
	CountUnread(ctx context.Context, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID, id int64, now time.Time) error
	MarkAllRead(ctx context.Context, userID int64, now time.Time) (int64, error)
	Delete(ctx context.Context, userID, id int64) error
}

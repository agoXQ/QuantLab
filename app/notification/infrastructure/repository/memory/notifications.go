// Package memory provides in-memory implementations of the Notification
// repository ports. They keep the binary usable without a database (CI
// smoke tests, local exploration) and back the unit tests.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// NotificationRepository is the in-memory Notification repository.
type NotificationRepository struct {
	mu     sync.RWMutex
	rows   map[int64]*domNotif.Notification
	nextID int64
}

// NewNotificationRepository returns an empty repo.
func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{rows: map[int64]*domNotif.Notification{}}
}

func cloneNotification(n *domNotif.Notification) *domNotif.Notification {
	if n == nil {
		return nil
	}
	cp := *n
	if n.ReadAt != nil {
		t := *n.ReadAt
		cp.ReadAt = &t
	}
	return &cp
}

// Create inserts a row and assigns the generated id.
func (r *NotificationRepository) Create(_ context.Context, n *domNotif.Notification) error {
	if n == nil {
		return notifErr.ErrInvalidNotification
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	n.ID = r.nextID
	r.rows[n.ID] = cloneNotification(n)
	return nil
}

// Get fetches one row by user + id.
func (r *NotificationRepository) Get(_ context.Context, userID, id int64) (*domNotif.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.rows[id]
	if !ok || row.UserID != userID {
		return nil, notifErr.ErrNotificationNotFound
	}
	return cloneNotification(row), nil
}

// List filters rows by status / type and applies offset+limit paging.
func (r *NotificationRepository) List(_ context.Context, filter domNotif.ListFilter) ([]*domNotif.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	statusSet := map[valueobject.NotificationStatus]struct{}{}
	for _, s := range filter.Statuses {
		statusSet[s] = struct{}{}
	}
	typeSet := map[valueobject.NotificationType]struct{}{}
	for _, t := range filter.Types {
		typeSet[t] = struct{}{}
	}
	var pool []*domNotif.Notification
	for _, row := range r.rows {
		if row.UserID != filter.UserID {
			continue
		}
		if len(statusSet) > 0 {
			if _, ok := statusSet[row.Status]; !ok {
				continue
			}
		} else if row.Status == valueobject.NotificationStatusDeleted {
			continue
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[row.Type]; !ok {
				continue
			}
		}
		pool = append(pool, cloneNotification(row))
	}
	sort.Slice(pool, func(i, j int) bool {
		if !pool[i].CreatedAt.Equal(pool[j].CreatedAt) {
			return pool[i].CreatedAt.After(pool[j].CreatedAt)
		}
		return pool[i].ID > pool[j].ID
	})
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(pool) {
		return nil, nil
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = len(pool) - offset
	}
	end := offset + limit
	if end > len(pool) {
		end = len(pool)
	}
	return pool[offset:end], nil
}

// CountUnread returns the number of UNREAD rows the user owns.
func (r *NotificationRepository) CountUnread(_ context.Context, userID int64) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var n int64
	for _, row := range r.rows {
		if row.UserID == userID && row.Status == valueobject.NotificationStatusUnread {
			n++
		}
	}
	return n, nil
}

// MarkRead transitions one row to READ. Idempotent; an already read
// row returns nil so retries do not surface as conflicts.
func (r *NotificationRepository) MarkRead(_ context.Context, userID, id int64, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.rows[id]
	if !ok || row.UserID != userID {
		return notifErr.ErrNotificationNotFound
	}
	if row.Status == valueobject.NotificationStatusDeleted {
		return notifErr.ErrAlreadyDeleted
	}
	if row.Status == valueobject.NotificationStatusRead {
		return nil
	}
	row.Status = valueobject.NotificationStatusRead
	t := now
	row.ReadAt = &t
	return nil
}

// MarkAllRead transitions every UNREAD row to READ and returns the
// number of rows touched.
func (r *NotificationRepository) MarkAllRead(_ context.Context, userID int64, now time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for _, row := range r.rows {
		if row.UserID != userID || row.Status != valueobject.NotificationStatusUnread {
			continue
		}
		row.Status = valueobject.NotificationStatusRead
		t := now
		row.ReadAt = &t
		n++
	}
	return n, nil
}

// Delete soft-deletes one row.
func (r *NotificationRepository) Delete(_ context.Context, userID, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.rows[id]
	if !ok || row.UserID != userID {
		return notifErr.ErrNotificationNotFound
	}
	if row.Status == valueobject.NotificationStatusDeleted {
		return nil
	}
	row.Status = valueobject.NotificationStatusDeleted
	return nil
}

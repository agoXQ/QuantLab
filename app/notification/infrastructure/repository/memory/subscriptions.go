package memory

import (
	"context"
	"sort"
	"sync"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
)

// SubscriptionRepository is the in-memory NotificationSubscription repo.
type SubscriptionRepository struct {
	mu     sync.RWMutex
	rows   map[int64]*domSub.Subscription
	nextID int64
}

// NewSubscriptionRepository returns an empty repo.
func NewSubscriptionRepository() *SubscriptionRepository {
	return &SubscriptionRepository{rows: map[int64]*domSub.Subscription{}}
}

// Create inserts a row and assigns the generated id. Duplicates on
// (subscriber_id, object_type, object_id) are rejected so the
// uniqueness invariant matches the Postgres index.
func (r *SubscriptionRepository) Create(_ context.Context, s *domSub.Subscription) error {
	if s == nil {
		return notifErr.ErrInvalidObjectType
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, row := range r.rows {
		if row.SubscriberID == s.SubscriberID && row.ObjectType == s.ObjectType && row.ObjectID == s.ObjectID {
			return notifErr.ErrSubscriptionConflict
		}
	}
	r.nextID++
	s.ID = r.nextID
	cp := *s
	r.rows[s.ID] = &cp
	return nil
}

// Delete removes the row.
func (r *SubscriptionRepository) Delete(_ context.Context, subscriberID, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.rows[id]
	if !ok || row.SubscriberID != subscriberID {
		return notifErr.ErrSubscriptionNotFound
	}
	delete(r.rows, id)
	return nil
}

// Get fetches the row by subscriber + id.
func (r *SubscriptionRepository) Get(_ context.Context, subscriberID, id int64) (*domSub.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.rows[id]
	if !ok || row.SubscriberID != subscriberID {
		return nil, notifErr.ErrSubscriptionNotFound
	}
	cp := *row
	return &cp, nil
}

// List filters by subscriber + optional object type and applies paging.
func (r *SubscriptionRepository) List(_ context.Context, filter domSub.ListFilter) ([]*domSub.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pool []*domSub.Subscription
	objType := domSub.NormaliseObjectType(filter.ObjectType)
	for _, row := range r.rows {
		if row.SubscriberID != filter.SubscriberID {
			continue
		}
		if objType != "" && row.ObjectType != objType {
			continue
		}
		cp := *row
		pool = append(pool, &cp)
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

// Count returns the number of rows matching the filter.
func (r *SubscriptionRepository) Count(_ context.Context, filter domSub.ListFilter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	objType := domSub.NormaliseObjectType(filter.ObjectType)
	var n int64
	for _, row := range r.rows {
		if row.SubscriberID != filter.SubscriberID {
			continue
		}
		if objType != "" && row.ObjectType != objType {
			continue
		}
		n++
	}
	return n, nil
}

// ExistsByObject checks the (subscriber, object_type, object_id)
// triple so the application service can fail fast on duplicates.
func (r *SubscriptionRepository) ExistsByObject(_ context.Context, subscriberID int64, objectType string, objectID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	objType := domSub.NormaliseObjectType(objectType)
	for _, row := range r.rows {
		if row.SubscriberID == subscriberID && row.ObjectType == objType && row.ObjectID == objectID {
			return true, nil
		}
	}
	return false, nil
}


// ListSubscribers returns the deduplicated subscriber ids opted into
// (object_type, object_id). The result is stable to make it test-
// friendly: subscribers are sorted ascending so callers can compare
// snapshots without first sorting on their side.
func (r *SubscriptionRepository) ListSubscribers(_ context.Context, objectType string, objectID int64) ([]int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	objType := domSub.NormaliseObjectType(objectType)
	if objType == "" || objectID <= 0 {
		return nil, nil
	}
	seen := map[int64]struct{}{}
	var out []int64
	for _, row := range r.rows {
		if row.ObjectType != objType || row.ObjectID != objectID {
			continue
		}
		if _, ok := seen[row.SubscriberID]; ok {
			continue
		}
		seen[row.SubscriberID] = struct{}{}
		out = append(out, row.SubscriberID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

package memory

import (
	"context"
	"sync"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
)

// PreferenceRepository is the in-memory NotificationPreference repo.
type PreferenceRepository struct {
	mu   sync.RWMutex
	rows map[int64]*domPref.Preference
}

// NewPreferenceRepository returns an empty repo.
func NewPreferenceRepository() *PreferenceRepository {
	return &PreferenceRepository{rows: map[int64]*domPref.Preference{}}
}

// Get returns the row keyed by user id; ErrPreferenceNotFound when
// the row is absent so the application service can fall back to the
// platform defaults without surfacing a 500.
func (r *PreferenceRepository) Get(_ context.Context, userID int64) (*domPref.Preference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.rows[userID]
	if !ok {
		return nil, notifErr.ErrPreferenceNotFound
	}
	cp := *row
	return &cp, nil
}

// Upsert writes the row, replacing any prior value.
func (r *PreferenceRepository) Upsert(_ context.Context, p *domPref.Preference) error {
	if p == nil {
		return notifErr.ErrPreferenceNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.rows[p.UserID] = &cp
	return nil
}

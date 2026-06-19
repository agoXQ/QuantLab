// Package memory provides in-memory implementations of the User Service
// repository ports. They keep the binary usable without a database
// (CI smoke tests, local exploration) and back the unit tests.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
)

// UserRepository is the in-memory User repository. Email and username
// uniqueness are enforced via secondary indexes so the application
// service can reject conflicts before any disk persistence is wired.
type UserRepository struct {
	mu       sync.RWMutex
	data     map[int64]*domuser.User
	byEmail  map[string]int64
	byName   map[string]int64
	nextID   int64
}

// NewUserRepository returns an empty in-memory repo.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		data:    map[int64]*domuser.User{},
		byEmail: map[string]int64{},
		byName:  map[string]int64{},
	}
}

// Create inserts the user, assigning the generated id back, and updates
// the email / username indexes.
func (r *UserRepository) Create(_ context.Context, u *domuser.User) error {
	if u == nil {
		return userErr.ErrInvalidUser
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	emailKey := strings.ToLower(strings.TrimSpace(u.Email))
	nameKey := strings.ToLower(strings.TrimSpace(u.Username))
	if _, ok := r.byEmail[emailKey]; ok {
		return userErr.ErrEmailTaken
	}
	if _, ok := r.byName[nameKey]; ok {
		return userErr.ErrUsernameTaken
	}
	r.nextID++
	u.ID = r.nextID
	cp := *u
	r.data[u.ID] = &cp
	r.byEmail[emailKey] = u.ID
	r.byName[nameKey] = u.ID
	return nil
}

// Update writes back the mutable fields. The indexes are rebuilt so
// callers can rename / change email; conflicts surface as the same
// errors as Create would have raised.
func (r *UserRepository) Update(_ context.Context, u *domuser.User) error {
	if u == nil || u.ID == 0 {
		return userErr.ErrInvalidUser
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	prev, ok := r.data[u.ID]
	if !ok {
		return userErr.ErrUserNotFound
	}
	emailKey := strings.ToLower(strings.TrimSpace(u.Email))
	nameKey := strings.ToLower(strings.TrimSpace(u.Username))
	if id, exists := r.byEmail[emailKey]; exists && id != u.ID {
		return userErr.ErrEmailTaken
	}
	if id, exists := r.byName[nameKey]; exists && id != u.ID {
		return userErr.ErrUsernameTaken
	}
	delete(r.byEmail, strings.ToLower(strings.TrimSpace(prev.Email)))
	delete(r.byName, strings.ToLower(strings.TrimSpace(prev.Username)))
	cp := *u
	r.data[u.ID] = &cp
	r.byEmail[emailKey] = u.ID
	r.byName[nameKey] = u.ID
	return nil
}

// Get loads one user.
func (r *UserRepository) Get(_ context.Context, id int64) (*domuser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.data[id]
	if !ok {
		return nil, userErr.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

// GetByEmail returns the user by email or ErrUserNotFound.
func (r *UserRepository) GetByEmail(_ context.Context, email string) (*domuser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, userErr.ErrUserNotFound
	}
	cp := *r.data[id]
	return &cp, nil
}

// GetByUsername returns the user by username or ErrUserNotFound.
func (r *UserRepository) GetByUsername(_ context.Context, username string) (*domuser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byName[strings.ToLower(strings.TrimSpace(username))]
	if !ok {
		return nil, userErr.ErrUserNotFound
	}
	cp := *r.data[id]
	return &cp, nil
}

// FollowRepository is the in-memory Follow repository. (follower,
// followee) pairs are unique; the implementation enforces that with a
// composite key map.
type FollowRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domfollow.Follow
	pairs  map[string]int64
	nextID int64
}

// NewFollowRepository returns an empty in-memory follow repo.
func NewFollowRepository() *FollowRepository {
	return &FollowRepository{
		data:  map[int64]*domfollow.Follow{},
		pairs: map[string]int64{},
	}
}

func pairKey(followerID, followeeID int64) string {
	var b strings.Builder
	b.Grow(48)
	b.WriteString(itoa(followerID))
	b.WriteString(":")
	b.WriteString(itoa(followeeID))
	return b.String()
}

func itoa(v int64) string {
	// strconv.FormatInt avoids fmt allocations on the hot path; the
	// helper stays unexported because the pair key is the only call site.
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	negative := v < 0
	if negative {
		v = -v
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%10]
		v /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func (r *FollowRepository) Create(_ context.Context, f *domfollow.Follow) error {
	if f == nil {
		return userErr.ErrInvalidUser
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pairKey(f.FollowerID, f.FolloweeID)
	if _, ok := r.pairs[key]; ok {
		return userErr.ErrAlreadyFollowed
	}
	r.nextID++
	f.ID = r.nextID
	cp := *f
	r.data[f.ID] = &cp
	r.pairs[key] = f.ID
	return nil
}

func (r *FollowRepository) Delete(_ context.Context, followerID, followeeID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pairKey(followerID, followeeID)
	id, ok := r.pairs[key]
	if !ok {
		return userErr.ErrFollowNotFound
	}
	delete(r.pairs, key)
	delete(r.data, id)
	return nil
}

func (r *FollowRepository) Exists(_ context.Context, followerID, followeeID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.pairs[pairKey(followerID, followeeID)]
	return ok, nil
}

func (r *FollowRepository) ListFollowers(_ context.Context, userID int64, limit, offset int) ([]*domfollow.Follow, error) {
	return r.list(func(f *domfollow.Follow) bool { return f.FolloweeID == userID }, limit, offset), nil
}

func (r *FollowRepository) ListFollowing(_ context.Context, userID int64, limit, offset int) ([]*domfollow.Follow, error) {
	return r.list(func(f *domfollow.Follow) bool { return f.FollowerID == userID }, limit, offset), nil
}

func (r *FollowRepository) CountFollowers(_ context.Context, userID int64) (int64, error) {
	return r.count(func(f *domfollow.Follow) bool { return f.FolloweeID == userID }), nil
}

func (r *FollowRepository) CountFollowing(_ context.Context, userID int64) (int64, error) {
	return r.count(func(f *domfollow.Follow) bool { return f.FollowerID == userID }), nil
}

func (r *FollowRepository) list(match func(*domfollow.Follow) bool, limit, offset int) []*domfollow.Follow {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domfollow.Follow, 0)
	for _, f := range r.data {
		if match(f) {
			cp := *f
			out = append(out, &cp)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if offset > 0 && offset < len(out) {
		out = out[offset:]
	} else if offset >= len(out) {
		return nil
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (r *FollowRepository) count(match func(*domfollow.Follow) bool) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var n int64
	for _, f := range r.data {
		if match(f) {
			n++
		}
	}
	return n
}

// Package memory provides in-memory implementations of the Strategy
// Service repository ports. They keep the binary usable without a
// database (CI smoke tests, local exploration) and double as the
// substrate for unit tests.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// StrategyRepository is the in-memory Strategy repository.
type StrategyRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domstrategy.Strategy
	nextID int64
}

// NewStrategyRepository returns an empty in-memory repo.
func NewStrategyRepository() *StrategyRepository {
	return &StrategyRepository{data: map[int64]*domstrategy.Strategy{}}
}

func (r *StrategyRepository) Create(_ context.Context, s *domstrategy.Strategy) error {
	if s == nil {
		return stratErr.ErrInvalidStrategy
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	s.ID = r.nextID
	cp := *s
	r.data[s.ID] = &cp
	return nil
}

func (r *StrategyRepository) Update(_ context.Context, s *domstrategy.Strategy) error {
	if s == nil || s.ID == 0 {
		return stratErr.ErrInvalidStrategy
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[s.ID]; !ok {
		return stratErr.ErrStrategyNotFound
	}
	cp := *s
	r.data[s.ID] = &cp
	return nil
}

func (r *StrategyRepository) Get(_ context.Context, id int64) (*domstrategy.Strategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.data[id]
	if !ok {
		return nil, stratErr.ErrStrategyNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *StrategyRepository) List(_ context.Context, q domstrategy.ListQuery) ([]*domstrategy.Strategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domstrategy.Strategy, 0, len(r.data))
	for _, s := range r.data {
		if !matches(s, q) {
			continue
		}
		cp := *s
		out = append(out, &cp)
	}
	sortStrategies(out, q.Sort)
	if q.Offset > 0 && q.Offset < len(out) {
		out = out[q.Offset:]
	} else if q.Offset >= len(out) {
		out = nil
	}
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

func (r *StrategyRepository) IncrementForkCount(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.data[id]
	if !ok {
		return stratErr.ErrStrategyNotFound
	}
	s.ForkCount++
	return nil
}

func matches(s *domstrategy.Strategy, q domstrategy.ListQuery) bool {
	if q.AuthorID != 0 && s.AuthorID != q.AuthorID {
		return false
	}
	if q.Status != "" && s.Status != q.Status {
		return false
	}
	if q.Visibility != "" && s.Visibility != q.Visibility {
		return false
	}
	if q.Category != "" && !strings.EqualFold(s.Category, q.Category) {
		return false
	}
	if q.Tag != "" {
		needle := strings.ToLower(strings.TrimSpace(q.Tag))
		found := false
		for _, t := range s.Tags {
			if t == needle {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if q.Keyword != "" {
		needle := strings.ToLower(q.Keyword)
		if !strings.Contains(strings.ToLower(s.Title), needle) &&
			!strings.Contains(strings.ToLower(s.Description), needle) {
			return false
		}
	}
	return true
}

func sortStrategies(in []*domstrategy.Strategy, sortKey string) {
	switch sortKey {
	case "favorite_count_desc":
		sort.SliceStable(in, func(i, j int) bool { return in[i].FavoriteCount > in[j].FavoriteCount })
	case "fork_count_desc":
		sort.SliceStable(in, func(i, j int) bool { return in[i].ForkCount > in[j].ForkCount })
	case "view_count_desc":
		sort.SliceStable(in, func(i, j int) bool { return in[i].ViewCount > in[j].ViewCount })
	default:
		sort.SliceStable(in, func(i, j int) bool { return in[i].CreatedAt.After(in[j].CreatedAt) })
	}
}

// VersionRepository is the in-memory StrategyVersion repository.
type VersionRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domversion.StrategyVersion
	nextID int64
}

func NewVersionRepository() *VersionRepository {
	return &VersionRepository{data: map[int64]*domversion.StrategyVersion{}}
}

func (r *VersionRepository) Create(_ context.Context, v *domversion.StrategyVersion) error {
	if v == nil {
		return stratErr.ErrInvalidVersion
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	v.ID = r.nextID
	cp := *v
	r.data[v.ID] = &cp
	return nil
}

func (r *VersionRepository) Get(_ context.Context, id int64) (*domversion.StrategyVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[id]
	if !ok {
		return nil, stratErr.ErrVersionNotFound
	}
	cp := *v
	return &cp, nil
}

func (r *VersionRepository) ListByStrategy(_ context.Context, strategyID int64, limit int) ([]*domversion.StrategyVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domversion.StrategyVersion, 0)
	for _, v := range r.data {
		if v.StrategyID == strategyID {
			cp := *v
			out = append(out, &cp)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *VersionRepository) LatestNumber(_ context.Context, strategyID int64) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, v := range r.data {
		if v.StrategyID == strategyID {
			count++
		}
	}
	return count, nil
}

// ForkRepository is the in-memory StrategyFork repository.
type ForkRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domfork.StrategyFork
	nextID int64
}

func NewForkRepository() *ForkRepository {
	return &ForkRepository{data: map[int64]*domfork.StrategyFork{}}
}

func (r *ForkRepository) Create(_ context.Context, f *domfork.StrategyFork) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	f.ID = r.nextID
	cp := *f
	r.data[f.ID] = &cp
	return nil
}

func (r *ForkRepository) ListBySource(_ context.Context, sourceID int64, limit int) ([]*domfork.StrategyFork, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domfork.StrategyFork, 0)
	for _, f := range r.data {
		if f.SourceStrategyID == sourceID {
			cp := *f
			out = append(out, &cp)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// silence: imports valueobject only when needed for status filtering;
// keep the reference here so the package compiles even when callers do
// not exercise the Status filter directly.
var _ = valueobject.LifecycleStatusDraft

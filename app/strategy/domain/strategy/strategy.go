// Package strategy defines the Strategy aggregate.
//
// Strategy is the aggregate root for every user-owned research asset on
// the platform. The aggregate owns its lifecycle (Draft / Configured /
// Backtested / Published / Archived), basic metadata (title, tags,
// category, visibility), and a pointer to the current version. Heavy
// state — the version body itself — lives in a sibling aggregate so a
// metadata update never rewrites the formula payload.
package strategy

import (
	"context"
	"strings"
	"time"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
)

// Strategy is the aggregate root.
type Strategy struct {
	ID               int64                    `json:"id"`
	AuthorID         int64                    `json:"author_id"`
	Title            string                   `json:"title"`
	Description      string                   `json:"description,omitempty"`
	Category         string                   `json:"category,omitempty"`
	Tags             []string                 `json:"tags,omitempty"`
	CurrentVersionID int64                    `json:"current_version_id,omitempty"`
	Status           valueobject.LifecycleStatus `json:"status"`
	Visibility       valueobject.Visibility    `json:"visibility"`
	// Source records the strategy this one was forked from (zero when
	// the strategy is original). Forwarded into events so Ranking can
	// build the lineage tree.
	SourceStrategyID int64     `json:"source_strategy_id,omitempty"`
	ViewCount        int64     `json:"view_count"`
	FavoriteCount    int64     `json:"favorite_count"`
	ForkCount        int64     `json:"fork_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
}

// MaxTitleLength caps the title at a reasonable ceiling so a renderer
// never has to truncate; aligns with the SQL VARCHAR(256) cap.
const MaxTitleLength = 256

// Validate runs structural checks. Body / formula validation lives on
// the StrategyVersion aggregate so an empty Draft is allowed here.
func (s *Strategy) Validate() error {
	if s == nil {
		return stratErr.ErrInvalidStrategy
	}
	if strings.TrimSpace(s.Title) == "" {
		return stratErr.ErrTitleRequired
	}
	if len(s.Title) > MaxTitleLength {
		return stratErr.ErrTitleTooLong
	}
	if !s.Visibility.IsValid() {
		return stratErr.ErrInvalidVisibility
	}
	if !s.Status.IsValid() {
		return stratErr.ErrInvalidStatus
	}
	return nil
}

// MarkConfigured records that the strategy now has a runnable version.
// The transition is idempotent from Configured (re-saving a version
// keeps the strategy in the same state) and accepts Draft as the
// originating state. Backtested / Published strategies keep their
// status because Configured is a strictly less specific predicate.
func (s *Strategy) MarkConfigured(now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return stratErr.ErrStrategyArchived
	}
	switch s.Status {
	case valueobject.LifecycleStatusDraft, valueobject.LifecycleStatusConfigured:
		s.Status = valueobject.LifecycleStatusConfigured
		s.UpdatedAt = now
	}
	return nil
}

// MarkBacktested records the first successful run. Idempotent across
// repeated runs; the timestamp on UpdatedAt advances each call.
func (s *Strategy) MarkBacktested(now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return stratErr.ErrStrategyArchived
	}
	if s.Status == valueobject.LifecycleStatusPublished {
		s.UpdatedAt = now
		return nil
	}
	s.Status = valueobject.LifecycleStatusBacktested
	s.UpdatedAt = now
	return nil
}

// Publish flips the strategy public against a specific version. The
// caller is responsible for verifying the version belongs to this
// strategy; the aggregate only enforces lifecycle invariants and
// version presence.
func (s *Strategy) Publish(versionID int64, now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return stratErr.ErrStrategyArchived
	}
	if versionID <= 0 || s.CurrentVersionID == 0 {
		return stratErr.ErrPublishWithoutVersion
	}
	if s.Status == valueobject.LifecycleStatusPublished && s.CurrentVersionID == versionID {
		return stratErr.ErrAlreadyPublished
	}
	s.Status = valueobject.LifecycleStatusPublished
	s.Visibility = valueobject.VisibilityPublic
	t := now
	s.PublishedAt = &t
	s.UpdatedAt = now
	return nil
}

// Archive flips the strategy off the public surface for good. Idempotent
// from Archived.
func (s *Strategy) Archive(now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return nil
	}
	s.Status = valueobject.LifecycleStatusArchived
	s.Visibility = valueobject.VisibilityPrivate
	t := now
	s.ArchivedAt = &t
	s.UpdatedAt = now
	return nil
}

// UpdateMeta applies a metadata patch produced by the API layer. Empty
// strings keep the existing field (tags is treated as "set" only when
// the slice is non-nil; pass an empty []string{} to clear).
func (s *Strategy) UpdateMeta(patch MetaPatch, now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return stratErr.ErrStrategyArchived
	}
	if patch.Title != nil {
		t := strings.TrimSpace(*patch.Title)
		if t == "" {
			return stratErr.ErrTitleRequired
		}
		if len(t) > MaxTitleLength {
			return stratErr.ErrTitleTooLong
		}
		s.Title = t
	}
	if patch.Description != nil {
		s.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.Category != nil {
		s.Category = strings.TrimSpace(*patch.Category)
	}
	if patch.Tags != nil {
		s.Tags = NormaliseTags(*patch.Tags)
	}
	if patch.Visibility != nil {
		if !patch.Visibility.IsValid() {
			return stratErr.ErrInvalidVisibility
		}
		// Archived / Published-bound visibility flips are still allowed
		// here so a user can hide a previously-public strategy without
		// archiving it. The invariant is only that visibility never
		// becomes Public on an Archived row, which the early-return
		// above already enforces.
		s.Visibility = *patch.Visibility
	}
	s.UpdatedAt = now
	return nil
}

// AttachVersion records that versionID is the new "current" version.
// The aggregate stores the pointer so reads never have to query the
// version table to render the latest snapshot. The caller still inserts
// the version row in the same transaction.
func (s *Strategy) AttachVersion(versionID int64, now time.Time) error {
	if s.Status == valueobject.LifecycleStatusArchived {
		return stratErr.ErrStrategyArchived
	}
	if versionID <= 0 {
		return stratErr.ErrInvalidVersion
	}
	s.CurrentVersionID = versionID
	if s.Status == valueobject.LifecycleStatusDraft {
		s.Status = valueobject.LifecycleStatusConfigured
	}
	s.UpdatedAt = now
	return nil
}

// MetaPatch carries an optional metadata patch. Pointer fields keep
// "leave as-is" semantically distinct from "set to empty string".
type MetaPatch struct {
	Title       *string
	Description *string
	Category    *string
	Tags        *[]string
	Visibility  *valueobject.Visibility
}

// NormaliseTags trims, lowercases, deduplicates, and orders tags so the
// persisted slice is canonical regardless of caller input.
func NormaliseTags(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// Repository persists Strategy aggregates.
type Repository interface {
	Create(ctx context.Context, s *Strategy) error
	Update(ctx context.Context, s *Strategy) error
	Get(ctx context.Context, id int64) (*Strategy, error)
	List(ctx context.Context, q ListQuery) ([]*Strategy, error)
	IncrementForkCount(ctx context.Context, id int64) error
}

// ListQuery is the coarse filter used by list / search endpoints. The
// MVP keeps it straightforward; ElasticSearch-backed full-text search is
// out of scope here and slots in behind the same interface later.
type ListQuery struct {
	AuthorID   int64
	Status     valueobject.LifecycleStatus
	Visibility valueobject.Visibility
	Category   string
	Tag        string
	Keyword    string
	// Sort is one of: "created_at_desc" (default), "favorite_count_desc",
	// "fork_count_desc", "view_count_desc". Unknown values fall back to
	// the default to keep callers immune to typos.
	Sort   string
	Limit  int
	Offset int
}

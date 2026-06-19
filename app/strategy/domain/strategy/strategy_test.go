package strategy

import (
	"testing"
	"time"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
)

// TestStateMachine guards Draft -> Configured -> Backtested -> Published
// -> Archived and the related guards. The application service relies on
// these transitions being enforced by the aggregate, not by ad-hoc
// checks in the use case layer.
func TestStateMachine(t *testing.T) {
	now := time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC)

	t.Run("draft -> configured via AttachVersion", func(t *testing.T) {
		s := &Strategy{Status: valueobject.LifecycleStatusDraft, Visibility: valueobject.VisibilityPrivate}
		if err := s.AttachVersion(42, now); err != nil {
			t.Fatalf("AttachVersion: %v", err)
		}
		if s.Status != valueobject.LifecycleStatusConfigured {
			t.Fatalf("expected CONFIGURED, got %s", s.Status)
		}
		if s.CurrentVersionID != 42 {
			t.Fatalf("expected CurrentVersionID=42, got %d", s.CurrentVersionID)
		}
	})

	t.Run("configured -> backtested", func(t *testing.T) {
		s := &Strategy{Status: valueobject.LifecycleStatusConfigured}
		if err := s.MarkBacktested(now); err != nil {
			t.Fatalf("MarkBacktested: %v", err)
		}
		if s.Status != valueobject.LifecycleStatusBacktested {
			t.Fatalf("expected BACKTESTED, got %s", s.Status)
		}
	})

	t.Run("publish requires a current version", func(t *testing.T) {
		s := &Strategy{Status: valueobject.LifecycleStatusConfigured}
		if err := s.Publish(7, now); err != stratErr.ErrPublishWithoutVersion {
			t.Fatalf("expected ErrPublishWithoutVersion, got %v", err)
		}
	})

	t.Run("publish flips visibility to public", func(t *testing.T) {
		s := &Strategy{
			Status:           valueobject.LifecycleStatusConfigured,
			Visibility:       valueobject.VisibilityPrivate,
			CurrentVersionID: 7,
		}
		if err := s.Publish(7, now); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		if s.Status != valueobject.LifecycleStatusPublished {
			t.Fatalf("expected PUBLISHED, got %s", s.Status)
		}
		if s.Visibility != valueobject.VisibilityPublic {
			t.Fatalf("expected PUBLIC, got %s", s.Visibility)
		}
		if s.PublishedAt == nil {
			t.Fatal("expected PublishedAt set")
		}
	})

	t.Run("publish same version twice is rejected", func(t *testing.T) {
		s := &Strategy{
			Status:           valueobject.LifecycleStatusPublished,
			Visibility:       valueobject.VisibilityPublic,
			CurrentVersionID: 7,
		}
		if err := s.Publish(7, now); err != stratErr.ErrAlreadyPublished {
			t.Fatalf("expected ErrAlreadyPublished, got %v", err)
		}
	})

	t.Run("archive is idempotent and clears visibility", func(t *testing.T) {
		s := &Strategy{Status: valueobject.LifecycleStatusPublished, Visibility: valueobject.VisibilityPublic}
		if err := s.Archive(now); err != nil {
			t.Fatalf("Archive: %v", err)
		}
		if s.Status != valueobject.LifecycleStatusArchived {
			t.Fatalf("expected ARCHIVED, got %s", s.Status)
		}
		if s.Visibility != valueobject.VisibilityPrivate {
			t.Fatalf("expected PRIVATE after archive, got %s", s.Visibility)
		}
		// Idempotent re-archive.
		if err := s.Archive(now.Add(time.Hour)); err != nil {
			t.Fatalf("Archive (idempotent): %v", err)
		}
	})

	t.Run("edits on archived strategy are rejected", func(t *testing.T) {
		s := &Strategy{Status: valueobject.LifecycleStatusArchived}
		title := "new"
		if err := s.UpdateMeta(MetaPatch{Title: &title}, now); err != stratErr.ErrStrategyArchived {
			t.Fatalf("expected ErrStrategyArchived, got %v", err)
		}
		if err := s.AttachVersion(99, now); err != stratErr.ErrStrategyArchived {
			t.Fatalf("expected ErrStrategyArchived from AttachVersion, got %v", err)
		}
	})
}

// TestUpdateMetaTagsCanonical guards the NormaliseTags pipeline through
// the public UpdateMeta entry point: callers that pass a non-canonical
// slice get the canonical form back without explicit pre-processing.
func TestUpdateMetaTagsCanonical(t *testing.T) {
	now := time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC)
	s := &Strategy{
		Title:      "Mean reversion",
		Status:     valueobject.LifecycleStatusDraft,
		Visibility: valueobject.VisibilityPrivate,
	}
	tags := []string{" Alpha ", "alpha", "Beta", ""}
	if err := s.UpdateMeta(MetaPatch{Tags: &tags}, now); err != nil {
		t.Fatalf("UpdateMeta: %v", err)
	}
	if got := len(s.Tags); got != 2 {
		t.Fatalf("expected 2 tags, got %d (%v)", got, s.Tags)
	}
	if s.Tags[0] != "alpha" || s.Tags[1] != "beta" {
		t.Fatalf("tags not canonical: %v", s.Tags)
	}
}

// TestValidate_TitleRules enforces the persistence-shape contract.
func TestValidate_TitleRules(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		want    error
		visibility valueobject.Visibility
		status     valueobject.LifecycleStatus
	}{
		{"empty title", "", stratErr.ErrTitleRequired, valueobject.VisibilityPrivate, valueobject.LifecycleStatusDraft},
		{"too long", string(make([]byte, MaxTitleLength+1)), stratErr.ErrTitleTooLong, valueobject.VisibilityPrivate, valueobject.LifecycleStatusDraft},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Replace zero bytes with 'a' so only the length matters.
			b := []byte(tc.title)
			for i := range b {
				if b[i] == 0 {
					b[i] = 'a'
				}
			}
			s := &Strategy{
				Title:      string(b),
				Visibility: tc.visibility,
				Status:     tc.status,
			}
			if err := s.Validate(); err != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

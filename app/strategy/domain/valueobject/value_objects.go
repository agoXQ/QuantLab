// Package valueobject defines shared value objects for the Strategy
// Service domain.
package valueobject

import (
	"fmt"
	"strings"
	"time"
)

// LifecycleStatus is the canonical state of a Strategy aggregate.
//
// The state machine is intentionally narrow for the MVP:
//
//	Draft -> Configured -> Backtested -> Published -> Archived
//
// Backtested is a soft transition that the Backtest Engine flips by
// emitting a finished event the application service consumes; Published
// only requires a previously persisted version, not a backtested one,
// so the platform can surface configurable strategies even before the
// first replay finishes (the UI gates the Publish button independently).
type LifecycleStatus string

const (
	LifecycleStatusDraft       LifecycleStatus = "DRAFT"
	LifecycleStatusConfigured  LifecycleStatus = "CONFIGURED"
	LifecycleStatusBacktested  LifecycleStatus = "BACKTESTED"
	LifecycleStatusPublished   LifecycleStatus = "PUBLISHED"
	LifecycleStatusArchived    LifecycleStatus = "ARCHIVED"
)

// IsValid reports whether s is a recognised status.
func (s LifecycleStatus) IsValid() bool {
	switch s {
	case LifecycleStatusDraft, LifecycleStatusConfigured, LifecycleStatusBacktested,
		LifecycleStatusPublished, LifecycleStatusArchived:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether s blocks further edits. Only Archived is
// terminal in the MVP; the rest of the lifecycle remains mutable so
// users can keep iterating on a published strategy and have the change
// land as a new version.
func (s LifecycleStatus) IsTerminal() bool {
	return s == LifecycleStatusArchived
}

// String implements fmt.Stringer.
func (s LifecycleStatus) String() string { return string(s) }

// Visibility decides whether a strategy is part of the public surface.
//
// The Strategy Service writes the canonical visibility; downstream
// services (Ranking, Community) read it to decide eligibility. Unlisted
// is reserved for content the owner shares via direct link without
// surfacing it on lists; we keep the value here even though MVP UI does
// not expose it yet, so the persisted enum stays stable.
type Visibility string

const (
	VisibilityPrivate  Visibility = "PRIVATE"
	VisibilityPublic   Visibility = "PUBLIC"
	VisibilityUnlisted Visibility = "UNLISTED"
)

// IsValid reports whether v is recognised.
func (v Visibility) IsValid() bool {
	switch v {
	case VisibilityPrivate, VisibilityPublic, VisibilityUnlisted:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (v Visibility) String() string { return string(v) }

// ParseVisibility accepts the API-side casing.
func ParseVisibility(s string) (Visibility, error) {
	if s == "" {
		return VisibilityPrivate, nil
	}
	v := Visibility(strings.ToUpper(s))
	if !v.IsValid() {
		return "", fmt.Errorf("invalid visibility: %s", s)
	}
	return v, nil
}

// ParseLifecycleStatus accepts the API-side casing; empty input maps to
// Draft so the API can construct a fresh strategy without supplying it.
func ParseLifecycleStatus(s string) (LifecycleStatus, error) {
	if s == "" {
		return LifecycleStatusDraft, nil
	}
	v := LifecycleStatus(strings.ToUpper(s))
	if !v.IsValid() {
		return "", fmt.Errorf("invalid lifecycle status: %s", s)
	}
	return v, nil
}

// FormatTime formats t in RFC3339 with second precision; empty for zero
// values so consumers can rely on the field being absent.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

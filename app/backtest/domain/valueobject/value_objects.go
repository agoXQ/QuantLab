// Package valueobject defines shared value objects for the Backtest Engine domain.
package valueobject

import (
	"fmt"
	"strings"
	"time"
)

// JobStatus is the lifecycle state of a BacktestJob.
//
// The state machine is intentionally narrow for the MVP:
//
//	Created -> Queued -> Running -> Completed | Failed | Cancelled
//
// Archived is reserved for the post-MVP retention pipeline and is not
// produced by the synchronous executor.
type JobStatus string

const (
	JobStatusCreated   JobStatus = "CREATED"
	JobStatusQueued    JobStatus = "QUEUED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
)

// IsTerminal reports whether the status represents an end state.
func (s JobStatus) IsTerminal() bool {
	switch s {
	case JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValid reports whether s is one of the recognised statuses.
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusCreated, JobStatusQueued, JobStatusRunning,
		JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (s JobStatus) String() string { return string(s) }

// OrderSide is the direction of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// IsValid reports whether s is a recognised order side.
func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}

// OrderStatus is the lifecycle of a single order.
type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusFilled   OrderStatus = "FILLED"
	OrderStatusRejected OrderStatus = "REJECTED"
	OrderStatusCanceled OrderStatus = "CANCELED"
)

// RebalanceFrequency controls how often the strategy is re-evaluated.
//
// MVP supports the calendar-driven cadences below; ad-hoc / on-signal
// rebalancing arrives with the strategy-service integration.
type RebalanceFrequency string

const (
	RebalanceDaily   RebalanceFrequency = "daily"
	RebalanceWeekly  RebalanceFrequency = "weekly"
	RebalanceMonthly RebalanceFrequency = "monthly"
)

// IsValid reports whether the frequency is supported.
func (r RebalanceFrequency) IsValid() bool {
	switch r {
	case RebalanceDaily, RebalanceWeekly, RebalanceMonthly:
		return true
	default:
		return false
	}
}

// ParseRebalanceFrequency parses a string to a RebalanceFrequency, defaulting
// to daily when empty.
func ParseRebalanceFrequency(s string) (RebalanceFrequency, error) {
	if s == "" {
		return RebalanceDaily, nil
	}
	r := RebalanceFrequency(strings.ToLower(s))
	if !r.IsValid() {
		return "", fmt.Errorf("invalid rebalance frequency: %s", s)
	}
	return r, nil
}

// DateRange is an inclusive [Start, End] trade-date window.
//
// Mirrors the value object used by Market Data so we can pass ranges across
// the seam without re-marshalling. We deliberately do not import the market
// package here to keep Backtest's domain free of cross-context dependencies.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// IsZero reports whether the range is uninitialised.
func (r DateRange) IsZero() bool { return r.Start.IsZero() && r.End.IsZero() }

// Validate ensures the range is well-formed when set.
func (r DateRange) Validate() error {
	if r.IsZero() {
		return fmt.Errorf("date range is required")
	}
	if r.Start.IsZero() || r.End.IsZero() {
		return fmt.Errorf("date range start and end must both be set")
	}
	if r.End.Before(r.Start) {
		return fmt.Errorf("date range end %s is before start %s",
			r.End.Format("2006-01-02"), r.Start.Format("2006-01-02"))
	}
	return nil
}

// FormatDate formats t as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

// ParseDate accepts the two date shapes the rest of the platform uses.
func ParseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	for _, layout := range []string{"2006-01-02", "20060102"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}

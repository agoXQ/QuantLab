// Package valueobject defines shared value objects for the Market Data domain.
package valueobject

import (
	"fmt"
	"strings"
	"time"
)

// Period represents a K-line period.
type Period string

const (
	PeriodDay   Period = "day"
	PeriodWeek  Period = "week"
	PeriodMonth Period = "month"
)

// IsValid reports whether the period is supported.
func (p Period) IsValid() bool {
	switch p {
	case PeriodDay, PeriodWeek, PeriodMonth:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (p Period) String() string { return string(p) }

// ParsePeriod parses a string into a Period, defaulting to day.
func ParsePeriod(s string) (Period, error) {
	if s == "" {
		return PeriodDay, nil
	}
	p := Period(strings.ToLower(s))
	if !p.IsValid() {
		return "", fmt.Errorf("invalid period: %s", s)
	}
	return p, nil
}

// Adjustment represents a price adjustment mode.
type Adjustment string

const (
	AdjustmentNone Adjustment = "none" // raw, unadjusted prices
	AdjustmentPre  Adjustment = "pre"  // forward-adjusted (前复权)
	AdjustmentPost Adjustment = "post" // backward-adjusted (后复权)
)

// IsValid reports whether the adjustment mode is supported.
func (a Adjustment) IsValid() bool {
	switch a {
	case AdjustmentNone, AdjustmentPre, AdjustmentPost:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (a Adjustment) String() string { return string(a) }

// ParseAdjustment parses a string into an Adjustment, defaulting to pre.
func ParseAdjustment(s string) (Adjustment, error) {
	if s == "" {
		return AdjustmentPre, nil
	}
	a := Adjustment(strings.ToLower(s))
	if !a.IsValid() {
		return "", fmt.Errorf("invalid adjustment: %s", s)
	}
	return a, nil
}

// Market identifies a trading market.
type Market string

const (
	MarketCN Market = "CN" // 中国大陆 A 股
	MarketHK Market = "HK"
	MarketUS Market = "US"
)

// AssetType identifies the type of a security.
type AssetType string

const (
	AssetTypeStock AssetType = "STOCK"
	AssetTypeETF   AssetType = "ETF"
	AssetTypeIndex AssetType = "INDEX"
	AssetTypeFund  AssetType = "FUND"
	AssetTypeBond  AssetType = "BOND"
)

// SecurityStatus represents the lifecycle status of a security.
type SecurityStatus string

const (
	StatusListed    SecurityStatus = "LISTED"
	StatusSuspended SecurityStatus = "SUSPENDED"
	StatusDelisted  SecurityStatus = "DELISTED"
	StatusST        SecurityStatus = "ST"
)

// ReportType represents a financial report period type.
type ReportType string

const (
	ReportAnnual    ReportType = "annual"
	ReportInterim   ReportType = "interim"  // 半年报
	ReportQ1        ReportType = "q1"
	ReportQ3        ReportType = "q3"
	ReportTTM       ReportType = "ttm"
)

// IsValid reports whether the report type is supported.
func (r ReportType) IsValid() bool {
	switch r {
	case ReportAnnual, ReportInterim, ReportQ1, ReportQ3, ReportTTM:
		return true
	default:
		return false
	}
}

// CorporateActionType represents a type of corporate action.
type CorporateActionType string

const (
	ActionDividend CorporateActionType = "DIVIDEND"
	ActionBonus    CorporateActionType = "BONUS"     // 送股
	ActionRights   CorporateActionType = "RIGHTS"    // 配股
	ActionTransfer CorporateActionType = "TRANSFER"  // 转增
	ActionSplit    CorporateActionType = "SPLIT"
	ActionMerge    CorporateActionType = "MERGE"
)

// DateRange is an inclusive [Start, End] date range in trade-date semantics.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// IsZero reports whether the range is uninitialized.
func (r DateRange) IsZero() bool { return r.Start.IsZero() && r.End.IsZero() }

// Validate ensures the range is well-formed when set.
func (r DateRange) Validate() error {
	if r.IsZero() {
		return nil
	}
	if !r.Start.IsZero() && !r.End.IsZero() && r.End.Before(r.Start) {
		return fmt.Errorf("date range end %s is before start %s",
			r.End.Format("2006-01-02"), r.Start.Format("2006-01-02"))
	}
	return nil
}

// ParseDate parses a YYYY-MM-DD or YYYYMMDD date string.
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

// FormatDate formats a date as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

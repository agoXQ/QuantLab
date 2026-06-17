// Package calendar defines the TradingCalendar aggregate.
package calendar

import (
	"context"
	"sort"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// TradingDay represents a single calendar entry indicating whether the
// market is open on a given day.
type TradingDay struct {
	TradeDate time.Time `json:"trade_date"`
	IsOpen    bool      `json:"is_open"`
}

// Calendar is an in-memory view of trading days, supporting fast lookups for
// adjustment and ingestion logic.
type Calendar struct {
	days []TradingDay
}

// NewCalendar returns a Calendar sorted by trade_date ascending.
func NewCalendar(days []TradingDay) *Calendar {
	cp := make([]TradingDay, len(days))
	copy(cp, days)
	sort.Slice(cp, func(i, j int) bool { return cp[i].TradeDate.Before(cp[j].TradeDate) })
	return &Calendar{days: cp}
}

// Days returns the underlying sorted slice. The result must not be mutated.
func (c *Calendar) Days() []TradingDay { return c.days }

// IsOpen reports whether the given calendar day is a trading day.
func (c *Calendar) IsOpen(at time.Time) bool {
	idx := sort.Search(len(c.days), func(i int) bool {
		return !c.days[i].TradeDate.Before(at)
	})
	if idx >= len(c.days) {
		return false
	}
	d := c.days[idx]
	return sameDay(d.TradeDate, at) && d.IsOpen
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

// Repository persists trading calendar data.
type Repository interface {
	Range(ctx context.Context, r valueobject.DateRange) ([]TradingDay, error)
	BulkUpsert(ctx context.Context, days []TradingDay) error
}

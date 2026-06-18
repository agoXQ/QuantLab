// Package marketdata provides marketdata.Provider adapters used by the
// Backtest engine. The in-memory provider is the default fixture for tests
// and CI smoke runs; a Market Data adapter (in-process service / gRPC
// client) joins it later without touching the application layer.
package marketdata

import (
	"context"
	"sort"
	"sync"
	"time"

	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
)

// InMemory is an in-process marketdata.Provider keyed by stock code.
type InMemory struct {
	mu       sync.RWMutex
	bars     map[string][]dommarket.Bar
	calendar []time.Time
}

// NewInMemory builds an empty InMemory provider.
func NewInMemory() *InMemory {
	return &InMemory{bars: make(map[string][]dommarket.Bar)}
}

// SetBars replaces the bars associated with a stock code. Bars are copied
// and sorted by trade date so callers can keep mutating their own slice.
func (m *InMemory) SetBars(stockCode string, bars []dommarket.Bar) {
	cp := make([]dommarket.Bar, len(bars))
	copy(cp, bars)
	sort.SliceStable(cp, func(i, j int) bool { return cp[i].TradeDate.Before(cp[j].TradeDate) })
	for i := range cp {
		cp[i].StockCode = stockCode
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bars[stockCode] = cp
}

// SetCalendar overrides the trading calendar. When unset, the provider
// derives the calendar from the union of bar dates so callers can rely on
// either path.
func (m *InMemory) SetCalendar(days []time.Time) {
	cp := make([]time.Time, len(days))
	copy(cp, days)
	sort.SliceStable(cp, func(i, j int) bool { return cp[i].Before(cp[j]) })
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calendar = cp
}

// LoadBars implements marketdata.Provider.
func (m *InMemory) LoadBars(_ context.Context, req dommarket.BarsRequest) (map[string][]dommarket.Bar, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string][]dommarket.Bar, len(req.StockCodes))
	for _, code := range req.StockCodes {
		bars := m.bars[code]
		if len(bars) == 0 {
			out[code] = nil
			continue
		}
		filtered := bars[:0:0]
		for _, b := range bars {
			if !req.Range.Start.IsZero() && b.TradeDate.Before(req.Range.Start) {
				continue
			}
			if !req.Range.End.IsZero() && b.TradeDate.After(req.Range.End) {
				continue
			}
			filtered = append(filtered, b)
		}
		out[code] = filtered
	}
	return out, nil
}

// TradingDays implements marketdata.Provider.
func (m *InMemory) TradingDays(_ context.Context, req dommarket.CalendarRequest) ([]time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.calendar) == 0 {
		// Falling back to nil tells the engine to derive a calendar from
		// the union of loaded bars.
		return nil, nil
	}
	out := make([]time.Time, 0, len(m.calendar))
	for _, d := range m.calendar {
		if !req.Range.Start.IsZero() && d.Before(req.Range.Start) {
			continue
		}
		if !req.Range.End.IsZero() && d.After(req.Range.End) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

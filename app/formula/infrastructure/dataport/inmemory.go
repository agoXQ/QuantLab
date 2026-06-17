// Package dataport provides DataPort adapters for the Formula Engine
// evaluator.
//
// The in-memory adapter is the default wired into the local service binary
// when no Market Data RPC client is configured. It lets the evaluator boot
// and exercise the full pipeline without external dependencies; tests and
// notebooks use it as their primary fixture.
package dataport

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// InMemory holds bars and financials in process memory. Bars are indexed by
// stock code; for each stock they must be sorted by timestamp ascending.
type InMemory struct {
	mu         sync.RWMutex
	bars       map[string][]series.Bar
	financials map[string]map[string]float64
}

// NewInMemory builds an empty in-memory data port.
func NewInMemory() *InMemory {
	return &InMemory{
		bars:       make(map[string][]series.Bar),
		financials: make(map[string]map[string]float64),
	}
}

// SetBars replaces the bars associated with stockCode. Bars are copied so
// later mutations on the caller side do not leak in.
func (m *InMemory) SetBars(stockCode string, bars []series.Bar) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]series.Bar, len(bars))
	copy(cp, bars)
	sort.SliceStable(cp, func(i, j int) bool {
		return cp[i].Timestamp.Before(cp[j].Timestamp)
	})
	m.bars[stockCode] = cp
}

// SetFinancials replaces the financial metric snapshot for stockCode.
func (m *InMemory) SetFinancials(stockCode string, metrics map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make(map[string]float64, len(metrics))
	for k, v := range metrics {
		cp[strings.ToUpper(k)] = v
	}
	m.financials[stockCode] = cp
}

// LoadBars implements domainEval.DataPort.
func (m *InMemory) LoadBars(_ context.Context, req domainEval.BarsRequest) (map[string][]series.Bar, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := req.AsOfDate
	out := make(map[string][]series.Bar, len(req.StockCodes))
	for _, code := range req.StockCodes {
		bars, ok := m.bars[code]
		if !ok {
			out[code] = nil
			continue
		}
		filtered := filterByCutoff(bars, cutoff)
		if req.LookbackBars > 0 && len(filtered) > req.LookbackBars+1 {
			filtered = filtered[len(filtered)-(req.LookbackBars+1):]
		}
		cp := make([]series.Bar, len(filtered))
		copy(cp, filtered)
		out[code] = cp
	}
	return out, nil
}

// LoadFinancialsLatest implements domainEval.DataPort.
//
// Cutoff handling is not modelled here: the in-memory adapter assumes the
// caller uploads a single point-in-time snapshot. The Market Data adapter is
// responsible for the actual as-of resolution.
func (m *InMemory) LoadFinancialsLatest(_ context.Context, req domainEval.FinancialsRequest) (map[string]map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]map[string]float64, len(req.StockCodes))
	for _, code := range req.StockCodes {
		fin, ok := m.financials[code]
		if !ok {
			continue
		}
		copyMetrics := make(map[string]float64, len(req.Metrics))
		if len(req.Metrics) == 0 {
			for k, v := range fin {
				copyMetrics[k] = v
			}
		} else {
			for _, name := range req.Metrics {
				key := strings.ToUpper(name)
				if v, ok := fin[key]; ok {
					copyMetrics[key] = v
				}
			}
		}
		out[code] = copyMetrics
	}
	return out, nil
}

func filterByCutoff(bars []series.Bar, cutoff time.Time) []series.Bar {
	if cutoff.IsZero() {
		return bars
	}
	// bars is timestamp-ascending; binary search for the first bar > cutoff.
	idx := sort.Search(len(bars), func(i int) bool {
		return bars[i].Timestamp.After(cutoff)
	})
	return bars[:idx]
}

// Compile-time assertion that InMemory satisfies the DataPort contract.
var _ domainEval.DataPort = (*InMemory)(nil)

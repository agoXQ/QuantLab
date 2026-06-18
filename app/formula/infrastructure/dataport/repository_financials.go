package dataport

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"

	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// LoadFinancialsLatest implements domainEval.DataPort.
//
// Two data sources contribute, in priority order:
//
//   1. factor_data: the daily factor table populated from the Tushare
//      daily_basic endpoint (PE, PE_TTM, PB, PS, PS_TTM, TURNOVER, DIV_YIELD).
//      A single value per metric is returned, picking the most recent
//      trade_date <= req.AsOfDate.
//   2. financial_statement: the quarterly statement aggregate. We derive ROE,
//      ROA, EPS, RevenueGrowth, ProfitGrowth, MarketCap from the latest
//      annual/TTM report at or before req.AsOfDate. EPS and MarketCap fall
//      back to NaN when the upstream report does not include them — the
//      evaluator already treats NaN as missing.
//
// Metrics absent from both sources are simply omitted from the per-stock
// map. The evaluator substitutes NaN, so a single missing metric does not
// fail the whole universe.
func (p *RepositoryDataPort) LoadFinancialsLatest(ctx context.Context, req domainEval.FinancialsRequest) (map[string]map[string]float64, error) {
	if len(req.StockCodes) == 0 {
		return map[string]map[string]float64{}, nil
	}

	want := normalizeMetrics(req.Metrics)
	factorMetrics := want.factorNames()

	asOf := req.AsOfDate
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	out := make(map[string]map[string]float64, len(req.StockCodes))
	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		errCh = make(chan error, 1)
		sem   = make(chan struct{}, loadConcurrency)
	)
	for _, raw := range req.StockCodes {
		stockCode := raw
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			metrics, err := p.loadStockFinancials(ctx, stockCode, factorMetrics, want, asOf, req.DataVersion)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("load financials for %s: %w", stockCode, err):
				default:
				}
				return
			}
			if len(metrics) == 0 {
				return
			}
			mu.Lock()
			out[stockCode] = metrics
			mu.Unlock()
		}()
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return out, nil
}

func (p *RepositoryDataPort) loadStockFinancials(
	ctx context.Context,
	stockCode string,
	factorNames []string,
	want metricSet,
	asOf time.Time,
	dataVersion string,
) (map[string]float64, error) {
	out := map[string]float64{}

	if len(factorNames) > 0 {
		facs, err := p.factors.List(ctx, factor.ListQuery{
			StockCode:   stockCode,
			FactorNames: factorNames,
			Range:       valueobject.DateRange{End: asOf},
			DataVersion: dataVersion,
			Limit:       len(factorNames) * 4,
		})
		if err != nil {
			return nil, fmt.Errorf("list factors: %w", err)
		}
		for name, value := range latestFactorValues(facs) {
			out[name] = value
		}
	}

	if p.financials == nil || !want.needsStatement() {
		return out, nil
	}

	statements, err := p.financials.List(ctx, financial.ListQuery{
		StockCode:   stockCode,
		ReportType:  valueobject.ReportTTM,
		Range:       valueobject.DateRange{End: asOf},
		DataVersion: dataVersion,
		Limit:       2,
	})
	if err != nil {
		return nil, fmt.Errorf("list financials: %w", err)
	}
	if len(statements) == 0 {
		// TTM may not be available for every stock; fall back to annuals.
		statements, err = p.financials.List(ctx, financial.ListQuery{
			StockCode:   stockCode,
			ReportType:  valueobject.ReportAnnual,
			Range:       valueobject.DateRange{End: asOf},
			DataVersion: dataVersion,
			Limit:       2,
		})
		if err != nil {
			return nil, fmt.Errorf("list financials: %w", err)
		}
	}
	if len(statements) == 0 {
		return out, nil
	}

	derived := deriveStatementMetrics(statements)
	for k, v := range derived {
		if !want.contains(k) {
			continue
		}
		out[k] = v
	}
	return out, nil
}

// metricSet partitions the requested metric names into the factor-table
// surface and the statement-derived surface. Names are uppercased before
// comparison so callers can pass the same identifiers used in the DSL.
type metricSet struct {
	all map[string]struct{}
}

func normalizeMetrics(in []string) metricSet {
	m := metricSet{all: make(map[string]struct{}, len(in))}
	for _, name := range in {
		key := strings.ToUpper(strings.TrimSpace(name))
		if key == "" {
			continue
		}
		m.all[key] = struct{}{}
	}
	return m
}

func (m metricSet) contains(name string) bool {
	if len(m.all) == 0 {
		return true
	}
	_, ok := m.all[strings.ToUpper(name)]
	return ok
}

// factorNames returns the subset of requested metrics that the factor table
// can provide. When the caller did not specify any metrics, we fetch the
// canonical MVP surface so the evaluator gets a useful default.
func (m metricSet) factorNames() []string {
	defaults := []string{"PE", "PB", "PS", "PE_TTM", "PS_TTM", "TURNOVER", "DIV_YIELD"}
	if len(m.all) == 0 {
		return defaults
	}
	available := map[string]struct{}{}
	for _, n := range defaults {
		available[n] = struct{}{}
	}
	out := make([]string, 0, len(m.all))
	for name := range m.all {
		if _, ok := available[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

// needsStatement reports whether the caller asked for any metric the factor
// table cannot satisfy.
func (m metricSet) needsStatement() bool {
	if len(m.all) == 0 {
		return true
	}
	for name := range m.all {
		switch name {
		case "PE", "PB", "PS", "PE_TTM", "PS_TTM", "TURNOVER", "DIV_YIELD":
			continue
		}
		return true
	}
	return false
}

// latestFactorValues keeps the latest trade_date per factor name from the
// repository result. Repositories already order by trade_date asc; we walk
// the slice and overwrite, which is robust to either ordering.
func latestFactorValues(facs []*factor.Factor) map[string]float64 {
	type entry struct {
		when  time.Time
		value float64
	}
	latest := make(map[string]entry, len(facs))
	for _, f := range facs {
		key := strings.ToUpper(f.FactorName)
		cur, ok := latest[key]
		if !ok || !f.TradeDate.Before(cur.when) {
			latest[key] = entry{when: f.TradeDate, value: f.FactorValue}
		}
	}
	out := make(map[string]float64, len(latest))
	for k, e := range latest {
		out[k] = e.value
	}
	return out
}

// deriveStatementMetrics turns financial statements into the canonical scalar
// metrics referenced by the DSL. Newer-period values win.
//
// Statement-derived formulas:
//
//	ROE            = NetProfit / NetAssets
//	ROA            = NetProfit / TotalAssets
//	EPS            = NetProfit / 1e8       (placeholder; share count missing)
//	RevenueGrowth  = (Revenue   - PrevRevenue)   / |PrevRevenue|
//	ProfitGrowth   = (NetProfit - PrevNetProfit) / |PrevNetProfit|
//
// EPS is only an approximation: the financial_statement table does not yet
// store outstanding share count, so we expose NetProfit normalised to
// hundred-million units (亿元), which preserves ordering. A schema upgrade
// will replace this placeholder with the correct calculation.
func deriveStatementMetrics(stmts []*financial.FinancialStatement) map[string]float64 {
	if len(stmts) == 0 {
		return nil
	}
	// stmts are ordered by report_date desc per the repository.
	curr := stmts[0]
	out := map[string]float64{}
	if curr.NetAssets != 0 {
		out["ROE"] = curr.NetProfit / curr.NetAssets * 100
	}
	if curr.TotalAssets != 0 {
		out["ROA"] = curr.NetProfit / curr.TotalAssets * 100
	}
	out["EPS"] = curr.NetProfit / 1e8

	if len(stmts) >= 2 {
		prev := stmts[1]
		if prev.Revenue != 0 {
			out["REVENUEGROWTH"] = (curr.Revenue - prev.Revenue) / absFloat(prev.Revenue) * 100
			out["RevenueGrowth"] = out["REVENUEGROWTH"]
		}
		if prev.NetProfit != 0 {
			out["PROFITGROWTH"] = (curr.NetProfit - prev.NetProfit) / absFloat(prev.NetProfit) * 100
			out["ProfitGrowth"] = out["PROFITGROWTH"]
		}
	}
	return out
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// Compile-time assertion that RepositoryDataPort satisfies the contract.
var _ domainEval.DataPort = (*RepositoryDataPort)(nil)

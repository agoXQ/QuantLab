package tests

import (
	"context"
	"strings"
	"sort"
	"sync"
	"testing"
	"time"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
	infraDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"

	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraMarketAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
)

// --- in-memory repos ---

type stubBarRepo struct {
	mu   sync.Mutex
	data map[string][]*marketbar.MarketBar
}

func newStubBarRepo() *stubBarRepo {
	return &stubBarRepo{data: map[string][]*marketbar.MarketBar{}}
}

func (r *stubBarRepo) seed(stockCode string, bars []*marketbar.MarketBar) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[strings.ToUpper(stockCode)] = bars
}

func (r *stubBarRepo) Range(_ context.Context, q marketbar.RangeQuery) ([]*marketbar.MarketBar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bars := r.data[strings.ToUpper(q.StockCode)]
	out := make([]*marketbar.MarketBar, 0, len(bars))
	for _, b := range bars {
		if !q.Range.Start.IsZero() && b.TradeDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && b.TradeDate.After(q.Range.End) {
			continue
		}
		cp := *b
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TradeDate.Before(out[j].TradeDate) })
	return out, nil
}

func (r *stubBarRepo) Latest(_ context.Context, _ string, _ valueobject.Period, _ time.Time) (*marketbar.MarketBar, error) {
	return nil, nil
}

func (r *stubBarRepo) BulkUpsert(_ context.Context, _ []*marketbar.MarketBar) error { return nil }

type stubFinancialRepo struct {
	data map[string][]*financial.FinancialStatement
}

func newStubFinancialRepo() *stubFinancialRepo {
	return &stubFinancialRepo{data: map[string][]*financial.FinancialStatement{}}
}

func (r *stubFinancialRepo) seed(stockCode string, list []*financial.FinancialStatement) {
	r.data[strings.ToUpper(stockCode)] = list
}

func (r *stubFinancialRepo) List(_ context.Context, q financial.ListQuery) ([]*financial.FinancialStatement, error) {
	stmts := r.data[strings.ToUpper(q.StockCode)]
	out := make([]*financial.FinancialStatement, 0, len(stmts))
	for _, f := range stmts {
		if q.ReportType != "" && f.ReportType != q.ReportType {
			continue
		}
		if !q.Range.End.IsZero() && f.ReportDate.After(q.Range.End) {
			continue
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ReportDate.After(out[j].ReportDate) })
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

func (r *stubFinancialRepo) BulkUpsert(_ context.Context, _ []*financial.FinancialStatement) error {
	return nil
}

type stubFactorRepo struct {
	data map[string][]*factor.Factor
}

func newStubFactorRepo() *stubFactorRepo {
	return &stubFactorRepo{data: map[string][]*factor.Factor{}}
}

func (r *stubFactorRepo) seed(stockCode string, facs []*factor.Factor) {
	r.data[strings.ToUpper(stockCode)] = facs
}

func (r *stubFactorRepo) List(_ context.Context, q factor.ListQuery) ([]*factor.Factor, error) {
	facs := r.data[strings.ToUpper(q.StockCode)]
	wanted := map[string]struct{}{}
	for _, n := range q.FactorNames {
		wanted[strings.ToUpper(n)] = struct{}{}
	}
	out := make([]*factor.Factor, 0, len(facs))
	for _, f := range facs {
		if len(wanted) > 0 {
			if _, ok := wanted[strings.ToUpper(f.FactorName)]; !ok {
				continue
			}
		}
		if !q.Range.End.IsZero() && f.TradeDate.After(q.Range.End) {
			continue
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TradeDate.Before(out[j].TradeDate) })
	return out, nil
}

func (r *stubFactorRepo) BulkUpsert(_ context.Context, _ []*factor.Factor) error { return nil }

// --- helpers ---

func newRepoPort(t *testing.T) (*infraDataport.RepositoryDataPort, *stubBarRepo, *stubFinancialRepo, *stubFactorRepo) {
	t.Helper()
	bars := newStubBarRepo()
	fins := newStubFinancialRepo()
	facs := newStubFactorRepo()
	port, err := infraDataport.NewRepository(infraDataport.RepositoryConfig{
		Bars:       bars,
		Financials: fins,
		Factors:    facs,
		Adjuster:   infraMarketAdj.NewFactorAdjuster(),
		Adjustment: valueobject.AdjustmentNone,
	})
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	return port, bars, fins, facs
}

func mkRepoBars(stockCode string, start time.Time, closes []float64) []*marketbar.MarketBar {
	out := make([]*marketbar.MarketBar, len(closes))
	for i, c := range closes {
		out[i] = &marketbar.MarketBar{
			StockCode: stockCode,
			TradeDate: start.AddDate(0, 0, i),
			Period:    valueobject.PeriodDay,
			Open:      c,
			High:      c,
			Low:       c,
			Close:     c,
			Volume:    100,
			Amount:    c * 100,
			AdjFactor: 1,
		}
	}
	return out
}

// --- tests ---

func TestRepositoryDataPort_LoadBarsReturnsAdaptedSeries(t *testing.T) {
	port, bars, _, _ := newRepoPort(t)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	bars.seed("000001", mkRepoBars("000001", start, []float64{10, 11, 12, 13, 14}))

	out, err := port.LoadBars(context.Background(), domainEval.BarsRequest{
		StockCodes:   []string{"000001"},
		AsOfDate:     start.AddDate(0, 0, 4),
		LookbackBars: 10,
	})
	if err != nil {
		t.Fatalf("LoadBars: %v", err)
	}
	got := out["000001"]
	if len(got) != 5 {
		t.Fatalf("expected 5 bars, got %d", len(got))
	}
	if got[0].Close != 10 || got[4].Close != 14 {
		t.Errorf("unexpected bar values: %+v", got)
	}
	if got[0].Volume != 100 {
		t.Errorf("expected volume 100, got %v", got[0].Volume)
	}
}

func TestRepositoryDataPort_MissingStockReturnsNilSlice(t *testing.T) {
	port, _, _, _ := newRepoPort(t)
	out, err := port.LoadBars(context.Background(), domainEval.BarsRequest{
		StockCodes:   []string{"GHOST"},
		AsOfDate:     time.Now(),
		LookbackBars: 10,
	})
	if err != nil {
		t.Fatalf("LoadBars: %v", err)
	}
	if bars := out["GHOST"]; len(bars) != 0 {
		t.Errorf("expected empty slice, got %v", bars)
	}
}

func TestRepositoryDataPort_LoadFinancialsLatestFromFactors(t *testing.T) {
	port, _, _, facs := newRepoPort(t)
	asOf := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	facs.seed("A", []*factor.Factor{
		{StockCode: "A", TradeDate: asOf.AddDate(0, 0, -10), FactorName: "PE", FactorValue: 12},
		{StockCode: "A", TradeDate: asOf.AddDate(0, 0, -5), FactorName: "PE", FactorValue: 14},
		{StockCode: "A", TradeDate: asOf.AddDate(0, 0, -5), FactorName: "PB", FactorValue: 1.5},
	})

	out, err := port.LoadFinancialsLatest(context.Background(), domainEval.FinancialsRequest{
		StockCodes: []string{"A"},
		Metrics:    []string{"PE", "PB"},
		AsOfDate:   asOf,
	})
	if err != nil {
		t.Fatalf("LoadFinancialsLatest: %v", err)
	}
	if got := out["A"]["PE"]; got != 14 {
		t.Errorf("PE = %v, want 14", got)
	}
	if got := out["A"]["PB"]; got != 1.5 {
		t.Errorf("PB = %v, want 1.5", got)
	}
}

func TestRepositoryDataPort_DerivesROEAndROAFromStatements(t *testing.T) {
	port, _, fins, _ := newRepoPort(t)
	asOf := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	fins.seed("A", []*financial.FinancialStatement{
		{
			StockCode:   "A",
			ReportDate:  time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
			ReportType:  valueobject.ReportTTM,
			Revenue:     220,
			NetProfit:   20,
			NetAssets:   100,
			TotalAssets: 200,
			BasicEPS:    1.25,
			DilutedEPS:  1.20,
		},
		{
			StockCode:   "A",
			ReportDate:  time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			ReportType:  valueobject.ReportTTM,
			Revenue:     200,
			NetProfit:   15,
			NetAssets:   90,
			TotalAssets: 180,
			BasicEPS:    0.95,
			DilutedEPS:  0.90,
		},
	})

	out, err := port.LoadFinancialsLatest(context.Background(), domainEval.FinancialsRequest{
		StockCodes: []string{"A"},
		Metrics:    []string{"ROE", "ROA", "EPS", "REVENUEGROWTH", "PROFITGROWTH"},
		AsOfDate:   asOf,
	})
	if err != nil {
		t.Fatalf("LoadFinancialsLatest: %v", err)
	}
	metrics := out["A"]
	if metrics["ROE"] != 20 {
		t.Errorf("ROE = %v, want 20", metrics["ROE"])
	}
	if metrics["ROA"] != 10 {
		t.Errorf("ROA = %v, want 10", metrics["ROA"])
	}
	if metrics["EPS"] != 1.25 {
		t.Errorf("EPS = %v, want 1.25 (basic_eps from latest statement)", metrics["EPS"])
	}
	if v := metrics["REVENUEGROWTH"]; v <= 0 {
		t.Errorf("REVENUEGROWTH = %v, expected positive", v)
	}
	if v := metrics["PROFITGROWTH"]; v <= 0 {
		t.Errorf("PROFITGROWTH = %v, expected positive", v)
	}
}

// TestRepositoryDataPort_EPSFallsBackToDilutedAndOmitsWhenAbsent locks in
// the statement-priority story: basic EPS wins, diluted is the fallback,
// and if both are zero we drop the EPS metric so the evaluator can
// substitute NaN instead of treating zero as a real reading.
func TestRepositoryDataPort_EPSFallsBackToDilutedAndOmitsWhenAbsent(t *testing.T) {
	port, _, fins, _ := newRepoPort(t)
	asOf := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

	fins.seed("DILUTED", []*financial.FinancialStatement{{
		StockCode:  "DILUTED",
		ReportDate: time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		ReportType: valueobject.ReportTTM,
		NetProfit:  9,
		DilutedEPS: 0.42,
	}})
	fins.seed("MISSING", []*financial.FinancialStatement{{
		StockCode:  "MISSING",
		ReportDate: time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		ReportType: valueobject.ReportTTM,
		NetProfit:  9,
	}})

	out, err := port.LoadFinancialsLatest(context.Background(), domainEval.FinancialsRequest{
		StockCodes: []string{"DILUTED", "MISSING"},
		Metrics:    []string{"EPS"},
		AsOfDate:   asOf,
	})
	if err != nil {
		t.Fatalf("LoadFinancialsLatest: %v", err)
	}
	if got := out["DILUTED"]["EPS"]; got != 0.42 {
		t.Errorf("DILUTED EPS = %v, want 0.42", got)
	}
	if _, ok := out["MISSING"]["EPS"]; ok {
		t.Errorf("MISSING should omit EPS, got %v", out["MISSING"])
	}
}

func TestRepositoryDataPort_ParityWithInMemoryForFinancialFilter(t *testing.T) {
	repo, _, fins, facs := newRepoPort(t)
	asOf := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

	// Repository-backed seed: PE comes from the daily factor table, ROE is
	// derived from the financial statement aggregate (NetProfit/NetAssets).
	facs.seed("X", []*factor.Factor{{StockCode: "X", TradeDate: asOf, FactorName: "PE", FactorValue: 18}})
	facs.seed("Y", []*factor.Factor{{StockCode: "Y", TradeDate: asOf, FactorName: "PE", FactorValue: 30}})
	fins.seed("X", []*financial.FinancialStatement{{
		StockCode: "X", ReportDate: asOf, ReportType: valueobject.ReportTTM,
		NetProfit: 25, NetAssets: 100,
	}})
	fins.seed("Y", []*financial.FinancialStatement{{
		StockCode: "Y", ReportDate: asOf, ReportType: valueobject.ReportTTM,
		NetProfit: 5, NetAssets: 100,
	}})

	// Mirror seed in the in-memory port. ROE is stored in percent units to
	// match the repository's NetProfit/NetAssets * 100 derivation.
	memPort := infraDataport.NewInMemory()
	memPort.SetFinancials("X", map[string]float64{"PE": 18, "ROE": 25})
	memPort.SetFinancials("Y", map[string]float64{"PE": 30, "ROE": 5})

	svc, _ := newEvaluatorService(t)
	universe := []string{"X", "Y"}

	resRepo, err := svc.Evaluate(context.Background(), appFormulaEvaluateRequest("ROE > 15 AND PE < 20", universe, asOf, repo))
	if err != nil {
		t.Fatalf("repo evaluate: %v", err)
	}
	resMem, err := svc.Evaluate(context.Background(), appFormulaEvaluateRequest("ROE > 15 AND PE < 20", universe, asOf, memPort))
	if err != nil {
		t.Fatalf("mem evaluate: %v", err)
	}

	gotRepo := strings.Join(resRepo.Result.Selection.StockCodes, ",")
	gotMem := strings.Join(resMem.Result.Selection.StockCodes, ",")
	if gotRepo != gotMem {
		t.Fatalf("parity mismatch: repo=%q mem=%q", gotRepo, gotMem)
	}
	if gotRepo != "X" {
		t.Errorf("expected only X, got %q", gotRepo)
	}
}

// appFormulaEvaluateRequest is a small builder used by parity tests so the
// long-form EvaluateRequest literal stays out of the call sites.
func appFormulaEvaluateRequest(formula string, universe []string, asOf time.Time, port domainEval.DataPort) appFormula.EvaluateRequest {
	return appFormula.EvaluateRequest{
		Formula:      formula,
		Universe:     universe,
		AsOfDate:     asOf,
		LookbackBars: 30,
		DataPort:     port,
	}
}

// avoid unused import warnings when test file is loaded standalone.
var _ = series.Bar{}

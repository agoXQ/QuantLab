package formula

import (
	"context"
	"fmt"
	"testing"

	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

func TestScreen_DefaultUniverseFilterUsesAllExchangesListedStocks(t *testing.T) {
	source := &fakeScreenSecuritySource{
		version: "v1",
		items: []*security.Security{{
			StockCode: "600519",
			StockName: "贵州茅台",
			Exchange:  "SSE",
			AssetType: valueobject.AssetTypeStock,
			Status:    valueobject.StatusListed,
		}},
	}
	evaluator := &fakeScreenEvaluator{}
	svc := NewScreenService(evaluator, source, fakeScreenDataPort{})

	res, err := svc.Screen(context.Background(), ScreenRequest{Formula: "ROE > 15", Limit: 100})
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}
	if source.lastQuery.Exchange != "" {
		t.Fatalf("expected empty exchange to mean all exchanges, got %q", source.lastQuery.Exchange)
	}
	if source.lastQuery.AssetType != valueobject.AssetTypeStock {
		t.Fatalf("expected default asset type STOCK, got %q", source.lastQuery.AssetType)
	}
	if source.lastQuery.Status != valueobject.StatusListed {
		t.Fatalf("expected default status LISTED, got %q", source.lastQuery.Status)
	}
	if got := evaluator.lastRequest.Universe; len(got) != 1 || got[0] != "600519" {
		t.Fatalf("unexpected evaluator universe: %#v", got)
	}
	if res.DataVersion != "v1" {
		t.Fatalf("expected latest data version v1, got %q", res.DataVersion)
	}
	if len(res.Items) != 1 || res.Items[0].StockName != "贵州茅台" {
		t.Fatalf("unexpected screen items: %#v", res.Items)
	}
}

func TestScreen_LoadsAllUniversePagesBeforeApplyingResultLimit(t *testing.T) {
	items := make([]*security.Security, 0, screenUniversePageSize+1)
	for i := 0; i < screenUniversePageSize+1; i++ {
		items = append(items, &security.Security{
			ID:        int64(i + 1),
			StockCode: fmt.Sprintf("%06d", i+1),
			Exchange:  "SZSE",
			AssetType: valueobject.AssetTypeStock,
			Status:    valueobject.StatusListed,
		})
	}
	source := &fakeScreenSecuritySource{version: "v1", items: items}
	evaluator := &fakeScreenEvaluator{}
	svc := NewScreenService(evaluator, source, fakeScreenDataPort{})

	res, err := svc.Screen(context.Background(), ScreenRequest{Formula: "C > 0", Limit: 10})
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}
	if source.calls != 2 {
		t.Fatalf("expected two security pages to be loaded, got %d", source.calls)
	}
	if got := len(evaluator.lastRequest.Universe); got != screenUniversePageSize+1 {
		t.Fatalf("expected evaluator to see full universe, got %d", got)
	}
	if res.UniverseSize != screenUniversePageSize+1 {
		t.Fatalf("expected full universe size, got %d", res.UniverseSize)
	}
	if got := len(res.Items); got != 10 {
		t.Fatalf("expected result limit to cap returned items, got %d", got)
	}
}

type fakeScreenSecuritySource struct {
	version   string
	items     []*security.Security
	lastQuery security.ListQuery
	calls     int
}

func (s *fakeScreenSecuritySource) GetSecurity(context.Context, string) (*security.Security, error) {
	return nil, nil
}

func (s *fakeScreenSecuritySource) ListSecurities(_ context.Context, q security.ListQuery) ([]*security.Security, string, error) {
	s.lastQuery = q
	s.calls++
	start := 0
	if q.Cursor != "" {
		_, _ = fmt.Sscanf(q.Cursor, "%d", &start)
	}
	limit := q.Limit
	if limit <= 0 {
		limit = len(s.items)
	}
	end := start + limit
	if end > len(s.items) {
		end = len(s.items)
	}
	next := ""
	if end < len(s.items) {
		next = fmt.Sprintf("%d", end)
	}
	return s.items[start:end], next, nil
}

func (s *fakeScreenSecuritySource) LatestDataVersion(context.Context) (string, error) {
	return s.version, nil
}

type fakeScreenEvaluator struct {
	lastRequest EvaluateRequest
}

func (e *fakeScreenEvaluator) Evaluate(_ context.Context, req EvaluateRequest) (*EvaluateResult, error) {
	e.lastRequest = req
	return &EvaluateResult{
		FormulaHash: "hash",
		Result: &domainEval.Result{
			PlanType:  domainCompiler.PlanTypeFilter,
			Selection: &domainEval.Selection{StockCodes: append([]string(nil), req.Universe...)},
		},
	}, nil
}

type fakeScreenDataPort struct{}

func (fakeScreenDataPort) LoadBars(context.Context, domainEval.BarsRequest) (map[string][]series.Bar, error) {
	return nil, nil
}

func (fakeScreenDataPort) LoadFinancialsLatest(context.Context, domainEval.FinancialsRequest) (map[string]map[string]float64, error) {
	return nil, nil
}

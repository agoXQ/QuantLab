package tests

import (
	"context"
	"testing"
	"time"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
	infraDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"
	infraEvaluator "github.com/agoXQ/QuantLab/app/formula/infrastructure/evaluator"
	infraInd "github.com/agoXQ/QuantLab/app/formula/infrastructure/indicators"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newEvaluatorService(t *testing.T) (appFormula.EvaluatorService, *infraDataport.InMemory) {
	t.Helper()
	svc := newService()
	port := infraDataport.NewInMemory()
	eval := infraEvaluator.New(infraInd.NewLibrary(), infraVar.NewRegistry())
	return appFormula.NewEvaluatorService(svc, eval), port
}

func mkBars(start time.Time, closes []float64) []series.Bar {
	bars := make([]series.Bar, len(closes))
	for i, c := range closes {
		bars[i] = series.Bar{
			Timestamp: start.AddDate(0, 0, i),
			Open:      c,
			High:      c,
			Low:       c,
			Close:     c,
			Volume:    1,
			Amount:    c,
		}
	}
	return bars
}

func TestEvaluator_FilterByFinancials(t *testing.T) {
	svc, port := newEvaluatorService(t)

	port.SetFinancials("000001", map[string]float64{"ROE": 20, "PE": 12})
	port.SetFinancials("000002", map[string]float64{"ROE": 5, "PE": 35})
	port.SetFinancials("000003", map[string]float64{"ROE": 16, "PE": 18})

	res, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:      "ROE > 15 AND PE < 20",
		Universe:     []string{"000001", "000002", "000003"},
		AsOfDate:     time.Now(),
		LookbackBars: 10,
		DataPort:     port,
	})
	if err != nil {
		t.Fatalf("Evaluate err: %v", err)
	}
	if res.Result.Selection == nil {
		t.Fatal("expected Selection result")
	}
	got := res.Result.Selection.StockCodes
	if len(got) != 2 || got[0] != "000001" || got[1] != "000003" {
		t.Errorf("expected [000001 000003], got %v", got)
	}
}

func TestEvaluator_FilterByMA(t *testing.T) {
	svc, port := newEvaluatorService(t)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// Strong uptrend -> MA5 should be above MA20 by the end.
	up := make([]float64, 30)
	for i := range up {
		up[i] = float64(i + 1)
	}
	// Flat series -> MA5 equals MA20.
	flat := make([]float64, 30)
	for i := range flat {
		flat[i] = 100
	}
	port.SetBars("UP", mkBars(start, up))
	port.SetBars("FLAT", mkBars(start, flat))

	res, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:      "MA(CLOSE, 5) > MA(CLOSE, 20)",
		Universe:     []string{"UP", "FLAT"},
		AsOfDate:     start.AddDate(0, 0, 29),
		LookbackBars: 30,
		DataPort:     port,
	})
	if err != nil {
		t.Fatalf("Evaluate err: %v", err)
	}
	got := res.Result.Selection.StockCodes
	if len(got) != 1 || got[0] != "UP" {
		t.Errorf("expected [UP], got %v", got)
	}
}

func TestEvaluator_RankingPlan(t *testing.T) {
	svc, port := newEvaluatorService(t)

	port.SetFinancials("A", map[string]float64{"ROE": 10})
	port.SetFinancials("B", map[string]float64{"ROE": 30})
	port.SetFinancials("C", map[string]float64{"ROE": 20})

	res, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:  "ROE * 2",
		Universe: []string{"A", "B", "C"},
		AsOfDate: time.Now(),
		DataPort: port,
	})
	if err != nil {
		t.Fatalf("Evaluate err: %v", err)
	}
	if res.Result.Ranking == nil {
		t.Fatal("expected Ranking result")
	}
	got := res.Result.Ranking.StockCodes
	if len(got) != 3 || got[0] != "B" || got[1] != "C" || got[2] != "A" {
		t.Errorf("expected ranking [B C A], got %v", got)
	}
}

func TestEvaluator_MissingFinancialFallsToNaN(t *testing.T) {
	svc, port := newEvaluatorService(t)
	port.SetFinancials("A", map[string]float64{"ROE": 25})
	// Stock B has no financials at all.

	res, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:  "ROE > 15",
		Universe: []string{"A", "B"},
		AsOfDate: time.Now(),
		DataPort: port,
	})
	if err != nil {
		t.Fatalf("Evaluate err: %v", err)
	}
	got := res.Result.Selection.StockCodes
	if len(got) != 1 || got[0] != "A" {
		t.Errorf("expected only A, got %v", got)
	}
}

func TestEvaluator_RejectInvalidFormula(t *testing.T) {
	svc, port := newEvaluatorService(t)
	_, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:  "UNKNOWN_VAR > 15",
		Universe: []string{"A"},
		DataPort: port,
	})
	if err == nil {
		t.Fatal("expected error for invalid formula")
	}
}

func TestEvaluator_RequiresDataPort(t *testing.T) {
	svc, _ := newEvaluatorService(t)
	_, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:  "ROE > 15",
		Universe: []string{"A"},
		AsOfDate: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error when DataPort is nil")
	}
}

func TestEvaluator_CrossSignal(t *testing.T) {
	svc, port := newEvaluatorService(t)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// Flat 50 for 25 bars, then a single jump to 60 on bar index 25.
	// MA5 jumps to 52, MA20 inches to 50.5 -> CROSS fires on that bar.
	closes := []float64{50, 50, 50, 50, 50, 50, 50, 50, 50, 50,
		50, 50, 50, 50, 50, 50, 50, 50, 50, 50,
		50, 50, 50, 50, 50, 60}
	port.SetBars("X", mkBars(start, closes))

	res, err := svc.Evaluate(context.Background(), appFormula.EvaluateRequest{
		Formula:      "CROSS(MA(CLOSE, 5), MA(CLOSE, 20))",
		Universe:     []string{"X"},
		AsOfDate:     start.AddDate(0, 0, 25),
		LookbackBars: 30,
		DataPort:     port,
	})
	if err != nil {
		t.Fatalf("Evaluate err: %v", err)
	}
	if res.Result.Selection == nil {
		t.Fatal("expected Selection result for SIGNAL plan")
	}
	if len(res.Result.Selection.StockCodes) != 1 {
		t.Errorf("expected X selected on cross, got %v", res.Result.Selection.StockCodes)
	}
}

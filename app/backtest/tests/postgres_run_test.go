package tests

import (
	"context"
	"testing"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	infraEvent "github.com/agoXQ/QuantLab/app/backtest/infrastructure/event"
	infraMatching "github.com/agoXQ/QuantLab/app/backtest/infrastructure/matching"
	infraPg "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/postgres"
)

// TestPostgres_EndToEndRun exercises the full backtest engine against the
// real Postgres repositories. The Formula evaluator and market provider
// reuse the in-memory fixtures so the test stays focused on persistence:
// a successful run must read back its job, report, trades, and snapshots
// without losing fidelity through the JSON / array round-trips.
func TestPostgres_EndToEndRun(t *testing.T) {
	fx := openPg(t)
	mem := newFixture(t)

	clock := func() time.Time { return mem.clock }
	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       infraPg.NewJobRepository(fx.db),
		Orders:     infraPg.NewOrderRepository(fx.db),
		Trades:     infraPg.NewTradeRepository(fx.db),
		Portfolios: infraPg.NewPortfolioRepository(fx.db),
		Reports:    infraPg.NewReportRepository(fx.db),
		Executor:   mem.executor,
		Matching:   infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{}),
		MarketData: mem.provider,
		Publisher:  infraEvent.Noop{},
		Clock:      clock,
	})

	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 12)
	bars := linearBars(calendar, 25, 0.4)
	mem.provider.SetBars("000001", bars)
	mem.provider.SetCalendar(calendar)
	mem.formula.SetBars("000001", formulaBars(bars))
	mem.formula.SetFinancials("000001", map[string]float64{"ROE": 25})

	created, err := svc.Create(context.Background(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100_000,
		Range:          valueobject.DateRange{Start: calendar[0], End: calendar[len(calendar)-1]},
		Config: backtestjob.Config{
			RebalanceFrequency: valueobject.RebalanceWeekly,
			MaxPositionCount:   1,
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Job.ID == 0 {
		t.Fatal("expected generated job id")
	}

	res, err := svc.Run(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Job.Status != valueobject.JobStatusCompleted {
		t.Fatalf("expected COMPLETED, got %s", res.Job.Status)
	}
	if res.Report == nil || res.Report.TradeCount == 0 {
		t.Fatalf("expected non-empty report, got %+v", res.Report)
	}

	gotJob, err := svc.Get(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if gotJob.Status != valueobject.JobStatusCompleted {
		t.Errorf("expected COMPLETED on reload, got %s", gotJob.Status)
	}
	gotReport, err := svc.GetReport(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if gotReport.TradeCount == 0 {
		t.Errorf("expected positive trade count, got %+v", gotReport)
	}
	gotTrades, err := svc.GetTrades(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("get trades: %v", err)
	}
	if len(gotTrades) == 0 {
		t.Errorf("expected trades to round-trip, got 0")
	}
	gotSnaps, err := svc.GetSnapshots(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("get snapshots: %v", err)
	}
	if len(gotSnaps) != len(calendar) {
		t.Errorf("expected %d snapshots, got %d", len(calendar), len(gotSnaps))
	}
}

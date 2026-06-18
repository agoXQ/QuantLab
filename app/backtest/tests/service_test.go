package tests

import (
	"context"
	"testing"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	domevent "github.com/agoXQ/QuantLab/app/backtest/domain/event"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	infraEvent "github.com/agoXQ/QuantLab/app/backtest/infrastructure/event"
	infraMarketData "github.com/agoXQ/QuantLab/app/backtest/infrastructure/marketdata"
	infraMatching "github.com/agoXQ/QuantLab/app/backtest/infrastructure/matching"
	infraMemory "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/memory"
	infraStrategy "github.com/agoXQ/QuantLab/app/backtest/infrastructure/strategy"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	formulaDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"
	formulaInfraEval "github.com/agoXQ/QuantLab/app/formula/infrastructure/evaluator"
	formulaInfraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	formulaInfraInd "github.com/agoXQ/QuantLab/app/formula/infrastructure/indicators"
	formulaInfraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	formulaInfraOpt "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	formulaInfraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	formulaInfraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	formulaInfraVal "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	formulaInfraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
	formulaSeries "github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// fixture wires the same dependency graph the production servicecontext
// builds, but with in-memory adapters so the test stays sandbox-friendly.
type fixture struct {
	svc      appBacktest.Service
	provider *infraMarketData.InMemory
	formula  *formulaDataport.InMemory
	clock    time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	clock := time.Date(2024, 6, 28, 15, 0, 0, 0, time.UTC)
	now := func() time.Time { return clock }

	publisher := infraEvent.Noop{}
	provider := infraMarketData.NewInMemory()
	formulaPort := formulaDataport.NewInMemory()

	funcReg := formulaInfraFunc.NewRegistry()
	varReg := formulaInfraVar.NewRegistry()
	formulaSvc := appFormula.NewService(
		formulaInfraLexer.NewLexer(),
		formulaInfraParser.NewParser(funcReg, varReg),
		formulaInfraVal.NewValidator(funcReg, varReg),
		formulaInfraOpt.NewOptimizer(),
		formulaInfraPlanner.NewPlanner(),
		funcReg,
	)
	evaluator := formulaInfraEval.New(formulaInfraInd.NewLibrary(), varReg)
	evalSvc := appFormula.NewEvaluatorService(formulaSvc, evaluator)
	executor := infraStrategy.NewFormulaExecutor(evalSvc, formulaPort)

	matching := infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{
		CommissionRate: 0.0003,
		StampDutyRate:  0.001,
		MinCommission:  0,
		SlippageRate:   0,
		Now:            now,
	})

	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       infraMemory.NewJobRepository(),
		Orders:     infraMemory.NewOrderRepository(),
		Trades:     infraMemory.NewTradeRepository(),
		Portfolios: infraMemory.NewPortfolioRepository(),
		Reports:    infraMemory.NewReportRepository(),
		Executor:   executor,
		Matching:   matching,
		MarketData: provider,
		Publisher:  publisher,
		Clock:      now,
	})

	return &fixture{svc: svc, provider: provider, formula: formulaPort, clock: clock}
}

// daily builds n consecutive UTC trading-day timestamps starting at start.
func daily(start time.Time, n int) []time.Time {
	out := make([]time.Time, n)
	for i := 0; i < n; i++ {
		out[i] = start.AddDate(0, 0, i)
	}
	return out
}

// linearBars generates closes that grow linearly so MA crosses are easy
// to predict, with open == close to keep slippage out of the picture.
func linearBars(days []time.Time, base, slope float64) []dommarket.Bar {
	bars := make([]dommarket.Bar, len(days))
	for i, d := range days {
		px := base + slope*float64(i)
		bars[i] = dommarket.Bar{
			TradeDate: d,
			Open:      px,
			High:      px,
			Low:       px,
			Close:     px,
			Volume:    1000,
			Amount:    px * 1000,
		}
	}
	return bars
}

// formulaBars converts a market-data slice into the Formula bar shape so
// the Formula DataPort can answer indicator queries from the same numbers.
func formulaBars(bars []dommarket.Bar) []formulaSeries.Bar {
	out := make([]formulaSeries.Bar, len(bars))
	for i, b := range bars {
		out[i] = formulaSeries.Bar{
			Timestamp: b.TradeDate,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    float64(b.Volume),
			Amount:    b.Amount,
		}
	}
	return out
}

func ctxBg() context.Context { return context.Background() }

func TestService_FullRun_FilterFinancials(t *testing.T) {
	fx := newFixture(t)

	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 30)
	good := linearBars(calendar, 10, 0.10) // monotonically up
	bad := linearBars(calendar, 20, -0.05) // monotonically down

	fx.provider.SetBars("000001", good)
	fx.provider.SetBars("000002", bad)
	fx.provider.SetCalendar(calendar)

	// Formula port mirrors prices for indicator-based formulas. We rely on
	// financial scalars here so we can also seed PE/ROE.
	fx.formula.SetBars("000001", formulaBars(good))
	fx.formula.SetBars("000002", formulaBars(bad))
	fx.formula.SetFinancials("000001", map[string]float64{"ROE": 22, "PE": 12})
	fx.formula.SetFinancials("000002", map[string]float64{"ROE": 6, "PE": 35})

	created, err := fx.svc.Create(ctxBg(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 15 AND PE < 20",
		Universe:       []string{"000001", "000002"},
		InitialCapital: 1_000_000,
		Range: valueobject.DateRange{
			Start: calendar[0],
			End:   calendar[len(calendar)-1],
		},
		Config: backtestjob.Config{
			RebalanceFrequency: valueobject.RebalanceWeekly,
			MaxPositionCount:   2,
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Job.Status != valueobject.JobStatusCreated {
		t.Fatalf("expected status CREATED, got %s", created.Job.Status)
	}

	res, err := fx.svc.Run(ctxBg(), created.Job.ID)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Job.Status != valueobject.JobStatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", res.Job.Status)
	}
	if res.Report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(res.Snapshots) != len(calendar) {
		t.Fatalf("expected %d snapshots, got %d", len(calendar), len(res.Snapshots))
	}
	if res.Report.TotalReturn <= 0 {
		t.Errorf("expected positive total return for the up-trending stock, got %v", res.Report.TotalReturn)
	}

	// All buys should land on 000001 (ROE filter rejects 000002).
	buyCodes := map[string]int{}
	for _, tr := range res.Trades {
		if tr.Side == valueobject.OrderSideBuy {
			buyCodes[tr.StockCode]++
		}
	}
	if buyCodes["000002"] != 0 {
		t.Errorf("did not expect any BUY on 000002, got %d", buyCodes["000002"])
	}
	if buyCodes["000001"] == 0 {
		t.Errorf("expected at least one BUY on 000001")
	}
}

func TestService_RunIdempotent(t *testing.T) {
	fx := newFixture(t)

	start := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 10)
	bars := linearBars(calendar, 50, 0.5)
	fx.provider.SetBars("000300", bars)
	fx.provider.SetCalendar(calendar)
	fx.formula.SetBars("000300", formulaBars(bars))
	fx.formula.SetFinancials("000300", map[string]float64{"ROE": 30})

	created, err := fx.svc.Create(ctxBg(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 10",
		Universe:       []string{"000300"},
		InitialCapital: 100_000,
		Range:          valueobject.DateRange{Start: calendar[0], End: calendar[len(calendar)-1]},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	first, err := fx.svc.Run(ctxBg(), created.Job.ID)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	second, err := fx.svc.Run(ctxBg(), created.Job.ID)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if first.Report == nil || second.Report == nil {
		t.Fatalf("expected non-nil reports on both runs")
	}
	if first.Report.FinalAsset != second.Report.FinalAsset {
		t.Errorf("expected idempotent final_asset, got %v vs %v",
			first.Report.FinalAsset, second.Report.FinalAsset)
	}
}

func TestService_ValidationErrors(t *testing.T) {
	fx := newFixture(t)
	cases := []struct {
		name string
		req  appBacktest.CreateBacktestRequest
	}{
		{name: "missing formula", req: appBacktest.CreateBacktestRequest{
			Universe: []string{"000001"}, InitialCapital: 1, Range: valueobject.DateRange{Start: time.Now(), End: time.Now()},
		}},
		{name: "missing universe", req: appBacktest.CreateBacktestRequest{
			Formula: "ROE > 0", InitialCapital: 1, Range: valueobject.DateRange{Start: time.Now(), End: time.Now()},
		}},
		{name: "zero capital", req: appBacktest.CreateBacktestRequest{
			Formula: "ROE > 0", Universe: []string{"000001"}, Range: valueobject.DateRange{Start: time.Now(), End: time.Now()},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := fx.svc.Create(ctxBg(), tc.req); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

// recordingPublisher captures every published event so tests can assert on
// the lifecycle envelope.
type recordingPublisher struct{ events []domevent.Event }

func (r *recordingPublisher) Publish(_ context.Context, e domevent.Event) error {
	r.events = append(r.events, e)
	return nil
}

func TestService_PublishesLifecycleEvents(t *testing.T) {
	fx := newFixture(t)
	rec := &recordingPublisher{}

	// Rebuild service with the recording publisher; everything else
	// reuses the fixture's wiring.
	fx.svc = appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       infraMemory.NewJobRepository(),
		Orders:     infraMemory.NewOrderRepository(),
		Trades:     infraMemory.NewTradeRepository(),
		Portfolios: infraMemory.NewPortfolioRepository(),
		Reports:    infraMemory.NewReportRepository(),
		Executor:   nil, // will be patched below
		Matching:   infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{}),
		MarketData: fx.provider,
		Publisher:  rec,
		Clock:      func() time.Time { return fx.clock },
	})

	created, err := fx.svc.Create(ctxBg(), appBacktest.CreateBacktestRequest{
		Formula:        "ROE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100_000,
		Range:          valueobject.DateRange{Start: fx.clock, End: fx.clock.AddDate(0, 0, 5)},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Job.Status != valueobject.JobStatusCreated {
		t.Fatalf("expected created, got %s", created.Job.Status)
	}
	if len(rec.events) != 1 || rec.events[0].EventType != domevent.EventBacktestCreated {
		t.Fatalf("expected BacktestCreated event, got %+v", rec.events)
	}
}

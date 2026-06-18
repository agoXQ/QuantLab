package tests

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	btvo "github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	infraEvent "github.com/agoXQ/QuantLab/app/backtest/infrastructure/event"
	infraBtMd "github.com/agoXQ/QuantLab/app/backtest/infrastructure/marketdata"
	infraMatching "github.com/agoXQ/QuantLab/app/backtest/infrastructure/matching"
	infraMemory "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/memory"
	infraStrategy "github.com/agoXQ/QuantLab/app/backtest/infrastructure/strategy"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"
	infraEvaluator "github.com/agoXQ/QuantLab/app/formula/infrastructure/evaluator"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraIndicators "github.com/agoXQ/QuantLab/app/formula/infrastructure/indicators"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOpt "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	marketPg "github.com/agoXQ/QuantLab/app/market/infrastructure/postgres"
)

// TestMarketData_EndToEndRun seeds the Market Data Postgres tables with a
// synthetic stock + calendar, then drives a Backtest Job through the full
// engine using FromMarketService (Backtest's marketdata.Provider) and
// RepositoryDataPort (Formula's evaluator.DataPort) against the very same
// database.
//
// The test is the wiring contract for the platform stack: if this passes,
// any Tushare-ingested data lands at the engine without further glue.
//
// MARKET_TEST_DSN gates the run so CI machines without docker keep
// skipping. Use the same DSN the Market Data service writes to:
//
//   MARKET_TEST_DSN=postgres://quantlab:quantlab_dev@localhost:5432/quantlab_market_data?sslmode=disable
func TestMarketData_EndToEndRun(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MARKET_TEST_DSN"))
	if dsn == "" {
		t.Skip("MARKET_TEST_DSN not set; skipping market data wiring test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("open market data db: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("ping market data db: %v", err)
	}
	if err := marketPg.EnsureSchema(ctx, db); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Reset the surface we are about to write so reruns stay deterministic.
	cleanCtx, cancelClean := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelClean()
	for _, stmt := range []string{
		`DELETE FROM market_bar    WHERE stock_code = '000001'`,
		`DELETE FROM trading_calendar WHERE trade_date BETWEEN '2024-04-01' AND '2024-04-30'`,
	} {
		if _, err := db.ExecContext(cleanCtx, stmt); err != nil {
			t.Fatalf("clean fixture: %v", err)
		}
	}

	// Seed: 12 daily bars + matching calendar entries.
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	days := make([]time.Time, 12)
	for i := range days {
		days[i] = start.AddDate(0, 0, i)
	}
	bars := make([]*marketbar.MarketBar, len(days))
	for i, d := range days {
		px := 25 + 0.4*float64(i)
		bars[i] = &marketbar.MarketBar{
			StockCode: "000001",
			TradeDate: d,
			Period:    marketVO.PeriodDay,
			Open:      px,
			High:      px,
			Low:       px,
			Close:     px,
			Volume:    1000,
			Amount:    px * 1000,
			AdjFactor: 1,
		}
	}
	if err := marketPg.NewMarketBarRepository(db).BulkUpsert(ctx, bars); err != nil {
		t.Fatalf("seed bars: %v", err)
	}
	td := make([]calendar.TradingDay, len(days))
	for i, d := range days {
		td[i] = calendar.TradingDay{TradeDate: d, IsOpen: true}
	}
	if err := marketPg.NewCalendarRepository(db).BulkUpsert(ctx, td); err != nil {
		t.Fatalf("seed calendar: %v", err)
	}

	// Wire the same stack ServiceContext.buildMarketStack would build.
	deps := appMarket.Dependencies{
		Securities:   marketPg.NewSecurityRepository(db),
		Bars:         marketPg.NewMarketBarRepository(db),
		Financials:   marketPg.NewFinancialRepository(db),
		Factors:      marketPg.NewFactorRepository(db),
		Indexes:      marketPg.NewIndexBarRepository(db),
		Calendar:     marketPg.NewCalendarRepository(db),
		DataVersions: marketPg.NewDataVersionRepository(db),
		Adjuster:     infraAdj.NewFactorAdjuster(),
	}
	marketSvc := appMarket.NewService(deps)
	provider := infraBtMd.NewFromMarketService(marketSvc, marketVO.AdjustmentPre)
	dataPort, err := infraDataport.NewRepository(infraDataport.RepositoryConfig{
		Bars:       deps.Bars,
		Financials: deps.Financials,
		Factors:    deps.Factors,
		Adjuster:   deps.Adjuster,
		Adjustment: marketVO.AdjustmentPre,
	})
	if err != nil {
		t.Fatalf("build repository data port: %v", err)
	}

	// Formula stack identical to the production composition root.
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()
	formulaSvc := appFormula.NewService(
		infraLexer.NewLexer(),
		infraParser.NewParser(funcReg, varReg),
		infraValidator.NewValidator(funcReg, varReg),
		infraOpt.NewOptimizer(),
		infraPlanner.NewPlanner(),
		funcReg,
	)
	evaluator := infraEvaluator.New(infraIndicators.NewLibrary(), varReg)
	evalSvc := appFormula.NewEvaluatorService(formulaSvc, evaluator)
	executor := infraStrategy.NewFormulaExecutor(evalSvc, dataPort)

	clock := func() time.Time { return time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC) }
	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       infraMemory.NewJobRepository(),
		Orders:     infraMemory.NewOrderRepository(),
		Trades:     infraMemory.NewTradeRepository(),
		Portfolios: infraMemory.NewPortfolioRepository(),
		Reports:    infraMemory.NewReportRepository(),
		Executor:   executor,
		Matching:   infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{}),
		MarketData: provider,
		Publisher:  infraEvent.Noop{},
		Clock:      clock,
	})

	// CLOSE > 0 selects every bar; the engine should pick up 000001 and
	// generate at least one trade against the seeded prices.
	created, err := svc.Create(context.Background(), appBacktest.CreateBacktestRequest{
		Formula:        "CLOSE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100_000,
		Range:          btvo.DateRange{Start: days[0], End: days[len(days)-1]},
		Config: backtestjob.Config{
			RebalanceFrequency: btvo.RebalanceWeekly,
			MaxPositionCount:   1,
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	res, err := svc.Run(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Job.Status != btvo.JobStatusCompleted {
		t.Fatalf("expected COMPLETED, got %s", res.Job.Status)
	}
	if res.Report == nil || res.Report.TradeCount == 0 {
		t.Fatalf("expected non-empty report, got %+v", res.Report)
	}
}

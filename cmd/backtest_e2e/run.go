package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
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
	infraMarketEvent "github.com/agoXQ/QuantLab/app/market/infrastructure/event"
	infraMarketAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	marketPg "github.com/agoXQ/QuantLab/app/market/infrastructure/postgres"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/provider/tushare"
)

// platform groups the dependencies the harness needs to run a scenario
// end-to-end. It is the same shape the production servicecontext builds,
// but stripped to what the smoke test exercises.
type platform struct {
	db        *sql.DB
	market    appMarket.Service
	ingest    appMarket.IngestionService
	executor  *infraStrategy.FormulaExecutor
	provider  *infraBtMd.FromMarketService
	close     func() error
}

// platformConfig captures the inputs needed to construct the harness'
// dependency graph.
type platformConfig struct {
	dsn          string
	tushareToken string
}

// buildPlatform opens the Market Data database and wires the seven
// repositories, the ingestion service (when a Tushare token is present),
// the Backtest market provider, and the Formula evaluator. The harness
// shares one *sql.DB across both data paths.
func buildPlatform(ctx context.Context, cfg platformConfig) (*platform, error) {
	if strings.TrimSpace(cfg.dsn) == "" {
		return nil, fmt.Errorf("market data DSN is required (set MARKET_DATA_DSN or pass -dsn)")
	}
	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := marketPg.EnsureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	deps := appMarket.Dependencies{
		Securities:   marketPg.NewSecurityRepository(db),
		Bars:         marketPg.NewMarketBarRepository(db),
		Financials:   marketPg.NewFinancialRepository(db),
		Factors:      marketPg.NewFactorRepository(db),
		Indexes:      marketPg.NewIndexBarRepository(db),
		Calendar:     marketPg.NewCalendarRepository(db),
		DataVersions: marketPg.NewDataVersionRepository(db),
		Adjuster:     infraMarketAdj.NewFactorAdjuster(),
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
		_ = db.Close()
		return nil, fmt.Errorf("build repository data port: %w", err)
	}

	executor := buildFormulaExecutor(dataPort)

	pf := &platform{
		db:       db,
		market:   marketSvc,
		executor: executor,
		provider: provider,
		close:    db.Close,
	}

	if token := strings.TrimSpace(cfg.tushareToken); token != "" {
		client, err := tushare.NewClient(token)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("build tushare client: %w", err)
		}
		dataProvider := tushare.NewProvider(client)

		pf.ingest = appMarket.NewIngestionService(appMarket.IngestionDeps{
			Provider:     dataProvider,
			Securities:   deps.Securities,
			Bars:         deps.Bars,
			Financials:   deps.Financials,
			Factors:      deps.Factors,
			Indexes:      deps.Indexes,
			Calendar:     deps.Calendar,
			DataVersions: deps.DataVersions,
			Publisher:    infraMarketEvent.Noop{},
			Clock:        time.Now,
		})
	} else {
		log.Printf("[platform] TUSHARE_TOKEN missing; ingestion disabled")
	}

	return pf, nil
}

// buildFormulaExecutor wires the Formula evaluator with the same
// composition the Formula service uses, minus the cache / event / log
// decorators (the harness compiles each formula at most once per run).
func buildFormulaExecutor(dataPort *infraDataport.RepositoryDataPort) *infraStrategy.FormulaExecutor {
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
	return infraStrategy.NewFormulaExecutor(evalSvc, dataPort)
}

// runBacktest drives a single backtest replay against the harness'
// platform. The Backtest application service uses in-memory job /
// order / trade / report repositories so the harness does not need a
// dedicated database; the Market Data DB stays the only side-effect.
func runBacktest(ctx context.Context, pf *platform, sc *Scenario, dataVersion string) (*appBacktest.RunResult, error) {
	rng, err := sc.BacktestRange()
	if err != nil {
		return nil, err
	}
	cfg, err := sc.BacktestConfig()
	if err != nil {
		return nil, err
	}

	clock := func() time.Time { return Now() }
	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       infraMemory.NewJobRepository(),
		Orders:     infraMemory.NewOrderRepository(),
		Trades:     infraMemory.NewTradeRepository(),
		Portfolios: infraMemory.NewPortfolioRepository(),
		Reports:    infraMemory.NewReportRepository(),
		Executor:   pf.executor,
		Matching:   infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{}),
		MarketData: pf.provider,
		Publisher:  infraEvent.Noop{},
		Clock:      clock,
	})

	created, err := svc.Create(ctx, appBacktest.CreateBacktestRequest{
		Name:           sc.Name,
		Formula:        sc.Backtest.Formula,
		Universe:       sc.Backtest.Universe,
		Benchmark:      sc.Backtest.Benchmark,
		DataVersion:    chooseDataVersion(sc.Backtest.DataVersion, dataVersion),
		InitialCapital: sc.Backtest.InitialCapital,
		Range:          rng,
		Config:         cfg,
	})
	if err != nil {
		return nil, fmt.Errorf("create backtest: %w", err)
	}
	res, err := svc.Run(ctx, created.Job.ID)
	if err != nil {
		return nil, fmt.Errorf("run backtest: %w", err)
	}
	return res, nil
}

// chooseDataVersion prefers the scenario-pinned version when set so
// reruns can target an earlier ingestion.
func chooseDataVersion(scenarioPin, ingested string) string {
	if strings.TrimSpace(scenarioPin) != "" {
		return scenarioPin
	}
	return ingested
}

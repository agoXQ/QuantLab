// Package svc is the composition root for the Backtest Engine service. It
// wires the application service, the in-memory or Postgres-backed
// repositories, and the Formula / Market Data adapters chosen by config.
package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	domevent "github.com/agoXQ/QuantLab/app/backtest/domain/event"
	domexec "github.com/agoXQ/QuantLab/app/backtest/domain/executor"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	domorder "github.com/agoXQ/QuantLab/app/backtest/domain/order"
	domportfolio "github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	domreport "github.com/agoXQ/QuantLab/app/backtest/domain/report"
	domtrade "github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/internal/config"
	infraEvent "github.com/agoXQ/QuantLab/app/backtest/infrastructure/event"
	infraMarketData "github.com/agoXQ/QuantLab/app/backtest/infrastructure/marketdata"
	infraMatching "github.com/agoXQ/QuantLab/app/backtest/infrastructure/matching"
	infraMemory "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/postgres"
	infraStrategy "github.com/agoXQ/QuantLab/app/backtest/infrastructure/strategy"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
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
)

// ServiceContext is the composition root for the Backtest Engine service.
type ServiceContext struct {
	Config      config.Config
	BacktestSvc appBacktest.Service
	// MarketProvider is the in-memory market provider. It is non-nil only
	// when the service falls back to the in-memory adapters (CI smoke
	// tests, local exploration without the Market Data database). In
	// production both data paths read from the Market Data Postgres
	// tables and this field stays nil.
	MarketProvider *infraMarketData.InMemory
	// FormulaPort is the in-memory formula data port and is set under the
	// same conditions as MarketProvider above.
	FormulaPort *formulaDataport.InMemory
	Executor    domexec.StrategyExecutor
	DB          *sql.DB
	// MarketDataDB is the *sql.DB connected to the Market Data database
	// when the platform stack is wired; nil when running in-memory.
	MarketDataDB *sql.DB
}

// NewServiceContext wires the dependency graph.
//
// Repository selection: when Postgres.DSN is configured we wire the
// Postgres-backed repositories and (optionally) ensure the schema; when
// the DSN is missing or the connection fails we fall back to the in-memory
// repositories so the binary still boots for CI smoke tests and local
// exploration. The Formula / Market data adapters degrade the same way.
func NewServiceContext(c config.Config) *ServiceContext {
	publisher := buildPublisher(c)
	matching := infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{})

	// Pick the data sources. Backtest end-to-end runs need the same bars
	// on both sides: the engine itself walks the calendar through
	// marketdata.Provider, and the FormulaExecutor reads bars / financials
	// through evaluator.DataPort. buildMarketStack wires both off the same
	// Market Data Postgres database when MarketData.DSN is configured;
	// otherwise we fall back to the in-memory adapters so the binary
	// still boots without the platform DB.
	stack, err := buildMarketStack(c.MarketData)
	if err != nil {
		log.Printf("[backtest] warning: market data stack unavailable: %v; falling back to in-memory", err)
		stack = marketStack{}
	}

	var (
		marketProvider  dommarket.Provider
		formulaDataPort domainEval.DataPort
		inMemMarket     *infraMarketData.InMemory
		inMemFormula    *formulaDataport.InMemory
	)
	if stack.provider != nil && stack.dataPort != nil {
		marketProvider = stack.provider
		formulaDataPort = stack.dataPort
	} else {
		inMemMarket = infraMarketData.NewInMemory()
		inMemFormula = formulaDataport.NewInMemory()
		marketProvider = inMemMarket
		formulaDataPort = inMemFormula
	}

	_, evaluatorSvc := buildFormulaStack()
	executor := infraStrategy.NewFormulaExecutor(evaluatorSvc, formulaDataPort)

	jobs, orders, trades, portfolios, reports, db := buildRepositories(c)

	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       jobs,
		Orders:     orders,
		Trades:     trades,
		Portfolios: portfolios,
		Reports:    reports,
		Executor:   executor,
		Matching:   matching,
		MarketData: marketProvider,
		Publisher:  publisher,
		Clock:      time.Now,
	})

	return &ServiceContext{
		Config:         c,
		BacktestSvc:    svc,
		MarketProvider: inMemMarket,
		FormulaPort:    inMemFormula,
		Executor:       executor,
		DB:             db,
		MarketDataDB:   stack.db,
	}
}


// buildRepositories returns the repository set and the underlying *sql.DB
// (nil when running in-memory). The function never panics: any Postgres
// failure logs a warning and falls back to in-memory so a temporary DB
// outage does not take the whole service down during local development.
func buildRepositories(c config.Config) (
	backtestjob.Repository,
	domorder.Repository,
	domtrade.Repository,
	domportfolio.Repository,
	domreport.Repository,
	*sql.DB,
) {
	if c.Postgres.DSN == "" {
		log.Printf("[backtest] postgres DSN empty, using in-memory repositories")
		return infraMemory.NewJobRepository(),
			infraMemory.NewOrderRepository(),
			infraMemory.NewTradeRepository(),
			infraMemory.NewPortfolioRepository(),
			infraMemory.NewReportRepository(),
			nil
	}

	db, err := openPostgres(c.Postgres)
	if err != nil {
		log.Printf("[backtest] warning: postgres unavailable: %v; falling back to in-memory", err)
		return infraMemory.NewJobRepository(),
			infraMemory.NewOrderRepository(),
			infraMemory.NewTradeRepository(),
			infraMemory.NewPortfolioRepository(),
			infraMemory.NewReportRepository(),
			nil
	}
	if c.Postgres.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := infraPg.EnsureSchema(ctx, db); err != nil {
			log.Printf("[backtest] warning: ensure schema: %v", err)
		}
	}
	log.Printf("[backtest] postgres repositories wired")
	return infraPg.NewJobRepository(db),
		infraPg.NewOrderRepository(db),
		infraPg.NewTradeRepository(db),
		infraPg.NewPortfolioRepository(db),
		infraPg.NewReportRepository(db),
		db
}

// openPostgres opens the connection, applies pool sizing, and pings.
func openPostgres(cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

// buildPublisher returns a Kafka-backed publisher when Brokers is set,
// otherwise the no-op publisher so MVP runs locally.
func buildPublisher(c config.Config) domevent.Publisher {
	if len(c.Kafka.Brokers) == 0 {
		return infraEvent.Noop{}
	}
	return infraEvent.NewKafkaPublisher(c.Kafka.Brokers)
}

// buildFormulaStack wires a minimal in-process Formula stack: lexer, parser,
// validator, optimizer, planner, and the AST evaluator. We deliberately
// skip the cache / event / log decorators here; backtest replays compile
// the same formula many times so any meaningful caching belongs at the
// EvaluatorService layer or higher, which the MVP can defer.
func buildFormulaStack() (appFormula.Service, appFormula.EvaluatorService) {
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
	return formulaSvc, evalSvc
}

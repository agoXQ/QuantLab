// Package svc is the composition root for the Backtest Engine service. It
// wires the application service, the in-memory or Postgres-backed
// repositories, and the Formula / Market Data adapters chosen by config.
package svc

import (
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	domevent "github.com/agoXQ/QuantLab/app/backtest/domain/event"
	domexec "github.com/agoXQ/QuantLab/app/backtest/domain/executor"
	"github.com/agoXQ/QuantLab/app/backtest/internal/config"
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
)

// ServiceContext is the composition root for the Backtest Engine service.
type ServiceContext struct {
	Config         config.Config
	BacktestSvc    appBacktest.Service
	MarketProvider *infraMarketData.InMemory
	FormulaPort    *formulaDataport.InMemory
	Executor       domexec.StrategyExecutor
}

// NewServiceContext wires the dependency graph.
//
// External dependencies degrade gracefully: missing Kafka -> Noop publisher,
// missing Postgres -> in-memory repos. The Formula stack is always built
// in-process because the MVP runs the engine inside the same binary; the
// gRPC adapter slots in alongside the in-process executor without changing
// the application service.
func NewServiceContext(c config.Config) *ServiceContext {
	publisher := buildPublisher(c)
	jobs := infraMemory.NewJobRepository()
	orders := infraMemory.NewOrderRepository()
	trades := infraMemory.NewTradeRepository()
	portfolios := infraMemory.NewPortfolioRepository()
	reports := infraMemory.NewReportRepository()
	matching := infraMatching.NewNextOpenEngine(infraMatching.EngineConfig{})

	provider := infraMarketData.NewInMemory()
	formulaPort := formulaDataport.NewInMemory()

	_, evaluatorSvc := buildFormulaStack()
	executor := infraStrategy.NewFormulaExecutor(evaluatorSvc, formulaPort)

	svc := appBacktest.NewService(appBacktest.Dependencies{
		Jobs:       jobs,
		Orders:     orders,
		Trades:     trades,
		Portfolios: portfolios,
		Reports:    reports,
		Executor:   executor,
		Matching:   matching,
		MarketData: provider,
		Publisher:  publisher,
		Clock:      time.Now,
	})

	return &ServiceContext{
		Config:         c,
		BacktestSvc:    svc,
		MarketProvider: provider,
		FormulaPort:    formulaPort,
		Executor:       executor,
	}
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

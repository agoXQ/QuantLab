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
	infraMemQueue "github.com/agoXQ/QuantLab/app/backtest/infrastructure/queue/memory"
	infraMemory "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/postgres"
	infraStrategy "github.com/agoXQ/QuantLab/app/backtest/infrastructure/strategy"
	infraStrategySync "github.com/agoXQ/QuantLab/app/backtest/infrastructure/strategysync"
	appStrategySync "github.com/agoXQ/QuantLab/app/backtest/application/strategysync"
	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
	infraWorker "github.com/agoXQ/QuantLab/app/backtest/infrastructure/worker"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"

	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"

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
	// Queue / Workers carry the async pipeline. They are non-nil when
	// QueueConfig.Enabled is true; the Close hook below drains them
	// during graceful shutdown.
	Queue   domqueue.Queue
	Workers *infraWorker.Pool

	// strategySyncCancel stops the Strategy events consumer; nil when
	// the consumer is disabled in config.
	strategySyncCancel context.CancelFunc
	strategySyncDone   chan struct{}
	strategySyncCloser interface{ Close() error }
	strategyClient     zrpc.Client
	// StrategyResolver resolves strategy version → formula for the gRPC
	// CreateBacktest logic layer. It is wired whenever Strategy endpoints
	// are configured, regardless of whether StrategySync is enabled.
	StrategyResolver domsync.Resolver
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

	var (
		queueImpl domqueue.Queue
		pool      *infraWorker.Pool
	)
	if c.Queue.Enabled {
		queueImpl = infraMemQueue.New(c.Queue.Buffer)
	}

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
		Queue:      queueImpl,
		Clock:      time.Now,
	})

	if queueImpl != nil {
		pool = infraWorker.New(queueImpl, runnerAdapter{svc: svc}, infraWorker.Config{
			Workers:    c.Queue.Workers,
			JobTimeout: time.Duration(c.Queue.JobTimeoutSeconds) * time.Second,
		})
		pool.Start(context.Background())
		log.Printf("[backtest] async pipeline started: workers=%d buffer=%d", c.Queue.Workers, c.Queue.Buffer)
	} else {
		log.Printf("[backtest] async pipeline disabled; only inline ?run=true is available")
	}

	// Recover from an ungraceful shutdown: re-enqueue jobs left in
	// QUEUED, fail jobs left in RUNNING. Run synchronously at boot so
	// we do not race incoming traffic; the budget is bounded by the
	// number of jobs in those states (capped at 500 each).
	reconcileCtx, reconcileCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if r, err := svc.Reconcile(reconcileCtx); err != nil {
		log.Printf("[backtest] warning: reconcile on boot: %v", err)
	} else if r.Requeued > 0 || r.FailedStuck > 0 {
		log.Printf("[backtest] reconcile: requeued=%d failed_stuck=%d inspected=%d",
			r.Requeued, r.FailedStuck, r.Inspected)
	}
	reconcileCancel()

	startStrategySync, syncCancel, syncDone, syncCloser, strategyClient := buildStrategySync(c, svc, marketProvider)

	// Build a strategy resolver for the gRPC CreateBacktest logic even
	// when StrategySync is disabled — the logic layer needs to resolve
	// formula_text from a strategy version id.
	var strategyResolver domsync.Resolver
	if strategyClient != nil {
		strategyResolver = infraStrategySync.NewGRPCResolver(strategyClient)
	} else {
		strategyResolver = buildStandaloneResolver(c)
	}

	sc := &ServiceContext{
		Config:             c,
		BacktestSvc:        svc,
		MarketProvider:     inMemMarket,
		FormulaPort:        inMemFormula,
		Executor:           executor,
		DB:                 db,
		MarketDataDB:       stack.db,
		Queue:              queueImpl,
		Workers:            pool,
		strategySyncCancel: syncCancel,
		strategySyncDone:   syncDone,
		strategySyncCloser: syncCloser,
		strategyClient:     strategyClient,
		StrategyResolver: strategyResolver,
	}
	if startStrategySync != nil {
		startStrategySync()
	}
	return sc
}

// Close drains the worker pool and closes the queue. Safe to call once;
// subsequent invocations are no-ops because the underlying queue and
// pool guard against double-close.
func (sc *ServiceContext) Close() {
	if sc == nil {
		return
	}
	if sc.strategySyncCancel != nil {
		sc.strategySyncCancel()
	}
	if sc.strategySyncDone != nil {
		<-sc.strategySyncDone
	}
	if sc.strategySyncCloser != nil {
		_ = sc.strategySyncCloser.Close()
	}
	if sc.Workers != nil {
		sc.Workers.Stop()
	}
	if sc.Queue != nil {
		_ = sc.Queue.Close()
	}
	if sc.DB != nil {
		_ = sc.DB.Close()
	}
	if sc.MarketDataDB != nil {
		_ = sc.MarketDataDB.Close()
	}
}

// runnerAdapter bridges the application Service to the Pool's Runner
// port without forcing the application package to depend on
// infrastructure/worker. The infrastructure layer takes the import; the
// application keeps the same surface area.
type runnerAdapter struct {
	svc appBacktest.Service
}

func (r runnerAdapter) RunQueued(ctx context.Context, jobID int64) error {
	return r.svc.RunQueued(ctx, jobID)
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


// buildStrategySync wires the Strategy events consumer + baseline
// handler. The function is forgiving: any configuration error is logged
// and the consumer is left disabled so a misconfigured Kafka broker
// list does not take the backtest service down. The returned start
// closure is invoked once the ServiceContext is built so the consumer
// only runs after the application service is ready.
func buildStrategySync(
	c config.Config,
	svc appBacktest.Service,
	marketProvider dommarket.Provider,
) (start func(), cancel context.CancelFunc, done chan struct{}, closer interface{ Close() error }, client zrpc.Client) {
	if !c.StrategySync.Enabled {
		return nil, nil, nil, nil, nil
	}
	if len(c.StrategySync.Brokers) == 0 {
		log.Printf("[backtest] strategy sync enabled but no Kafka brokers; skipping")
		return nil, nil, nil, nil, nil
	}
	clientConf, err := buildStrategyClientConf(c.StrategySync.Strategy)
	if err != nil {
		log.Printf("[backtest] strategy sync: invalid strategy client config: %v", err)
		return nil, nil, nil, nil, nil
	}
	cli, err := zrpc.NewClient(clientConf)
	if err != nil {
		log.Printf("[backtest] strategy sync: dial strategy service: %v", err)
		return nil, nil, nil, nil, nil
	}
	resolver := infraStrategySync.NewGRPCResolver(cli)

	baseline := c.StrategySync.Baseline
	handler := appStrategySync.NewBaselineHandler(appStrategySync.Config{
		Universe:       baseline.Universe,
		Lookback:       time.Duration(baseline.LookbackDays) * 24 * time.Hour,
		InitialCapital: baseline.InitialCapital,
		Benchmark:      baseline.Benchmark,
		AutoSubmit:     baseline.AutoSubmit,
		Tag:            baseline.Tag,
	}, resolver, svc, marketProvider, time.Now)

	consumer, err := infraStrategySync.NewConsumer(infraStrategySync.Config{
		Brokers: c.StrategySync.Brokers,
		Topic:   c.StrategySync.Topic,
		GroupID: c.StrategySync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[backtest] strategy sync: build consumer: %v", err)
		return nil, nil, nil, nil, cli
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	startFn := func() {
		go func() {
			defer close(doneCh)
			log.Printf("[backtest] strategy sync started topic=%s group=%s", c.StrategySync.Topic, c.StrategySync.GroupID)
			if err := consumer.Run(ctx); err != nil {
				log.Printf("[backtest] strategy sync stopped: %v", err)
			}
		}()
	}
	return startFn, cancelFn, doneCh, consumer, cli
}

// buildStrategyClientConf maps the slimmed-down config struct into the
// full zrpc.RpcClientConf. Direct endpoints take precedence; falling
// back to Etcd-discovered targets means MVP deployments running both
// services on a single host need only Endpoints set.
func buildStrategyClientConf(cfg config.StrategyClientConfig) (zrpc.RpcClientConf, error) {
	out := zrpc.RpcClientConf{
		Endpoints: cfg.Endpoints,
		Timeout:   cfg.Timeout,
		NonBlock:  cfg.NonBlock,
	}
	if len(cfg.Etcd.Hosts) > 0 {
		out.Etcd = discov.EtcdConf{Hosts: cfg.Etcd.Hosts, Key: cfg.Etcd.Key}
	}
	if len(out.Endpoints) == 0 && len(out.Etcd.Hosts) == 0 {
		return zrpc.RpcClientConf{}, fmt.Errorf("either Endpoints or Etcd.Hosts must be set")
	}
	return out, nil
}


// buildStandaloneResolver wires a gRPC strategy resolver when
// StrategySync is disabled but Strategy endpoints are still configured.
// This lets the gRPC CreateBacktest logic resolve formula_text without
// the full strategy-sync consumer pipeline.
func buildStandaloneResolver(c config.Config) domsync.Resolver {
	if len(c.StrategySync.Strategy.Endpoints) == 0 {
		return nil
	}
	clientConf, err := buildStrategyClientConf(c.StrategySync.Strategy)
	if err != nil {
		log.Printf("[backtest] standalone resolver: invalid strategy client config: %v", err)
		return nil
	}
	cli, err := zrpc.NewClient(clientConf)
	if err != nil {
		log.Printf("[backtest] standalone resolver: dial strategy service: %v", err)
		return nil
	}
	return infraStrategySync.NewGRPCResolver(cli)
}

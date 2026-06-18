package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	infraCache "github.com/agoXQ/QuantLab/app/formula/infrastructure/cache"
	infraDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"
	infraMarketAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	infraMarketPg "github.com/agoXQ/QuantLab/app/market/infrastructure/postgres"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraEvaluator "github.com/agoXQ/QuantLab/app/formula/infrastructure/evaluator"
	infraIndicators "github.com/agoXQ/QuantLab/app/formula/infrastructure/indicators"
	infraLog "github.com/agoXQ/QuantLab/app/formula/infrastructure/compilelog"
	infraEvent "github.com/agoXQ/QuantLab/app/formula/infrastructure/event"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraMetrics "github.com/agoXQ/QuantLab/app/formula/infrastructure/metrics"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	"github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
	"github.com/agoXQ/QuantLab/app/formula/internal/config"
)

// ServiceContext is the composition root for the Formula Engine service.
type ServiceContext struct {
	Config           config.Config
	FormulaService   formula.Service
	EvaluatorService formula.EvaluatorService
	// DataPort is the read-side adapter used by the evaluator. It is the
	// RepositoryDataPort when MarketData.DSN is configured, otherwise the
	// InMemory adapter used for tests.
	DataPort         domainEval.DataPort
	// InMemoryPort is exposed so tests and the local HTTP handler can
	// seed fixtures without reaching for the data port interface.
	InMemoryPort     *infraDataport.InMemory
	DB               *sql.DB
	Metrics          *infraMetrics.PrometheusCollector
}

// NewServiceContext creates a new ServiceContext with all dependencies wired.
func NewServiceContext(c config.Config) *ServiceContext {
	funcRegistry := function.NewRegistry()
	varRegistry := variable.NewRegistry()

	lexerImpl := lexer.NewLexer()
	parserImpl := parser.NewParser(funcRegistry, varRegistry)
	validatorImpl := validator.NewValidator(funcRegistry, varRegistry)
	optimizerImpl := optimizer.NewOptimizer()
	plannerImpl := planner.NewPlanner()

	baseService := formula.NewService(
		lexerImpl, parserImpl, validatorImpl,
		optimizerImpl, plannerImpl, funcRegistry,
	)

	// Initialize Prometheus metrics
	metrics := infraMetrics.NewPrometheusCollector()

	// Build the decorator chain (innermost to outermost):
	//   Metrics -> Cache -> Event -> Log
	svc := wrapWithMetrics(metrics, baseService)
	svc = wrapWithCache(c, svc, metrics)
	svc = wrapWithEventing(c, svc)
	var db *sql.DB
	svc = wrapWithCompileLog(c, svc, &db)

	// Build the evaluator pipeline. We pick the data port lazily based
	// on configuration: when MarketData.DSN is set we wire a
	// RepositoryDataPort that reads the Market Data tables directly
	// (Roadmap Phase 1 monolith mode); otherwise we fall back to the
	// in-memory port used by tests and local exploration. The
	// gRPC-backed adapter slots in here as a third option without
	// touching the rest of the graph.
	indicatorLib := infraIndicators.NewLibrary()
	evaluator := infraEvaluator.New(indicatorLib, varRegistry)
	evalSvc := formula.NewEvaluatorService(svc, evaluator)

	inMemoryPort := infraDataport.NewInMemory()
	var dataPort interface{} = inMemoryPort
	if repoPort, mdDB, err := buildRepositoryDataPort(c); err != nil {
		fmt.Printf("warning: repository data port unavailable: %v\n", err)
	} else if repoPort != nil {
		dataPort = repoPort
		if mdDB != nil && db == nil {
			db = mdDB
		}
	}
	_ = dataPort // surfaced via DataPort field below

	port, _ := dataPort.(domainEval.DataPort)
	if port == nil {
		port = inMemoryPort
	}

	return &ServiceContext{
		Config:           c,
		FormulaService:   svc,
		EvaluatorService: evalSvc,
		DataPort:         port,
		InMemoryPort:     inMemoryPort,
		DB:               db,
		Metrics:          metrics,
	}
}

func wrapWithMetrics(metrics *infraMetrics.PrometheusCollector, svc formula.Service) formula.Service {
	return formula.NewMetricsService(svc, metrics)
}

func wrapWithCache(c config.Config, svc formula.Service, metrics *infraMetrics.PrometheusCollector) formula.Service {
	rc := c.RedisCache
	if rc.Host == "" {
		return svc
	}
	addr := fmt.Sprintf("%s:%d", rc.Host, rc.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, Password: rc.Pass, DB: rc.DB,
	})
	cacheImpl := infraCache.NewRedis(rdb, time.Duration(rc.TTL)*time.Second)

	// CachedService is always created, but WithMetrics is set only when Redis is configured
	cachedSvc := formula.NewCachedService(svc, cacheImpl, 0).(*formula.CachedService)
	cachedSvc.WithMetrics(metrics)
	return cachedSvc
}

func wrapWithEventing(c config.Config, svc formula.Service) formula.Service {
	if len(c.Kafka.Brokers) == 0 {
		return svc
	}
	publisher := infraEvent.NewKafkaPublisher(c.Kafka.Brokers)
	return formula.NewEventingService(svc, publisher)
}

func wrapWithCompileLog(c config.Config, svc formula.Service, dbOut **sql.DB) formula.Service {
	dsn := c.Postgres.DSN
	if dsn == "" {
		return svc
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("warning: failed to open postgres for compile log: %v\n", err)
		return svc
	}
	*dbOut = db

	logRepo := infraLog.NewPostgresRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := logRepo.EnsureTable(ctx); err != nil {
		fmt.Printf("warning: failed to ensure compile_log table: %v\n", err)
	}

	return formula.NewLoggedService(svc, logRepo)
}


// buildRepositoryDataPort constructs the RepositoryDataPort when MarketData
// configuration is present. Returning a nil port and nil error means the
// caller should fall back to the in-memory adapter; a non-nil error signals
// a configuration problem worth surfacing in the logs.
func buildRepositoryDataPort(c config.Config) (*infraDataport.RepositoryDataPort, *sql.DB, error) {
	if c.MarketData.DSN == "" {
		return nil, nil, nil
	}
	db, err := sql.Open("postgres", c.MarketData.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("open market data db: %w", err)
	}
	if c.MarketData.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.MarketData.MaxOpenConns)
	}
	if c.MarketData.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.MarketData.MaxIdleConns)
	}
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping market data db: %w", err)
	}

	adjustmentMode := valueobject.Adjustment(c.MarketData.Adjustment)
	if !adjustmentMode.IsValid() {
		adjustmentMode = valueobject.AdjustmentPre
	}

	port, err := infraDataport.NewRepository(infraDataport.RepositoryConfig{
		Bars:       infraMarketPg.NewMarketBarRepository(db),
		Financials: infraMarketPg.NewFinancialRepository(db),
		Factors:    infraMarketPg.NewFactorRepository(db),
		Adjuster:   infraMarketAdj.NewFactorAdjuster(),
		Adjustment: adjustmentMode,
	})
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return port, db, nil
}

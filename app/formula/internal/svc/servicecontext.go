package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraCache "github.com/agoXQ/QuantLab/app/formula/infrastructure/cache"
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
	Config         config.Config
	FormulaService formula.Service
	DB             *sql.DB
	Metrics        *infraMetrics.PrometheusCollector
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

	return &ServiceContext{
		Config:         c,
		FormulaService: svc,
		DB:             db,
		Metrics:        metrics,
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

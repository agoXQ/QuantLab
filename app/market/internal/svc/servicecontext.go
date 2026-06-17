package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/cache"
	domainEvent "github.com/agoXQ/QuantLab/app/market/domain/event"
	domainProvider "github.com/agoXQ/QuantLab/app/market/domain/provider"
	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	infraCache "github.com/agoXQ/QuantLab/app/market/infrastructure/cache"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/env"
	infraEvent "github.com/agoXQ/QuantLab/app/market/infrastructure/event"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/postgres"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/provider/faketushare"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/provider/tushare"
	"github.com/agoXQ/QuantLab/app/market/internal/config"
)

// ServiceContext is the composition root for the Market Data service.
type ServiceContext struct {
	Config           config.Config
	DB               *sql.DB
	Redis            *redis.Client
	Provider         domainProvider.DataProvider
	MarketService    appMarket.Service
	IngestionService appMarket.IngestionService
	EventPublisher   domainEvent.Publisher
}

// NewServiceContext builds the dependency graph.
//
// External dependencies are best-effort: missing PostgreSQL or Redis means
// the corresponding feature degrades gracefully (e.g. no caching, no
// persistence). The service still boots so HTTP/gRPC handlers can return
// well-formed validation errors.
func NewServiceContext(c config.Config) *ServiceContext {
	loadEnv()

	db, err := openPostgres(c)
	if err != nil {
		log.Printf("[market] warning: postgres unavailable: %v", err)
	}

	rdb, err := openRedis(c)
	if err != nil {
		log.Printf("[market] warning: redis unavailable: %v", err)
	}

	publisher := buildPublisher(c)

	prov, err := buildProvider(c)
	if err != nil {
		log.Printf("[market] warning: data provider unavailable: %v", err)
	}

	deps := appMarket.Dependencies{
		Adjuster: infraAdj.NewFactorAdjuster(),
	}
	var ingestionDeps appMarket.IngestionDeps
	if db != nil {
		secRepo := postgres.NewSecurityRepository(db)
		barRepo := postgres.NewMarketBarRepository(db)
		finRepo := postgres.NewFinancialRepository(db)
		facRepo := postgres.NewFactorRepository(db)
		idxRepo := postgres.NewIndexBarRepository(db)
		calRepo := postgres.NewCalendarRepository(db)
		verRepo := postgres.NewDataVersionRepository(db)

		deps.Securities = secRepo
		deps.Bars = barRepo
		deps.Financials = finRepo
		deps.Factors = facRepo
		deps.Indexes = idxRepo
		deps.Calendar = calRepo
		deps.DataVersions = verRepo

		ingestionDeps = appMarket.IngestionDeps{
			Provider:     prov,
			Securities:   secRepo,
			Bars:         barRepo,
			Financials:   finRepo,
			Factors:      facRepo,
			Indexes:      idxRepo,
			Calendar:     calRepo,
			DataVersions: verRepo,
			Publisher:    publisher,
			Clock:        time.Now,
		}
	}

	var svc appMarket.Service
	if deps.DataVersions != nil {
		svc = appMarket.NewService(deps)
		if rdb != nil {
			svc = wrapWithCache(svc, rdb, time.Duration(c.RedisCache.TTL)*time.Second)
		}
	}

	var ingestionSvc appMarket.IngestionService
	if ingestionDeps.Provider != nil && ingestionDeps.DataVersions != nil {
		ingestionSvc = appMarket.NewIngestionService(ingestionDeps)
	}

	return &ServiceContext{
		Config:           c,
		DB:               db,
		Redis:            rdb,
		Provider:         prov,
		MarketService:    svc,
		IngestionService: ingestionSvc,
		EventPublisher:   publisher,
	}
}

func loadEnv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, ".env"),
			filepath.Join(wd, "..", ".env"),
			filepath.Join(wd, "..", "..", ".env"),
		)
	}
	if err := env.Load(candidates...); err != nil {
		log.Printf("[market] warning: load .env failed: %v", err)
	}
}

func openPostgres(c config.Config) (*sql.DB, error) {
	if c.Postgres.DSN == "" {
		return nil, nil
	}
	db, err := sql.Open("postgres", c.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if c.Postgres.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.Postgres.MaxOpenConns)
	}
	if c.Postgres.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.Postgres.MaxIdleConns)
	}
	if c.Postgres.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Postgres.ConnMaxLifetime) * time.Second)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if c.Postgres.AutoMigrate {
		migrateCtx, cancelMig := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelMig()
		if err := postgres.EnsureSchema(migrateCtx, db); err != nil {
			log.Printf("[market] warning: ensure schema: %v", err)
		}
	}
	return db, nil
}

func openRedis(c config.Config) (*redis.Client, error) {
	if c.RedisCache.Host == "" {
		return nil, nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.RedisCache.Host, c.RedisCache.Port),
		Password: c.RedisCache.Pass,
		DB:       c.RedisCache.DB,
	})
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

func buildPublisher(c config.Config) domainEvent.Publisher {
	if len(c.Kafka.Brokers) == 0 {
		return infraEvent.Noop{}
	}
	return infraEvent.NewKafkaPublisher(c.Kafka.Brokers)
}

func buildProvider(c config.Config) (domainProvider.DataProvider, error) {
	switch c.Provider.Driver {
	case "fake":
		return faketushare.NewProvider(), nil
	case "", "tushare":
		token := os.Getenv(c.Provider.TokenEnv)
		if token == "" {
			return nil, fmt.Errorf("missing env %s", c.Provider.TokenEnv)
		}
		opts := []tushare.Option{}
		if c.Provider.Endpoint != "" {
			opts = append(opts, tushare.WithEndpoint(c.Provider.Endpoint))
		}
		client, err := tushare.NewClient(token, opts...)
		if err != nil {
			return nil, err
		}
		return tushare.NewProvider(client), nil
	default:
		return nil, fmt.Errorf("unknown provider driver: %s", c.Provider.Driver)
	}
}

func wrapWithCache(inner appMarket.Service, rdb *redis.Client, ttl time.Duration) appMarket.Service {
	if rdb == nil {
		return inner
	}
	c := infraCache.NewRedis(rdb, ttl)
	return appMarket.NewCachedService(inner, ensureCache(c), ttl)
}

func ensureCache(c cache.Cache) cache.Cache { return c }

// Package svc is the composition root for the Strategy Service. It
// wires the application service against the in-memory or Postgres-
// backed repositories chosen by config and exposes the resulting
// Service for the HTTP / gRPC handlers to call.
package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
	infraEvent "github.com/agoXQ/QuantLab/app/strategy/infrastructure/event"
	infraMemory "github.com/agoXQ/QuantLab/app/strategy/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/strategy/infrastructure/repository/postgres"
	"github.com/agoXQ/QuantLab/app/strategy/internal/config"
)

// ServiceContext is the composition root for the Strategy Service.
type ServiceContext struct {
	Config      config.Config
	StrategySvc appStrategy.Service
	DB          *sql.DB
}

// NewServiceContext wires the dependency graph.
func NewServiceContext(c config.Config) *ServiceContext {
	publisher := buildPublisher(c)
	strategies, versions, forks, db := buildRepositories(c)

	svc := appStrategy.NewService(appStrategy.Dependencies{
		Strategies: strategies,
		Versions:   versions,
		Forks:      forks,
		Publisher:  publisher,
		Clock:      time.Now,
	})

	return &ServiceContext{
		Config:      c,
		StrategySvc: svc,
		DB:          db,
	}
}

// Close releases the underlying database handle. Safe to call multiple
// times.
func (sc *ServiceContext) Close() {
	if sc == nil {
		return
	}
	if sc.DB != nil {
		_ = sc.DB.Close()
	}
}

// buildRepositories returns the repository set and the underlying
// *sql.DB (nil when running in-memory). The function never panics; any
// Postgres failure logs a warning and falls back to in-memory so a
// transient outage does not take the whole service down during local
// development.
func buildRepositories(c config.Config) (
	domstrategy.Repository,
	domversion.Repository,
	domfork.Repository,
	*sql.DB,
) {
	if c.Postgres.DSN == "" {
		log.Printf("[strategy] postgres DSN empty, using in-memory repositories")
		return infraMemory.NewStrategyRepository(),
			infraMemory.NewVersionRepository(),
			infraMemory.NewForkRepository(),
			nil
	}
	db, err := openPostgres(c.Postgres)
	if err != nil {
		log.Printf("[strategy] warning: postgres unavailable: %v; falling back to in-memory", err)
		return infraMemory.NewStrategyRepository(),
			infraMemory.NewVersionRepository(),
			infraMemory.NewForkRepository(),
			nil
	}
	if c.Postgres.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := infraPg.EnsureSchema(ctx, db); err != nil {
			log.Printf("[strategy] warning: ensure schema: %v", err)
		}
	}
	log.Printf("[strategy] postgres repositories wired")
	return infraPg.NewStrategyRepository(db),
		infraPg.NewVersionRepository(db),
		infraPg.NewForkRepository(db),
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

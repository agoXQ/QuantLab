// Package svc is the composition root for the Notification Service.
// It wires the application service against the in-memory or Postgres-
// backed repositories chosen by config and exposes the resulting
// Service for the gRPC / HTTP handlers + the cross-service consumers.
package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/postgres"
	"github.com/agoXQ/QuantLab/app/notification/internal/config"
)

// ServiceContext is the composition root.
type ServiceContext struct {
	Config        config.Config
	Service       appNotif.Service
	Subscriptions domSub.Repository
	DB            *sql.DB

	runners []*syncRunner
}

// NewServiceContext wires the dependency graph.
func NewServiceContext(c config.Config) *ServiceContext {
	notifs, prefs, subs, db := buildRepositories(c)
	svc := appNotif.NewService(appNotif.Dependencies{
		Notifications: notifs,
		Preferences:   prefs,
		Subscriptions: subs,
		Clock:         time.Now,
	})

	sc := &ServiceContext{
		Config:        c,
		Service:       svc,
		Subscriptions: subs,
		DB:            db,
	}
	sc.runners = appendRunner(sc.runners, buildUserSync(c, svc))
	sc.runners = appendRunner(sc.runners, buildStrategySync(c, svc, subs))
	sc.runners = appendRunner(sc.runners, buildBacktestSync(c, svc))
	for _, r := range sc.runners {
		r.start()
	}
	return sc
}

// Close releases the database handle and shuts down every cross-
// service event consumer. Safe to call multiple times.
func (sc *ServiceContext) Close() {
	if sc == nil {
		return
	}
	for _, r := range sc.runners {
		if r == nil {
			continue
		}
		if r.cancel != nil {
			r.cancel()
		}
		if r.done != nil {
			<-r.done
		}
		if r.closer != nil {
			_ = r.closer.Close()
		}
	}
	sc.runners = nil
	if sc.DB != nil {
		_ = sc.DB.Close()
	}
}

func appendRunner(in []*syncRunner, r *syncRunner) []*syncRunner {
	if r == nil {
		return in
	}
	return append(in, r)
}

// buildRepositories returns the repository set and the *sql.DB; nil
// when running in-memory.
func buildRepositories(c config.Config) (
	domNotif.Repository,
	domPref.Repository,
	domSub.Repository,
	*sql.DB,
) {
	if c.Postgres.DSN == "" {
		log.Printf("[notification] postgres DSN empty, using in-memory repositories")
		return infraMemory.NewNotificationRepository(),
			infraMemory.NewPreferenceRepository(),
			infraMemory.NewSubscriptionRepository(),
			nil
	}
	db, err := openPostgres(c.Postgres)
	if err != nil {
		log.Printf("[notification] warning: postgres unavailable: %v; falling back to in-memory", err)
		return infraMemory.NewNotificationRepository(),
			infraMemory.NewPreferenceRepository(),
			infraMemory.NewSubscriptionRepository(),
			nil
	}
	if c.Postgres.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := infraPg.EnsureSchema(ctx, db); err != nil {
			log.Printf("[notification] warning: ensure schema: %v", err)
		}
	}
	log.Printf("[notification] postgres repositories wired")
	return infraPg.NewNotificationRepository(db),
		infraPg.NewPreferenceRepository(db),
		infraPg.NewSubscriptionRepository(db),
		db
}

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

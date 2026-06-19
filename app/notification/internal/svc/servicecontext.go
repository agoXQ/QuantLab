// Package svc is the composition root for the Notification Service.
// It wires the application service against the in-memory or Postgres-
// backed repositories chosen by config and exposes the resulting
// Service for the gRPC / HTTP handlers + the user-events consumer.
package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appUserSync "github.com/agoXQ/QuantLab/app/notification/application/usersync"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	infraMemory "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/notification/infrastructure/repository/postgres"
	infraUserSync "github.com/agoXQ/QuantLab/app/notification/infrastructure/usersync"
	"github.com/agoXQ/QuantLab/app/notification/internal/config"
)

// ServiceContext is the composition root.
type ServiceContext struct {
	Config  config.Config
	Service appNotif.Service
	DB      *sql.DB

	userSync *syncRunner
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
		Config:   c,
		Service:  svc,
		DB:       db,
		userSync: buildUserSync(c, svc),
	}
	if sc.userSync != nil {
		sc.userSync.start()
	}
	return sc
}

// Close releases the database handle and shuts down the cross-service
// event consumer. Safe to call multiple times.
func (sc *ServiceContext) Close() {
	if sc == nil {
		return
	}
	if sc.userSync != nil {
		if sc.userSync.cancel != nil {
			sc.userSync.cancel()
		}
		if sc.userSync.done != nil {
			<-sc.userSync.done
		}
		if sc.userSync.closer != nil {
			_ = sc.userSync.closer.Close()
		}
	}
	if sc.DB != nil {
		_ = sc.DB.Close()
	}
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

// syncRunner is the small lifecycle owner returned by buildUserSync;
// mirrors the User Service shape so a future BacktestSync etc. can
// extend the same pattern.
type syncRunner struct {
	start  func()
	cancel context.CancelFunc
	done   chan struct{}
	closer interface{ Close() error }
}

func buildUserSync(c config.Config, svc appNotif.Service) *syncRunner {
	if !c.UserSync.Enabled {
		return nil
	}
	if len(c.UserSync.Brokers) == 0 {
		log.Printf("[notification] user sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appUserSync.NewHandler(svc)
	consumer, err := infraUserSync.NewConsumer(infraUserSync.Config{
		Brokers: c.UserSync.Brokers,
		Topic:   c.UserSync.Topic,
		GroupID: c.UserSync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[notification] user sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[notification] user sync started topic=%s group=%s",
					orDefault(c.UserSync.Topic, infraUserSync.DefaultTopic),
					orDefault(c.UserSync.GroupID, infraUserSync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[notification] user sync stopped: %v", err)
				}
			}()
		},
		cancel: cancel,
		done:   doneCh,
		closer: consumer,
	}
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

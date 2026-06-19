// Package svc is the composition root for the User Service. It wires
// the application service against the in-memory or Postgres-backed
// repositories chosen by config and exposes the resulting Service for
// the HTTP / gRPC handlers to call.
package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
	infraEvent "github.com/agoXQ/QuantLab/app/user/infrastructure/event"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/password"
	infraMemory "github.com/agoXQ/QuantLab/app/user/infrastructure/repository/memory"
	infraPg "github.com/agoXQ/QuantLab/app/user/infrastructure/repository/postgres"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
	"github.com/agoXQ/QuantLab/app/user/internal/config"
)

// ServiceContext is the composition root for the User Service.
type ServiceContext struct {
	Config  config.Config
	UserSvc appUser.Service
	DB      *sql.DB

	tokenIssuer *token.JWTIssuer
}

// NewServiceContext wires the dependency graph.
func NewServiceContext(c config.Config) *ServiceContext {
	users, follows, db := buildRepositories(c)
	hasher := password.NewBcryptHasher(c.Password.BcryptCost)
	issuer := token.NewJWTIssuer(token.Config{
		Secret:     c.Token.Secret,
		Issuer:     c.Token.Issuer,
		AccessTTL:  time.Duration(c.Token.AccessTTLSeconds) * time.Second,
		RefreshTTL: time.Duration(c.Token.RefreshTTLSeconds) * time.Second,
	})
	publisher := buildPublisher(c)

	svc := appUser.NewService(appUser.Dependencies{
		Users:     users,
		Follows:   follows,
		Hasher:    hasher,
		Tokens:    issuer,
		Publisher: publisher,
		Clock:     time.Now,
	})

	return &ServiceContext{
		Config:      c,
		UserSvc:     svc,
		DB:          db,
		tokenIssuer: issuer,
	}
}

// TokenIssuer exposes the issuer so future middleware can verify JWTs
// against the same secret used to mint them.
func (sc *ServiceContext) TokenIssuer() *token.JWTIssuer {
	if sc == nil {
		return nil
	}
	return sc.tokenIssuer
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
// transient outage does not take the service down during local dev.
func buildRepositories(c config.Config) (
	domuser.Repository,
	domfollow.Repository,
	*sql.DB,
) {
	if c.Postgres.DSN == "" {
		log.Printf("[user] postgres DSN empty, using in-memory repositories")
		return infraMemory.NewUserRepository(), infraMemory.NewFollowRepository(), nil
	}
	db, err := openPostgres(c.Postgres)
	if err != nil {
		log.Printf("[user] warning: postgres unavailable: %v; falling back to in-memory", err)
		return infraMemory.NewUserRepository(), infraMemory.NewFollowRepository(), nil
	}
	if c.Postgres.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := infraPg.EnsureSchema(ctx, db); err != nil {
			log.Printf("[user] warning: ensure schema: %v", err)
		}
	}
	log.Printf("[user] postgres repositories wired")
	return infraPg.NewUserRepository(db), infraPg.NewFollowRepository(db), db
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

func buildPublisher(c config.Config) domevent.Publisher {
	if len(c.Kafka.Brokers) == 0 {
		return infraEvent.Noop{}
	}
	return infraEvent.NewKafkaPublisher(c.Kafka.Brokers)
}

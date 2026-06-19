// Package strategy is the application layer for the Strategy Service.
// It composes the Strategy / StrategyVersion / StrategyFork repositories,
// the event publisher, and a small clock dependency into a flat set of
// use cases. The shape mirrors Backtest's application service so the
// rest of the platform can read all services with one mental model.
package strategy

import (
	"context"
	"time"

	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// Service is the application-level interface for the Strategy Service.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*CreateResult, error)
	Update(ctx context.Context, req UpdateRequest) (*domstrategy.Strategy, error)
	Get(ctx context.Context, id int64) (*domstrategy.Strategy, error)
	List(ctx context.Context, q ListQuery) ([]*domstrategy.Strategy, error)
	Delete(ctx context.Context, id int64, callerID int64) error

	CreateVersion(ctx context.Context, req CreateVersionRequest) (*CreateVersionResult, error)
	GetVersion(ctx context.Context, id int64) (*domversion.StrategyVersion, error)
	ListVersions(ctx context.Context, strategyID int64, limit int) ([]*domversion.StrategyVersion, error)

	Publish(ctx context.Context, req PublishRequest) (*domstrategy.Strategy, error)
	Archive(ctx context.Context, req ArchiveRequest) (*domstrategy.Strategy, error)
	Fork(ctx context.Context, req ForkRequest) (*ForkResult, error)

	MarkBacktested(ctx context.Context, req MarkBacktestedRequest) (*domstrategy.Strategy, error)
}

// Dependencies bundles the ports the service needs.
type Dependencies struct {
	Strategies domstrategy.Repository
	Versions   domversion.Repository
	Forks      domfork.Repository
	Publisher  domevent.Publisher
	Clock      func() time.Time
}

type service struct {
	deps Dependencies
}

// NewService builds the default application service. Publisher is
// optional (nil = no events); Clock defaults to time.Now.
func NewService(deps Dependencies) Service {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	return &service{deps: deps}
}

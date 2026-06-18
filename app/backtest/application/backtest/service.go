// Package backtest is the application layer for the Backtest Engine. It
// wires the executor (Formula Engine adapter), the matching engine, the
// portfolio/report repositories, and the event publisher into a small set
// of use cases: create a job, run it synchronously (MVP), and read the
// resulting trades / positions / report.
package backtest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/backtest/domain/event"
	domexec "github.com/agoXQ/QuantLab/app/backtest/domain/executor"
	dommatch "github.com/agoXQ/QuantLab/app/backtest/domain/matching"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	domorder "github.com/agoXQ/QuantLab/app/backtest/domain/order"
	domportfolio "github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"
	domreport "github.com/agoXQ/QuantLab/app/backtest/domain/report"
	domtrade "github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Service is the application-level interface for the Backtest Engine.
type Service interface {
	Create(ctx context.Context, req CreateBacktestRequest) (*CreateBacktestResult, error)
	// Run executes the job synchronously in the calling goroutine. It is
	// kept for the inline ?run=true / ?wait=true paths and the e2e
	// regression harness; production traffic should go through Submit.
	Run(ctx context.Context, jobID int64) (*RunResult, error)
	// Submit transitions the job into QUEUED and hands it to the queue
	// port. It returns immediately; workers pick the job up asynchronously
	// and call RunQueued. Returns ErrQueueUnavailable when no queue is
	// configured (in-memory dev binary without async wiring).
	Submit(ctx context.Context, jobID int64) (*backtestjob.BacktestJob, error)
	// Cancel requests termination of a non-terminal job. Jobs already
	// running are flipped to CANCELLED in the repository; the worker
	// observes that on its next status read and stops.
	Cancel(ctx context.Context, jobID int64, reason string) (*backtestjob.BacktestJob, error)
	// RunQueued is the worker callback. It mirrors Run but does not
	// surface a RunResult (the worker has nowhere to send it) and writes
	// the failure reason back to ErrorMessage on the job row.
	RunQueued(ctx context.Context, jobID int64) error
	Get(ctx context.Context, jobID int64) (*backtestjob.BacktestJob, error)
	List(ctx context.Context, q ListJobsQuery) ([]*backtestjob.BacktestJob, error)
	GetReport(ctx context.Context, jobID int64) (*domreport.PerformanceReport, error)
	GetTrades(ctx context.Context, jobID int64) ([]*domtrade.Trade, error)
	GetSnapshots(ctx context.Context, jobID int64) ([]domportfolio.Snapshot, error)
	// Reconcile recovers from an ungraceful shutdown. See reconcile.go
	// for the precise behaviour.
	Reconcile(ctx context.Context) (ReconcileResult, error)
}

// Dependencies bundles the ports the service needs.
type Dependencies struct {
	Jobs       backtestjob.Repository
	Orders     domorder.Repository
	Trades     domtrade.Repository
	Portfolios domportfolio.Repository
	Reports    domreport.Repository
	Executor   domexec.StrategyExecutor
	Matching   dommatch.Engine
	MarketData dommarket.Provider
	Publisher  domevent.Publisher
	// Queue is the asynchronous transport for queued jobs. It is
	// optional; when nil, Submit returns ErrQueueUnavailable so callers
	// fall back to the synchronous Run path.
	Queue domqueue.Queue
	Clock func() time.Time
}

type service struct {
	deps Dependencies
}

// NewService builds the default application service.
//
// All ports are required except Publisher (events become no-ops when nil)
// and Clock (defaults to time.Now). The constructor does not run the
// dependency check; missing ports surface as nil-deref at the use-case
// boundary, matching the pattern used by Market Data.
func NewService(deps Dependencies) Service {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	return &service{deps: deps}
}

// Create persists a new BacktestJob in the CREATED state.
func (s *service) Create(ctx context.Context, req CreateBacktestRequest) (*CreateBacktestResult, error) {
	universe := normaliseUniverse(req.Universe)
	cfg := req.Config
	cfg.Normalize()

	job := &backtestjob.BacktestJob{
		UserID:         req.UserID,
		StrategyID:     req.StrategyID,
		VersionID:      req.VersionID,
		Name:           strings.TrimSpace(req.Name),
		Formula:        strings.TrimSpace(req.Formula),
		Universe:       universe,
		Benchmark:      strings.TrimSpace(req.Benchmark),
		DataVersion:    strings.TrimSpace(req.DataVersion),
		InitialCapital: req.InitialCapital,
		Range:          req.Range,
		Config:         cfg,
		Status:         valueobject.JobStatusCreated,
		CreatedAt:      s.deps.Clock(),
	}
	if err := job.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Jobs.Create(ctx, job); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventBacktestCreated, job.ID, domevent.BacktestCreatedPayload{
		JobID:      job.ID,
		UserID:     job.UserID,
		StrategyID: job.StrategyID,
		Formula:    job.Formula,
	})
	return &CreateBacktestResult{Job: job}, nil
}

// Get returns the job by ID.
func (s *service) Get(ctx context.Context, jobID int64) (*backtestjob.BacktestJob, error) {
	if jobID <= 0 {
		return nil, bterr.ErrInvalidJob
	}
	return s.deps.Jobs.Get(ctx, jobID)
}

// List returns jobs filtered by the query.
func (s *service) List(ctx context.Context, q ListJobsQuery) ([]*backtestjob.BacktestJob, error) {
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.deps.Jobs.List(ctx, backtestjob.ListQuery{
		UserID:     q.UserID,
		StrategyID: q.StrategyID,
		Status:     q.Status,
		Limit:      limit,
	})
}

// GetReport returns the performance report for a finished job.
func (s *service) GetReport(ctx context.Context, jobID int64) (*domreport.PerformanceReport, error) {
	if jobID <= 0 {
		return nil, bterr.ErrInvalidJob
	}
	return s.deps.Reports.Get(ctx, jobID)
}

// GetTrades returns the trades produced by a job.
func (s *service) GetTrades(ctx context.Context, jobID int64) ([]*domtrade.Trade, error) {
	if jobID <= 0 {
		return nil, bterr.ErrInvalidJob
	}
	return s.deps.Trades.ListByJob(ctx, jobID)
}

// GetSnapshots returns the portfolio snapshots produced by a job.
func (s *service) GetSnapshots(ctx context.Context, jobID int64) ([]domportfolio.Snapshot, error) {
	if jobID <= 0 {
		return nil, bterr.ErrInvalidJob
	}
	return s.deps.Portfolios.ListSnapshots(ctx, jobID)
}

// publish is a forgiving wrapper around the event publisher: missing
// publishers and publish errors degrade to logs because the MVP value chain
// must work even when Kafka is offline.
func (s *service) publish(ctx context.Context, t domevent.EventType, jobID int64, payload any) {
	if s.deps.Publisher == nil {
		return
	}
	_ = s.deps.Publisher.Publish(ctx, domevent.Event{
		EventID:       uuid.NewString(),
		EventType:     t,
		EventVersion:  domevent.EventVersionV1,
		OccurredAt:    s.deps.Clock(),
		AggregateType: domevent.AggregateTypeBacktest,
		AggregateID:   fmt.Sprintf("%d", jobID),
		Producer:      domevent.ProducerBacktest,
		Payload:       payload,
	})
}

// normaliseUniverse trims, deduplicates, and sorts the universe so the
// engine produces deterministic output regardless of caller ordering.
func normaliseUniverse(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, code := range in {
		c := strings.TrimSpace(code)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

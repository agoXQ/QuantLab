// Package memory provides in-process implementations of the Backtest
// repositories. They power local tests, CI smoke runs, and dev mode
// when no Postgres DSN is configured. Postgres adapters slot in alongside
// these without changing the application layer.
package memory

import (
	"context"
	"sort"
	"sync"

	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
)

// JobRepository is an in-memory backtestjob.Repository.
type JobRepository struct {
	mu   sync.RWMutex
	seq  int64
	data map[int64]*backtestjob.BacktestJob
}

// NewJobRepository returns an empty JobRepository.
func NewJobRepository() *JobRepository {
	return &JobRepository{data: make(map[int64]*backtestjob.BacktestJob)}
}

// Create assigns a monotonic ID and stores a deep copy of the job.
func (r *JobRepository) Create(_ context.Context, job *backtestjob.BacktestJob) error {
	if job == nil {
		return bterr.ErrInvalidJob
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	job.ID = r.seq
	cp := *job
	r.data[cp.ID] = &cp
	return nil
}

// Update overwrites the stored copy. Updates are by-value so callers can
// keep mutating their local handle without poisoning the repository.
func (r *JobRepository) Update(_ context.Context, job *backtestjob.BacktestJob) error {
	if job == nil || job.ID == 0 {
		return bterr.ErrInvalidJob
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[job.ID]; !ok {
		return bterr.ErrJobNotFound
	}
	cp := *job
	r.data[job.ID] = &cp
	return nil
}

// Get returns a copy of the stored job or ErrJobNotFound.
func (r *JobRepository) Get(_ context.Context, id int64) (*backtestjob.BacktestJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.data[id]
	if !ok {
		return nil, bterr.ErrJobNotFound
	}
	cp := *job
	return &cp, nil
}

// List returns jobs filtered by the query. Results are newest-first.
func (r *JobRepository) List(_ context.Context, q backtestjob.ListQuery) ([]*backtestjob.BacktestJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*backtestjob.BacktestJob, 0, len(r.data))
	for _, j := range r.data {
		if q.UserID != 0 && j.UserID != q.UserID {
			continue
		}
		if q.StrategyID != 0 && j.StrategyID != q.StrategyID {
			continue
		}
		if q.Status != "" && j.Status != q.Status {
			continue
		}
		cp := *j
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

// OrderRepository is an in-memory order.Repository.
type OrderRepository struct {
	mu     sync.RWMutex
	seq    int64
	byJob  map[int64][]*order.Order
}

// NewOrderRepository returns an empty OrderRepository.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{byJob: make(map[int64][]*order.Order)}
}

// BulkInsert assigns IDs and stores copies grouped by JobID.
func (r *OrderRepository) BulkInsert(_ context.Context, orders []*order.Order) error {
	if len(orders) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range orders {
		if o == nil {
			continue
		}
		r.seq++
		o.ID = r.seq
		cp := *o
		r.byJob[cp.JobID] = append(r.byJob[cp.JobID], &cp)
	}
	return nil
}

// ListByJob returns all orders attached to a job, ordered by submission time.
func (r *OrderRepository) ListByJob(_ context.Context, jobID int64) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.byJob[jobID]
	out := make([]*order.Order, len(src))
	for i, o := range src {
		cp := *o
		out[i] = &cp
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SubmittedAt.Before(out[j].SubmittedAt)
	})
	return out, nil
}

// TradeRepository is an in-memory trade.Repository.
type TradeRepository struct {
	mu    sync.RWMutex
	seq   int64
	byJob map[int64][]*trade.Trade
}

// NewTradeRepository returns an empty TradeRepository.
func NewTradeRepository() *TradeRepository {
	return &TradeRepository{byJob: make(map[int64][]*trade.Trade)}
}

// BulkInsert assigns IDs and stores copies grouped by JobID.
func (r *TradeRepository) BulkInsert(_ context.Context, trades []*trade.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range trades {
		if t == nil {
			continue
		}
		r.seq++
		t.ID = r.seq
		cp := *t
		r.byJob[cp.JobID] = append(r.byJob[cp.JobID], &cp)
	}
	return nil
}

// ListByJob returns trades for the job ordered by trade time then id.
func (r *TradeRepository) ListByJob(_ context.Context, jobID int64) ([]*trade.Trade, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.byJob[jobID]
	out := make([]*trade.Trade, len(src))
	for i, t := range src {
		cp := *t
		out[i] = &cp
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].TradeTime.Equal(out[j].TradeTime) {
			return out[i].TradeTime.Before(out[j].TradeTime)
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// PortfolioRepository is an in-memory portfolio snapshot store.
type PortfolioRepository struct {
	mu    sync.RWMutex
	byJob map[int64][]portfolio.Snapshot
}

// NewPortfolioRepository returns an empty PortfolioRepository.
func NewPortfolioRepository() *PortfolioRepository {
	return &PortfolioRepository{byJob: make(map[int64][]portfolio.Snapshot)}
}

// BulkInsertSnapshots appends snapshots to the per-job slice.
func (r *PortfolioRepository) BulkInsertSnapshots(_ context.Context, snapshots []portfolio.Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range snapshots {
		// Defensive copy of the position slice so callers can keep
		// mutating their local snapshots.
		positions := make([]portfolio.Position, len(s.Positions))
		copy(positions, s.Positions)
		s.Positions = positions
		r.byJob[s.JobID] = append(r.byJob[s.JobID], s)
	}
	return nil
}

// ListSnapshots returns snapshots ordered by trade date.
func (r *PortfolioRepository) ListSnapshots(_ context.Context, jobID int64) ([]portfolio.Snapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.byJob[jobID]
	out := make([]portfolio.Snapshot, len(src))
	copy(out, src)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].TradeDate.Before(out[j].TradeDate)
	})
	return out, nil
}

// ReportRepository is an in-memory report.Repository keyed by JobID.
type ReportRepository struct {
	mu   sync.RWMutex
	data map[int64]*report.PerformanceReport
}

// NewReportRepository returns an empty ReportRepository.
func NewReportRepository() *ReportRepository {
	return &ReportRepository{data: make(map[int64]*report.PerformanceReport)}
}

// Save writes a copy of the report.
func (r *ReportRepository) Save(_ context.Context, rep *report.PerformanceReport) error {
	if rep == nil {
		return bterr.ErrReportNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *rep
	cp.EquityCurve = append([]report.EquityPoint{}, rep.EquityCurve...)
	r.data[rep.JobID] = &cp
	return nil
}

// Get returns the saved report or ErrReportNotFound.
func (r *ReportRepository) Get(_ context.Context, jobID int64) (*report.PerformanceReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rep, ok := r.data[jobID]
	if !ok {
		return nil, bterr.ErrReportNotFound
	}
	cp := *rep
	cp.EquityCurve = append([]report.EquityPoint{}, rep.EquityCurve...)
	return &cp, nil
}


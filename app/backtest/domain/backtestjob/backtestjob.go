// Package backtestjob defines the BacktestJob aggregate.
//
// BacktestJob owns the lifecycle (Created/Queued/Running/.../Completed)
// and the immutable Config that determines deterministic replay. The
// aggregate is intentionally small: heavy state (positions, trades,
// reports) lives in sibling aggregates referenced by JobID, and a
// BacktestJob row never mutates after the job reaches a terminal status.
package backtestjob

import (
	"context"
	"time"

	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// BacktestJob is the aggregate root for one configured backtest run.
type BacktestJob struct {
	ID             int64                  `json:"id"`
	UserID         int64                  `json:"user_id,omitempty"`
	StrategyID     int64                  `json:"strategy_id,omitempty"`
	VersionID      int64                  `json:"version_id,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Formula        string                 `json:"formula"`
	Universe       []string               `json:"universe"`
	Benchmark      string                 `json:"benchmark,omitempty"`
	DataVersion    string                 `json:"data_version,omitempty"`
	InitialCapital float64                `json:"initial_capital"`
	Range          valueobject.DateRange  `json:"range"`
	Config         Config                 `json:"config"`
	Status         valueobject.JobStatus  `json:"status"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	// Progress is a coarse 0.0-1.0 indicator of replay completion. The
	// engine writes it periodically while iterating the calendar so the
	// HTTP /:id/status endpoint can show a progress bar without forcing
	// the runner into a per-bar I/O loop. Values outside [0,1] are
	// clamped on read so a stale row never breaks the projection.
	Progress       float64                `json:"progress"`
	CreatedAt      time.Time              `json:"created_at"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	FinishedAt     *time.Time             `json:"finished_at,omitempty"`
}

// Config bundles the parameters that determine the execution environment.
//
// All fields are read-only after the job is queued; they are bundled with
// the aggregate root rather than promoted to a sibling because they are
// always loaded together and never mutated independently.
type Config struct {
	CommissionRate     float64                       `json:"commission_rate"`
	SlippageRate       float64                       `json:"slippage_rate"`
	StampDutyRate      float64                       `json:"stamp_duty_rate"`
	MinCommission      float64                       `json:"min_commission"`
	MaxPositionCount   int                           `json:"max_position_count"`
	RebalanceFrequency valueobject.RebalanceFrequency `json:"rebalance_frequency"`
	LookbackBars       int                            `json:"lookback_bars"`
}

// DefaultConfig returns a sensible A-share defaults bundle. Callers may
// override any field; missing fields are filled in by Normalize.
func DefaultConfig() Config {
	return Config{
		CommissionRate:     0.0003, // 万三
		SlippageRate:       0.001,  // 0.1%
		StampDutyRate:      0.001,  // 千一，仅卖出
		MinCommission:      5.0,
		MaxPositionCount:   20,
		RebalanceFrequency: valueobject.RebalanceMonthly,
		LookbackBars:       60,
	}
}

// Normalize fills in defaults for any fields left at the zero value. It is
// safe to call repeatedly.
func (c *Config) Normalize() {
	def := DefaultConfig()
	if c.CommissionRate <= 0 {
		c.CommissionRate = def.CommissionRate
	}
	if c.SlippageRate < 0 {
		c.SlippageRate = def.SlippageRate
	}
	if c.StampDutyRate < 0 {
		c.StampDutyRate = def.StampDutyRate
	}
	if c.MinCommission < 0 {
		c.MinCommission = def.MinCommission
	}
	if c.MaxPositionCount <= 0 {
		c.MaxPositionCount = def.MaxPositionCount
	}
	if !c.RebalanceFrequency.IsValid() {
		c.RebalanceFrequency = def.RebalanceFrequency
	}
	if c.LookbackBars <= 0 {
		c.LookbackBars = def.LookbackBars
	}
}

// Validate runs structural and business-rule checks on the job.
func (j *BacktestJob) Validate() error {
	if j == nil {
		return bterr.ErrInvalidJob
	}
	if j.Formula == "" {
		return bterr.ErrInvalidFormula
	}
	if len(j.Universe) == 0 {
		return bterr.ErrInvalidUniverse
	}
	if err := j.Range.Validate(); err != nil {
		return bterr.ErrInvalidDateRange
	}
	if j.InitialCapital <= 0 {
		return bterr.ErrInvalidInitialCapital
	}
	if j.Config.RebalanceFrequency != "" && !j.Config.RebalanceFrequency.IsValid() {
		return bterr.ErrInvalidRebalanceFreq
	}
	return nil
}

// MarkQueued transitions the job into QUEUED. The transition is only valid
// from CREATED (a freshly persisted job) or QUEUED itself (idempotent
// re-submit). Terminal jobs cannot be re-queued; the API layer must
// surface a 409 in that case so callers know to clone the job instead.
func (j *BacktestJob) MarkQueued(now time.Time) error {
	switch j.Status {
	case valueobject.JobStatusCreated, valueobject.JobStatusQueued:
		j.Status = valueobject.JobStatusQueued
		j.ErrorMessage = ""
		// StartedAt / FinishedAt are reset so a re-queued job does not
		// inherit timestamps or progress from a previous attempt.
		j.StartedAt = nil
		j.FinishedAt = nil
		j.Progress = 0
		_ = now
		return nil
	default:
		return bterr.ErrInvalidStateTransition
	}
}

// MarkCancelled transitions the job into CANCELLED. Cancellation is only
// allowed before completion: CREATED, QUEUED, RUNNING. Terminal jobs are
// rejected with ErrJobNotCancellable so the caller can distinguish a
// state-machine refusal from a missing-job 404.
func (j *BacktestJob) MarkCancelled(now time.Time, reason string) error {
	switch j.Status {
	case valueobject.JobStatusCreated, valueobject.JobStatusQueued, valueobject.JobStatusRunning:
		j.Status = valueobject.JobStatusCancelled
		j.ErrorMessage = reason
		t := now
		j.FinishedAt = &t
		return nil
	default:
		return bterr.ErrJobNotCancellable
	}
}

// MarkRunning transitions the job into RUNNING. Returns an error if the
// current status is incompatible with that move.
func (j *BacktestJob) MarkRunning(now time.Time) error {
	switch j.Status {
	case valueobject.JobStatusCreated, valueobject.JobStatusQueued:
		j.Status = valueobject.JobStatusRunning
		t := now
		j.StartedAt = &t
		return nil
	default:
		return bterr.ErrInvalidStateTransition
	}
}

// MarkCompleted transitions the job into COMPLETED.
func (j *BacktestJob) MarkCompleted(now time.Time) error {
	if j.Status != valueobject.JobStatusRunning {
		return bterr.ErrInvalidStateTransition
	}
	j.Status = valueobject.JobStatusCompleted
	t := now
	j.FinishedAt = &t
	return nil
}

// MarkFailed transitions the job into FAILED with the given reason.
func (j *BacktestJob) MarkFailed(now time.Time, reason string) error {
	if j.Status.IsTerminal() {
		return bterr.ErrInvalidStateTransition
	}
	j.Status = valueobject.JobStatusFailed
	j.ErrorMessage = reason
	t := now
	j.FinishedAt = &t
	return nil
}

// Repository persists BacktestJob aggregates.
type Repository interface {
	Create(ctx context.Context, job *BacktestJob) error
	Update(ctx context.Context, job *BacktestJob) error
	Get(ctx context.Context, id int64) (*BacktestJob, error)
	List(ctx context.Context, q ListQuery) ([]*BacktestJob, error)
}

// ListQuery is a coarse filter used by the API list endpoint.
type ListQuery struct {
	UserID     int64
	StrategyID int64
	Status     valueobject.JobStatus
	Limit      int
}

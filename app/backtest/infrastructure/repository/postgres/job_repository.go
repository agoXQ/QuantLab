package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// JobRepository persists BacktestJob aggregates in PostgreSQL.
type JobRepository struct {
	db *sql.DB
}

// NewJobRepository wires the repository against an existing *sql.DB.
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create inserts a new job and assigns the generated ID back onto the
// aggregate so the caller can reference it without an extra round trip.
func (r *JobRepository) Create(ctx context.Context, job *backtestjob.BacktestJob) error {
	if job == nil {
		return bterr.ErrInvalidJob
	}
	cfgJSON, err := json.Marshal(job.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	const stmt = `
		INSERT INTO backtest_job (
			user_id, strategy_id, version_id, name, formula, universe,
			benchmark, data_version, initial_capital, start_date, end_date,
			status, error_message, config, created_at, started_at, finished_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17
		) RETURNING id`
	var id int64
	err = r.db.QueryRowContext(ctx, stmt,
		nullableInt64(job.UserID),
		nullableInt64(job.StrategyID),
		nullableInt64(job.VersionID),
		nullableString(job.Name),
		job.Formula,
		pq.Array(job.Universe),
		nullableString(job.Benchmark),
		nullableString(job.DataVersion),
		job.InitialCapital,
		job.Range.Start,
		job.Range.End,
		string(job.Status),
		nullableString(job.ErrorMessage),
		cfgJSON,
		job.CreatedAt,
		job.StartedAt,
		job.FinishedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert backtest_job: %w", err)
	}
	job.ID = id
	return nil
}

// Update writes back the mutable fields of a job. The persistent shape is
// append-only for trade / snapshot / report data, so Update only touches the
// row's lifecycle columns.
func (r *JobRepository) Update(ctx context.Context, job *backtestjob.BacktestJob) error {
	if job == nil || job.ID == 0 {
		return bterr.ErrInvalidJob
	}
	cfgJSON, err := json.Marshal(job.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	const stmt = `
		UPDATE backtest_job SET
			status        = $2,
			error_message = $3,
			config        = $4,
			started_at    = $5,
			finished_at   = $6
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, stmt,
		job.ID,
		string(job.Status),
		nullableString(job.ErrorMessage),
		cfgJSON,
		job.StartedAt,
		job.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("update backtest_job: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return bterr.ErrJobNotFound
	}
	return nil
}

// Get loads a single job. Returns ErrJobNotFound when the row is absent.
func (r *JobRepository) Get(ctx context.Context, id int64) (*backtestjob.BacktestJob, error) {
	const q = `
		SELECT id, COALESCE(user_id, 0), COALESCE(strategy_id, 0), COALESCE(version_id, 0),
		       COALESCE(name, ''), formula, universe,
		       COALESCE(benchmark, ''), COALESCE(data_version, ''),
		       initial_capital, start_date, end_date,
		       status, COALESCE(error_message, ''), config,
		       created_at, started_at, finished_at
		FROM backtest_job WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, id)
	job, err := scanJob(row)
	if err == sql.ErrNoRows {
		return nil, bterr.ErrJobNotFound
	}
	return job, err
}

// List returns jobs filtered by the supplied query, newest first.
func (r *JobRepository) List(ctx context.Context, q backtestjob.ListQuery) ([]*backtestjob.BacktestJob, error) {
	conditions := []string{"1 = 1"}
	args := []any{}
	if q.UserID != 0 {
		args = append(args, q.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if q.StrategyID != 0 {
		args = append(args, q.StrategyID)
		conditions = append(conditions, fmt.Sprintf("strategy_id = $%d", len(args)))
	}
	if q.Status != "" {
		args = append(args, string(q.Status))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, COALESCE(user_id, 0), COALESCE(strategy_id, 0), COALESCE(version_id, 0),
		       COALESCE(name, ''), formula, universe,
		       COALESCE(benchmark, ''), COALESCE(data_version, ''),
		       initial_capital, start_date, end_date,
		       status, COALESCE(error_message, ''), config,
		       created_at, started_at, finished_at
		FROM backtest_job
		WHERE %s
		ORDER BY id DESC
		LIMIT $%d`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query backtest_job: %w", err)
	}
	defer rows.Close()
	out := make([]*backtestjob.BacktestJob, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, job)
	}
	return out, rows.Err()
}

// scannable lets scanJob accept both *sql.Row and *sql.Rows.
type scannable interface {
	Scan(dest ...any) error
}

func scanJob(row scannable) (*backtestjob.BacktestJob, error) {
	var (
		job        backtestjob.BacktestJob
		universe   pq.StringArray
		statusStr  string
		startedAt  sql.NullTime
		finishedAt sql.NullTime
		cfgJSON    []byte
	)
	err := row.Scan(
		&job.ID,
		&job.UserID, &job.StrategyID, &job.VersionID,
		&job.Name, &job.Formula, &universe,
		&job.Benchmark, &job.DataVersion,
		&job.InitialCapital, &job.Range.Start, &job.Range.End,
		&statusStr, &job.ErrorMessage, &cfgJSON,
		&job.CreatedAt, &startedAt, &finishedAt,
	)
	if err != nil {
		return nil, err
	}
	job.Universe = []string(universe)
	job.Status = valueobject.JobStatus(statusStr)
	if startedAt.Valid {
		t := startedAt.Time
		job.StartedAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		job.FinishedAt = &t
	}
	if len(cfgJSON) > 0 {
		if err := json.Unmarshal(cfgJSON, &job.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}
	}
	return &job, nil
}

func nullableInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

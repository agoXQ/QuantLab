package postgres

import (
	"context"
	"database/sql"
	"fmt"

	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
)

// ForkRepository persists StrategyFork aggregates.
type ForkRepository struct {
	db *sql.DB
}

func NewForkRepository(db *sql.DB) *ForkRepository {
	return &ForkRepository{db: db}
}

func (r *ForkRepository) Create(ctx context.Context, f *domfork.StrategyFork) error {
	const stmt = `
		INSERT INTO strategy_fork (source_strategy_id, target_strategy_id, creator_id, created_at)
		VALUES ($1, $2, NULLIF($3, 0), $4) RETURNING id`
	return r.db.QueryRowContext(ctx, stmt,
		f.SourceStrategyID, f.TargetStrategyID, f.CreatorID, f.CreatedAt,
	).Scan(&f.ID)
}

func (r *ForkRepository) ListBySource(ctx context.Context, sourceID int64, limit int) ([]*domfork.StrategyFork, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, source_strategy_id, target_strategy_id, COALESCE(creator_id, 0), created_at
		FROM strategy_fork WHERE source_strategy_id = $1
		ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, sourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("query strategy_fork: %w", err)
	}
	defer rows.Close()
	out := make([]*domfork.StrategyFork, 0)
	for rows.Next() {
		var f domfork.StrategyFork
		if err := rows.Scan(&f.ID, &f.SourceStrategyID, &f.TargetStrategyID, &f.CreatorID, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

// silence sql.ErrNoRows usage path; we keep the symbol referenced so a
// future per-row Get does not need a fresh import line.
var _ = sql.ErrNoRows

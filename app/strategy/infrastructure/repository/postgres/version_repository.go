package postgres

import (
	"context"
	"database/sql"
	"fmt"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// VersionRepository persists StrategyVersion aggregates.
type VersionRepository struct {
	db *sql.DB
}

func NewVersionRepository(db *sql.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) Create(ctx context.Context, v *domversion.StrategyVersion) error {
	if v == nil {
		return stratErr.ErrInvalidVersion
	}
	const stmt = `
		INSERT INTO strategy_version (
			strategy_id, version_no, formula_text,
			buy_rule, sell_rule, risk_rule,
			position_rule, rebalance_rule, change_log,
			created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, 0), $11
		) RETURNING id`
	return r.db.QueryRowContext(ctx, stmt,
		v.StrategyID, v.VersionNo, v.FormulaText,
		nullableString(v.BuyRule), nullableString(v.SellRule), nullableString(v.RiskRule),
		nullableString(v.PositionRule), nullableString(v.RebalanceRule), nullableString(v.ChangeLog),
		v.CreatedBy, v.CreatedAt,
	).Scan(&v.ID)
}

func (r *VersionRepository) Get(ctx context.Context, id int64) (*domversion.StrategyVersion, error) {
	const q = `
		SELECT id, strategy_id, version_no, formula_text,
		       COALESCE(buy_rule, ''), COALESCE(sell_rule, ''), COALESCE(risk_rule, ''),
		       COALESCE(position_rule, ''), COALESCE(rebalance_rule, ''), COALESCE(change_log, ''),
		       COALESCE(created_by, 0), created_at
		FROM strategy_version WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, id)
	v, err := scanVersion(row)
	if err == sql.ErrNoRows {
		return nil, stratErr.ErrVersionNotFound
	}
	return v, err
}

func (r *VersionRepository) ListByStrategy(ctx context.Context, strategyID int64, limit int) ([]*domversion.StrategyVersion, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, strategy_id, version_no, formula_text,
		       COALESCE(buy_rule, ''), COALESCE(sell_rule, ''), COALESCE(risk_rule, ''),
		       COALESCE(position_rule, ''), COALESCE(rebalance_rule, ''), COALESCE(change_log, ''),
		       COALESCE(created_by, 0), created_at
		FROM strategy_version WHERE strategy_id = $1 ORDER BY id DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, strategyID, limit)
	if err != nil {
		return nil, fmt.Errorf("query strategy_version: %w", err)
	}
	defer rows.Close()
	out := make([]*domversion.StrategyVersion, 0)
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VersionRepository) LatestNumber(ctx context.Context, strategyID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM strategy_version WHERE strategy_id = $1`
	var count int
	if err := r.db.QueryRowContext(ctx, q, strategyID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count strategy_version: %w", err)
	}
	return count, nil
}

func scanVersion(row scannable) (*domversion.StrategyVersion, error) {
	var v domversion.StrategyVersion
	if err := row.Scan(
		&v.ID, &v.StrategyID, &v.VersionNo, &v.FormulaText,
		&v.BuyRule, &v.SellRule, &v.RiskRule,
		&v.PositionRule, &v.RebalanceRule, &v.ChangeLog,
		&v.CreatedBy, &v.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &v, nil
}

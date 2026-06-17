package compilelog

import (
	"context"
	"database/sql"
	"fmt"

	domainLog "github.com/agoXQ/QuantLab/app/formula/domain/compilelog"
)

const (
	tableName = "formula_compile_log"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL-backed compile log repository.
func NewPostgresRepository(db *sql.DB) domainLog.Repository {
	return &postgresRepository{db: db}
}

// EnsureTable ensures the compile log table exists.
func (r *postgresRepository) EnsureTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			formula_hash VARCHAR(64) NOT NULL,
			formula TEXT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT FALSE,
			error_code INT NOT NULL DEFAULT 0,
			compile_time_ms INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create compile_log table: %w", err)
	}

	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_hash ON %s (formula_hash)`,
		tableName, tableName)
	_, _ = r.db.ExecContext(ctx, indexQuery)

	return nil
}

func (r *postgresRepository) Save(ctx context.Context, record *domainLog.CompileLogRecord) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (formula_hash, formula, success, error_code, compile_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`, tableName)

	err := r.db.QueryRowContext(ctx, query,
		record.FormulaHash,
		record.Formula,
		record.Success,
		record.ErrorCode,
		record.CompileTimeMs,
		record.CreatedAt,
	).Scan(&record.ID)

	if err != nil {
		return fmt.Errorf("save compile log: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListByHash(ctx context.Context, formulaHash string, limit, offset int) ([]domainLog.CompileLogRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, formula_hash, formula, success, error_code, compile_time_ms, created_at
		FROM %s
		WHERE formula_hash = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, tableName)

	rows, err := r.db.QueryContext(ctx, query, formulaHash, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list compile logs: %w", err)
	}
	defer rows.Close()

	var records []domainLog.CompileLogRecord
	for rows.Next() {
		var rec domainLog.CompileLogRecord
		if err := rows.Scan(
			&rec.ID, &rec.FormulaHash, &rec.Formula,
			&rec.Success, &rec.ErrorCode, &rec.CompileTimeMs, &rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan compile log: %w", err)
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate compile logs: %w", err)
	}

	return records, nil
}

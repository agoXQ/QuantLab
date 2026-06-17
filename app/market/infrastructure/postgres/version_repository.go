package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
)

// DataVersionRepository persists DataVersion records.
type DataVersionRepository struct{ db *sql.DB }

// NewDataVersionRepository creates a new DataVersionRepository.
func NewDataVersionRepository(db *sql.DB) *DataVersionRepository {
	return &DataVersionRepository{db: db}
}

// Get implements dataversion.Repository.
func (r *DataVersionRepository) Get(ctx context.Context, version string) (*dataversion.DataVersion, error) {
	const q = `SELECT version, COALESCE(description, ''), created_at FROM data_version WHERE version = $1 LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, version)
	dv, err := scanVersion(row)
	if err == sql.ErrNoRows {
		return nil, domainErr.ErrDataVersionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query data_version: %w", err)
	}
	return dv, nil
}

// List implements dataversion.Repository.
func (r *DataVersionRepository) List(ctx context.Context, limit int) ([]*dataversion.DataVersion, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `SELECT version, COALESCE(description, ''), created_at FROM data_version ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("query data_version: %w", err)
	}
	defer rows.Close()

	out := make([]*dataversion.DataVersion, 0, limit)
	for rows.Next() {
		dv, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, dv)
	}
	return out, rows.Err()
}

// Latest implements dataversion.Repository.
func (r *DataVersionRepository) Latest(ctx context.Context) (*dataversion.DataVersion, error) {
	const q = `SELECT version, COALESCE(description, ''), created_at FROM data_version ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, q)
	dv, err := scanVersion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest data_version: %w", err)
	}
	return dv, nil
}

// Create implements dataversion.Repository.
func (r *DataVersionRepository) Create(ctx context.Context, dv *dataversion.DataVersion) error {
	const stmt = `INSERT INTO data_version (version, description, created_at) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, stmt, dv.Version, dv.Description, dv.CreatedAt); err != nil {
		return fmt.Errorf("insert data_version %s: %w", dv.Version, err)
	}
	return nil
}

func scanVersion(row rowScanner) (*dataversion.DataVersion, error) {
	var dv dataversion.DataVersion
	if err := row.Scan(&dv.Version, &dv.Description, &dv.CreatedAt); err != nil {
		return nil, err
	}
	return &dv, nil
}

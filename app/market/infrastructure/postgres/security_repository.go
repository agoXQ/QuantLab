package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// SecurityRepository persists Security aggregates in PostgreSQL.
type SecurityRepository struct {
	db *sql.DB
}

// NewSecurityRepository creates a new SecurityRepository.
func NewSecurityRepository(db *sql.DB) *SecurityRepository {
	return &SecurityRepository{db: db}
}

// GetByCode implements security.Repository.
func (r *SecurityRepository) GetByCode(ctx context.Context, stockCode string) (*security.Security, error) {
	const q = `
		SELECT id, stock_code, stock_name, market, exchange, asset_type,
		       industry, listing_date, delisting_date, status
		FROM security
		WHERE stock_code = $1
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, strings.ToUpper(strings.TrimSpace(stockCode)))
	sec, err := scanSecurity(row)
	if err == sql.ErrNoRows {
		return nil, domainErr.ErrSecurityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query security: %w", err)
	}
	return sec, nil
}

// List implements security.Repository.
func (r *SecurityRepository) List(ctx context.Context, q security.ListQuery) ([]*security.Security, string, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	cursorID, err := decodeCursor(q.Cursor)
	if err != nil {
		return nil, "", fmt.Errorf("decode cursor: %w", err)
	}

	args := []any{cursorID}
	conditions := []string{"id > $1"}
	if q.Market != "" {
		args = append(args, string(q.Market))
		conditions = append(conditions, fmt.Sprintf("market = $%d", len(args)))
	}
	if q.Exchange != "" {
		args = append(args, strings.ToUpper(q.Exchange))
		conditions = append(conditions, fmt.Sprintf("exchange = $%d", len(args)))
	}
	if q.AssetType != "" {
		args = append(args, string(q.AssetType))
		conditions = append(conditions, fmt.Sprintf("asset_type = $%d", len(args)))
	}
	args = append(args, limit+1)

	query := fmt.Sprintf(`
		SELECT id, stock_code, stock_name, market, exchange, asset_type,
		       industry, listing_date, delisting_date, status
		FROM security
		WHERE %s
		ORDER BY id ASC
		LIMIT $%d`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list securities: %w", err)
	}
	defer rows.Close()

	out := make([]*security.Security, 0, limit)
	for rows.Next() {
		sec, scanErr := scanSecurity(rows)
		if scanErr != nil {
			return nil, "", fmt.Errorf("scan security: %w", scanErr)
		}
		out = append(out, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(out) > limit {
		next = encodeCursor(out[limit-1].ID)
		out = out[:limit]
	}
	return out, next, nil
}

// Upsert implements security.Repository.
func (r *SecurityRepository) Upsert(ctx context.Context, sec *security.Security) error {
	return r.BulkUpsert(ctx, []*security.Security{sec})
}

// BulkUpsert implements security.Repository.
func (r *SecurityRepository) BulkUpsert(ctx context.Context, list []*security.Security) error {
	if len(list) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO security (
			stock_code, stock_name, market, exchange, asset_type,
			industry, listing_date, delisting_date, status, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (stock_code) DO UPDATE SET
			stock_name     = EXCLUDED.stock_name,
			market         = EXCLUDED.market,
			exchange       = EXCLUDED.exchange,
			asset_type     = EXCLUDED.asset_type,
			industry       = EXCLUDED.industry,
			listing_date   = EXCLUDED.listing_date,
			delisting_date = EXCLUDED.delisting_date,
			status         = EXCLUDED.status,
			updated_at     = NOW()`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, sec := range list {
		sec.Normalize()
		_, err := tx.ExecContext(ctx, stmt,
			sec.StockCode, sec.StockName, string(sec.Market), sec.Exchange,
			string(sec.AssetType), sec.Industry,
			nullDate(sec.ListingDate), nullDate(sec.DelistingDate), string(sec.Status),
		)
		if err != nil {
			return fmt.Errorf("upsert security %s: %w", sec.StockCode, err)
		}
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSecurity(row rowScanner) (*security.Security, error) {
	var (
		s            security.Security
		market       sql.NullString
		exchange     sql.NullString
		assetType    sql.NullString
		industry     sql.NullString
		status       sql.NullString
		listing      sql.NullTime
		delisting    sql.NullTime
		stockName    sql.NullString
	)
	if err := row.Scan(
		&s.ID, &s.StockCode, &stockName, &market, &exchange, &assetType,
		&industry, &listing, &delisting, &status,
	); err != nil {
		return nil, err
	}
	s.StockName = stockName.String
	s.Market = valueobject.Market(market.String)
	s.Exchange = exchange.String
	s.AssetType = valueobject.AssetType(assetType.String)
	s.Industry = industry.String
	s.Status = valueobject.SecurityStatus(status.String)
	if listing.Valid {
		s.ListingDate = listing.Time
	}
	if delisting.Valid {
		s.DelistingDate = delisting.Time
	}
	return &s, nil
}

func nullDate(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FinancialRepository persists FinancialStatement entities.
type FinancialRepository struct {
	db *sql.DB
}

// NewFinancialRepository creates a new FinancialRepository.
func NewFinancialRepository(db *sql.DB) *FinancialRepository { return &FinancialRepository{db: db} }

// List implements financial.Repository.
func (r *FinancialRepository) List(ctx context.Context, q financial.ListQuery) ([]*financial.FinancialStatement, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 40
	}
	args := []any{strings.ToUpper(q.StockCode)}
	conditions := []string{"stock_code = $1"}
	if q.ReportType != "" {
		args = append(args, string(q.ReportType))
		conditions = append(conditions, fmt.Sprintf("report_type = $%d", len(args)))
	}
	if !q.Range.Start.IsZero() {
		args = append(args, q.Range.Start)
		conditions = append(conditions, fmt.Sprintf("report_date >= $%d", len(args)))
	}
	if !q.Range.End.IsZero() {
		args = append(args, q.Range.End)
		conditions = append(conditions, fmt.Sprintf("report_date <= $%d", len(args)))
	}
	if q.DataVersion != "" {
		args = append(args, q.DataVersion)
		conditions = append(conditions, fmt.Sprintf("data_version = $%d", len(args)))
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT stock_code, report_date, report_type,
		       COALESCE(revenue, 0), COALESCE(net_profit, 0),
		       COALESCE(total_assets, 0), COALESCE(total_liabilities, 0),
		       COALESCE(net_assets, 0),
		       COALESCE(basic_eps, 0), COALESCE(diluted_eps, 0),
		       COALESCE(operating_cash_flow, 0),
		       COALESCE(investing_cash_flow, 0), COALESCE(financing_cash_flow, 0),
		       COALESCE(data_version, '')
		FROM financial_statement
		WHERE %s
		ORDER BY report_date DESC
		LIMIT $%d`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query financial_statement: %w", err)
	}
	defer rows.Close()

	out := make([]*financial.FinancialStatement, 0, limit)
	for rows.Next() {
		var (
			f          financial.FinancialStatement
			reportType string
		)
		if err := rows.Scan(
			&f.StockCode, &f.ReportDate, &reportType,
			&f.Revenue, &f.NetProfit, &f.TotalAssets, &f.TotalLiabilities,
			&f.NetAssets, &f.BasicEPS, &f.DilutedEPS,
			&f.OperatingCashFlow, &f.InvestingCashFlow,
			&f.FinancingCashFlow, &f.DataVersion,
		); err != nil {
			return nil, fmt.Errorf("scan financial_statement: %w", err)
		}
		f.ReportType = valueobject.ReportType(reportType)
		out = append(out, &f)
	}
	return out, rows.Err()
}

// BulkUpsert implements financial.Repository.
func (r *FinancialRepository) BulkUpsert(ctx context.Context, list []*financial.FinancialStatement) error {
	if len(list) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO financial_statement (
			stock_code, report_date, report_type,
			revenue, net_profit, total_assets, total_liabilities, net_assets,
			basic_eps, diluted_eps,
			operating_cash_flow, investing_cash_flow, financing_cash_flow, data_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (stock_code, report_date, report_type, data_version) DO UPDATE SET
			revenue             = EXCLUDED.revenue,
			net_profit          = EXCLUDED.net_profit,
			total_assets        = EXCLUDED.total_assets,
			total_liabilities   = EXCLUDED.total_liabilities,
			net_assets          = EXCLUDED.net_assets,
			basic_eps           = EXCLUDED.basic_eps,
			diluted_eps         = EXCLUDED.diluted_eps,
			operating_cash_flow = EXCLUDED.operating_cash_flow,
			investing_cash_flow = EXCLUDED.investing_cash_flow,
			financing_cash_flow = EXCLUDED.financing_cash_flow`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, f := range list {
		_, err := tx.ExecContext(ctx, stmt,
			strings.ToUpper(f.StockCode), f.ReportDate, string(f.ReportType),
			f.Revenue, f.NetProfit, f.TotalAssets, f.TotalLiabilities, f.NetAssets,
			f.BasicEPS, f.DilutedEPS,
			f.OperatingCashFlow, f.InvestingCashFlow, f.FinancingCashFlow, f.DataVersion,
		)
		if err != nil {
			return fmt.Errorf("upsert financial_statement %s %s: %w", f.StockCode, f.ReportDate, err)
		}
	}
	return tx.Commit()
}

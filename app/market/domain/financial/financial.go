// Package financial defines the FinancialStatement aggregate.
package financial

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FinancialStatement represents a single financial report snapshot.
type FinancialStatement struct {
	StockCode        string                 `json:"stock_code"`
	ReportDate       time.Time              `json:"report_date"`
	ReportType       valueobject.ReportType `json:"report_type"`
	Revenue          float64                `json:"revenue"`
	NetProfit        float64                `json:"net_profit"`
	TotalAssets      float64                `json:"total_assets"`
	TotalLiabilities float64                `json:"total_liabilities"`
	NetAssets        float64                `json:"net_assets"`
	OperatingCashFlow float64               `json:"operating_cash_flow"`
	InvestingCashFlow float64               `json:"investing_cash_flow"`
	FinancingCashFlow float64               `json:"financing_cash_flow"`
	DataVersion      string                 `json:"data_version"`
}

// Repository persists FinancialStatement entities.
type Repository interface {
	List(ctx context.Context, q ListQuery) ([]*FinancialStatement, error)
	BulkUpsert(ctx context.Context, list []*FinancialStatement) error
}

// ListQuery describes a query against financial statement storage.
type ListQuery struct {
	StockCode   string
	ReportType  valueobject.ReportType
	Range       valueobject.DateRange
	DataVersion string
	Cursor      string
	Limit       int
}

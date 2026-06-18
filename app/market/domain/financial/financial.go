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
	// BasicEPS is the basic earnings per share reported by the issuer (元/股).
	// When the upstream report does not include it, BasicEPS is zero and
	// downstream consumers should fall back to DilutedEPS or the
	// NetProfit / TotalShare derivation.
	BasicEPS         float64                `json:"basic_eps"`
	// DilutedEPS is the diluted earnings per share reported by the issuer.
	DilutedEPS       float64                `json:"diluted_eps"`
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

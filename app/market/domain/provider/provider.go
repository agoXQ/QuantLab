// Package provider defines the upstream data provider abstraction.
//
// Market Data Service is the only owner of third-party connections; downstream
// services consume normalized aggregates through the application layer. The
// DataProvider interface is intentionally narrow so multiple sources (Tushare,
// AkShare, Wind, ...) can be plugged in without leaking provider semantics
// into the domain.
package provider

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// SecurityFetcher fetches security master data.
type SecurityFetcher interface {
	FetchSecurities(ctx context.Context, market valueobject.Market) ([]*security.Security, error)
}

// CalendarFetcher fetches the trading calendar.
type CalendarFetcher interface {
	FetchCalendar(ctx context.Context, market valueobject.Market, r valueobject.DateRange) ([]calendar.TradingDay, error)
}

// BarFetcher fetches K-line bars for a security.
type BarFetcher interface {
	FetchBars(ctx context.Context, q BarQuery) ([]*marketbar.MarketBar, error)
}

// FinancialFetcher fetches financial statements.
type FinancialFetcher interface {
	FetchFinancials(ctx context.Context, q FinancialQuery) ([]*financial.FinancialStatement, error)
}

// FactorFetcher fetches factor values.
type FactorFetcher interface {
	FetchFactors(ctx context.Context, q FactorQuery) ([]*factor.Factor, error)
}

// IndexFetcher fetches index bars.
type IndexFetcher interface {
	FetchIndexBars(ctx context.Context, q IndexQuery) ([]*indexbar.IndexBar, error)
}

// CorporateActionFetcher fetches raw corporate actions used for adjustment.
type CorporateActionFetcher interface {
	FetchCorporateActions(ctx context.Context, stockCode string, r valueobject.DateRange) ([]adjustment.CorporateAction, error)
}

// DataProvider aggregates all per-domain fetcher interfaces. Concrete
// implementations may compose multiple upstream services to satisfy the full
// surface, or embed unimplemented stubs for unsupported queries (returning
// errors.ErrProviderUnavailable).
type DataProvider interface {
	Name() string
	SecurityFetcher
	CalendarFetcher
	BarFetcher
	FinancialFetcher
	FactorFetcher
	IndexFetcher
	CorporateActionFetcher
}

// BarQuery describes a request for K-line bars.
type BarQuery struct {
	StockCode string
	Period    valueobject.Period
	Range     valueobject.DateRange
}

// FinancialQuery describes a request for financial statements.
type FinancialQuery struct {
	StockCode  string
	ReportType valueobject.ReportType
	Range      valueobject.DateRange
}

// FactorQuery describes a request for factor values.
type FactorQuery struct {
	StockCode   string
	FactorNames []string
	Range       valueobject.DateRange
}

// IndexQuery describes a request for index bars.
type IndexQuery struct {
	IndexCode string
	Range     valueobject.DateRange
}

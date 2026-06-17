// Package market is the application layer of the Market Data service.
package market

import (
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// ListSecuritiesQuery describes the application input for listing securities.
type ListSecuritiesQuery struct {
	Market    valueobject.Market
	Exchange  string
	AssetType valueobject.AssetType
	Cursor    string
	Limit     int
}

// SecurityList is the paginated result of ListSecurities.
type SecurityList struct {
	Items      []*security.Security
	NextCursor string
	HasMore    bool
}

// GetBarsQuery describes the input for fetching K-line bars.
//
// All fields are optional except StockCode. Period defaults to day, Adjustment
// defaults to pre.
type GetBarsQuery struct {
	StockCode   string
	Period      valueobject.Period
	Adjustment  valueobject.Adjustment
	Range       valueobject.DateRange
	DataVersion string
	Limit       int
}

// BarList is the result of GetBars.
type BarList struct {
	Items       []*marketbar.MarketBar
	Adjustment  valueobject.Adjustment
	DataVersion string
}

// GetFinancialsQuery describes the input for fetching financial statements.
type GetFinancialsQuery struct {
	StockCode   string
	ReportType  valueobject.ReportType
	Range       valueobject.DateRange
	DataVersion string
	Cursor      string
	Limit       int
}

// FinancialList is the result of GetFinancials.
type FinancialList struct {
	Items       []*financial.FinancialStatement
	NextCursor  string
	HasMore     bool
	DataVersion string
}

// GetFactorsQuery describes the input for fetching factor values.
type GetFactorsQuery struct {
	StockCode   string
	FactorNames []string
	Range       valueobject.DateRange
	DataVersion string
	Limit       int
}

// FactorList is the result of GetFactors.
type FactorList struct {
	Items       []*factor.Factor
	DataVersion string
}

// GetIndexQuery describes the input for fetching index bars.
type GetIndexQuery struct {
	IndexCode   string
	Range       valueobject.DateRange
	DataVersion string
}

// IndexList is the result of GetIndex.
type IndexList struct {
	Items       []*indexbar.IndexBar
	DataVersion string
}

// CalendarQuery describes the input for fetching the trading calendar.
type CalendarQuery struct {
	Range valueobject.DateRange
}

// CalendarResult is the result of GetCalendar.
type CalendarResult struct {
	Days []calendar.TradingDay
}

// VersionsResult is the result of ListVersions.
type VersionsResult struct {
	Items []*dataversion.DataVersion
}

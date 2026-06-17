// Package faketushare offers an in-memory implementation of the domain
// DataProvider, primarily used for tests, demos, and offline development when
// no Tushare token is available.
package faketushare

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	domainProvider "github.com/agoXQ/QuantLab/app/market/domain/provider"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// Provider is an in-memory DataProvider backed by static fixture data.
type Provider struct {
	Securities       []*security.Security
	Calendar         []calendar.TradingDay
	Bars             []*marketbar.MarketBar
	Financials       []*financial.FinancialStatement
	Factors          []*factor.Factor
	IndexBars        []*indexbar.IndexBar
	CorporateActions []adjustment.CorporateAction
}

// NewProvider returns an empty Provider; populate the fields directly from
// tests.
func NewProvider() *Provider { return &Provider{} }

// Ensure interface compliance.
var _ domainProvider.DataProvider = (*Provider)(nil)

// Name implements domain.DataProvider.
func (p *Provider) Name() string { return "faketushare" }

// FetchSecurities implements provider.SecurityFetcher.
func (p *Provider) FetchSecurities(ctx context.Context, _ valueobject.Market) ([]*security.Security, error) {
	out := make([]*security.Security, len(p.Securities))
	for i, s := range p.Securities {
		cp := *s
		out[i] = &cp
	}
	return out, nil
}

// FetchCalendar implements provider.CalendarFetcher.
func (p *Provider) FetchCalendar(ctx context.Context, _ valueobject.Market, r valueobject.DateRange) ([]calendar.TradingDay, error) {
	out := make([]calendar.TradingDay, 0, len(p.Calendar))
	for _, d := range p.Calendar {
		if !r.Start.IsZero() && d.TradeDate.Before(r.Start) {
			continue
		}
		if !r.End.IsZero() && d.TradeDate.After(r.End) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

// FetchBars implements provider.BarFetcher.
func (p *Provider) FetchBars(ctx context.Context, q domainProvider.BarQuery) ([]*marketbar.MarketBar, error) {
	out := make([]*marketbar.MarketBar, 0, len(p.Bars))
	for _, b := range p.Bars {
		if b.StockCode != q.StockCode {
			continue
		}
		if q.Period != "" && b.Period != q.Period {
			continue
		}
		if !q.Range.Start.IsZero() && b.TradeDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && b.TradeDate.After(q.Range.End) {
			continue
		}
		cp := *b
		out = append(out, &cp)
	}
	return out, nil
}

// FetchFinancials implements provider.FinancialFetcher.
func (p *Provider) FetchFinancials(ctx context.Context, q domainProvider.FinancialQuery) ([]*financial.FinancialStatement, error) {
	out := make([]*financial.FinancialStatement, 0, len(p.Financials))
	for _, f := range p.Financials {
		if f.StockCode != q.StockCode {
			continue
		}
		if q.ReportType != "" && f.ReportType != q.ReportType {
			continue
		}
		if !q.Range.Start.IsZero() && f.ReportDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && f.ReportDate.After(q.Range.End) {
			continue
		}
		cp := *f
		out = append(out, &cp)
	}
	return out, nil
}

// FetchFactors implements provider.FactorFetcher.
func (p *Provider) FetchFactors(ctx context.Context, q domainProvider.FactorQuery) ([]*factor.Factor, error) {
	wanted := make(map[string]struct{}, len(q.FactorNames))
	for _, n := range q.FactorNames {
		wanted[n] = struct{}{}
	}
	out := make([]*factor.Factor, 0, len(p.Factors))
	for _, f := range p.Factors {
		if f.StockCode != q.StockCode {
			continue
		}
		if len(wanted) > 0 {
			if _, ok := wanted[f.FactorName]; !ok {
				continue
			}
		}
		if !q.Range.Start.IsZero() && f.TradeDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && f.TradeDate.After(q.Range.End) {
			continue
		}
		cp := *f
		out = append(out, &cp)
	}
	return out, nil
}

// FetchIndexBars implements provider.IndexFetcher.
func (p *Provider) FetchIndexBars(ctx context.Context, q domainProvider.IndexQuery) ([]*indexbar.IndexBar, error) {
	out := make([]*indexbar.IndexBar, 0, len(p.IndexBars))
	for _, b := range p.IndexBars {
		if b.IndexCode != q.IndexCode {
			continue
		}
		if !q.Range.Start.IsZero() && b.TradeDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && b.TradeDate.After(q.Range.End) {
			continue
		}
		cp := *b
		out = append(out, &cp)
	}
	return out, nil
}

// FetchCorporateActions implements provider.CorporateActionFetcher.
func (p *Provider) FetchCorporateActions(ctx context.Context, stockCode string, _ valueobject.DateRange) ([]adjustment.CorporateAction, error) {
	out := make([]adjustment.CorporateAction, 0, len(p.CorporateActions))
	for _, a := range p.CorporateActions {
		if a.StockCode != stockCode {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

package tushare

import (
	"context"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FetchSecurities returns the security master for a given market.
//
// Currently only the CN A-share market is wired up via Tushare's stock_basic
// endpoint. Other markets return an empty slice; callers that need them
// should compose this provider with another implementation.
func (p *Provider) FetchSecurities(ctx context.Context, market valueobject.Market) ([]*security.Security, error) {
	if market == "" {
		market = valueobject.MarketCN
	}
	if market != valueobject.MarketCN {
		return nil, nil
	}
	resp, err := p.client.Call(ctx, "stock_basic",
		map[string]any{"list_status": "L"},
		"ts_code,symbol,name,exchange,industry,list_date,delist_date,market",
	)
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)

	out := make([]*security.Security, 0, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		sec := &security.Security{
			StockCode:   strings.ToUpper(cellString(row[idx("symbol")])),
			StockName:   cellString(row[idx("name")]),
			Market:      valueobject.MarketCN,
			Exchange:    strings.ToUpper(cellString(row[idx("exchange")])),
			AssetType:   valueobject.AssetTypeStock,
			Industry:    cellString(row[idx("industry")]),
			Status:      valueobject.StatusListed,
		}
		sec.ListingDate = parseTushareDate(cellString(row[idx("list_date")]))
		sec.DelistingDate = parseTushareDate(cellString(row[idx("delist_date")]))
		if !sec.DelistingDate.IsZero() {
			sec.Status = valueobject.StatusDelisted
		}
		out = append(out, sec)
	}
	return out, nil
}

func indexer(fields []string) func(string) int {
	cache := make(map[string]int, len(fields))
	for i, f := range fields {
		cache[f] = i
	}
	return func(name string) int {
		if i, ok := cache[name]; ok {
			return i
		}
		return -1
	}
}

func parseTushareDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{"20060102", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

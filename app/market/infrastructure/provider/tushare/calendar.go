package tushare

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FetchCalendar implements provider.CalendarFetcher using Tushare trade_cal.
func (p *Provider) FetchCalendar(ctx context.Context, market valueobject.Market, r valueobject.DateRange) ([]calendar.TradingDay, error) {
	if market == "" {
		market = valueobject.MarketCN
	}
	params := map[string]any{"exchange": "SSE"}
	if !r.Start.IsZero() {
		params["start_date"] = r.Start.Format("20060102")
	}
	if !r.End.IsZero() {
		params["end_date"] = r.End.Format("20060102")
	}
	resp, err := p.client.Call(ctx, "trade_cal", params, "cal_date,is_open")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make([]calendar.TradingDay, 0, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		t := parseTushareDate(cellString(row[idx("cal_date")]))
		if t.IsZero() {
			continue
		}
		out = append(out, calendar.TradingDay{
			TradeDate: t,
			IsOpen:    cellInt(row[idx("is_open")]) == 1,
		})
	}
	return out, nil
}

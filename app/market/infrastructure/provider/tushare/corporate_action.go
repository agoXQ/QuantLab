package tushare

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FetchCorporateActions returns adjustment factors per trade date as the MVP
// surface. The PRD/TD reserve a richer corporate-action model for V2; we only
// need adj_factor to drive the price adjustment engine for the first version.
func (p *Provider) FetchCorporateActions(ctx context.Context, stockCode string, r valueobject.DateRange) ([]adjustment.CorporateAction, error) {
	tsCode, err := toTushareCode(stockCode)
	if err != nil {
		return nil, err
	}
	params := map[string]any{"ts_code": tsCode}
	if !r.Start.IsZero() {
		params["start_date"] = r.Start.Format("20060102")
	}
	if !r.End.IsZero() {
		params["end_date"] = r.End.Format("20060102")
	}
	resp, err := p.client.Call(ctx, "adj_factor", params, "ts_code,trade_date,adj_factor")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make([]adjustment.CorporateAction, 0, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		out = append(out, adjustment.CorporateAction{
			StockCode:  stockCode,
			ActionDate: cellString(row[idx("trade_date")]),
			ActionType: valueobject.ActionDividend,
			Factor:     cellFloat(row[idx("adj_factor")]),
		})
	}
	return out, nil
}

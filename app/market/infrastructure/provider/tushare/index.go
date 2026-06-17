package tushare

import (
	"context"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
)

// FetchIndexBars implements provider.IndexFetcher using Tushare's index_daily.
func (p *Provider) FetchIndexBars(ctx context.Context, q provider.IndexQuery) ([]*indexbar.IndexBar, error) {
	if strings.TrimSpace(q.IndexCode) == "" {
		return nil, nil
	}
	params := map[string]any{"ts_code": strings.ToUpper(q.IndexCode)}
	if !q.Range.Start.IsZero() {
		params["start_date"] = q.Range.Start.Format("20060102")
	}
	if !q.Range.End.IsZero() {
		params["end_date"] = q.Range.End.Format("20060102")
	}
	resp, err := p.client.Call(ctx, "index_daily", params,
		"ts_code,trade_date,open,high,low,close,vol,amount")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make([]*indexbar.IndexBar, 0, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		out = append(out, &indexbar.IndexBar{
			IndexCode: strings.ToUpper(cellString(row[idx("ts_code")])),
			TradeDate: parseTushareDate(cellString(row[idx("trade_date")])),
			Open:      cellFloat(row[idx("open")]),
			High:      cellFloat(row[idx("high")]),
			Low:       cellFloat(row[idx("low")]),
			Close:     cellFloat(row[idx("close")]),
			Volume:    cellInt(row[idx("vol")]) * 100,
			Amount:    cellFloat(row[idx("amount")]) * 1000,
		})
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

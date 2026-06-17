package tushare

import (
	"context"
	"fmt"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FetchBars implements provider.BarFetcher.
//
// Day bars are pulled from Tushare's `daily` endpoint, week/month bars from
// `weekly` and `monthly`. The adjustment factor lives in `adj_factor` so we
// fetch it in a follow-up call when the user requested adjusted bars.
func (p *Provider) FetchBars(ctx context.Context, q provider.BarQuery) ([]*marketbar.MarketBar, error) {
	api, err := apiNameForPeriod(q.Period)
	if err != nil {
		return nil, err
	}
	tsCode, err := toTushareCode(q.StockCode)
	if err != nil {
		return nil, err
	}
	params := map[string]any{"ts_code": tsCode}
	if !q.Range.Start.IsZero() {
		params["start_date"] = q.Range.Start.Format("20060102")
	}
	if !q.Range.End.IsZero() {
		params["end_date"] = q.Range.End.Format("20060102")
	}
	fields := "ts_code,trade_date,open,high,low,close,vol,amount"
	resp, err := p.client.Call(ctx, api, params, fields)
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)

	out := make([]*marketbar.MarketBar, 0, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		bar := &marketbar.MarketBar{
			StockCode: strings.ToUpper(stripSuffix(cellString(row[idx("ts_code")]))),
			TradeDate: parseTushareDate(cellString(row[idx("trade_date")])),
			Period:    q.Period,
			Open:      cellFloat(row[idx("open")]),
			High:      cellFloat(row[idx("high")]),
			Low:       cellFloat(row[idx("low")]),
			Close:     cellFloat(row[idx("close")]),
			Volume:    cellInt(row[idx("vol")]) * 100, // Tushare vol is in 手 (100 shares)
			Amount:    cellFloat(row[idx("amount")]) * 1000, // amount in 千元
			AdjFactor: 1,
		}
		out = append(out, bar)
	}

	// Tushare returns rows newest-first; reverse to chronological order.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	if q.Period == valueobject.PeriodDay {
		if err := p.attachDailyAdjFactors(ctx, tsCode, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (p *Provider) attachDailyAdjFactors(ctx context.Context, tsCode string, bars []*marketbar.MarketBar) error {
	if len(bars) == 0 {
		return nil
	}
	params := map[string]any{
		"ts_code":    tsCode,
		"start_date": bars[0].TradeDate.Format("20060102"),
		"end_date":   bars[len(bars)-1].TradeDate.Format("20060102"),
	}
	resp, err := p.client.Call(ctx, "adj_factor", params, "ts_code,trade_date,adj_factor")
	if err != nil {
		// Adjustment is best-effort; missing factors should not fail the call.
		return nil
	}
	idx := indexer(resp.Data.Fields)
	factors := make(map[string]float64, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		date := cellString(row[idx("trade_date")])
		factors[date] = cellFloat(row[idx("adj_factor")])
	}
	for _, b := range bars {
		key := b.TradeDate.Format("20060102")
		if v, ok := factors[key]; ok && v > 0 {
			b.AdjFactor = v
		}
	}
	return nil
}

func apiNameForPeriod(p valueobject.Period) (string, error) {
	switch p {
	case "", valueobject.PeriodDay:
		return "daily", nil
	case valueobject.PeriodWeek:
		return "weekly", nil
	case valueobject.PeriodMonth:
		return "monthly", nil
	default:
		return "", fmt.Errorf("unsupported period: %s", p)
	}
}

// toTushareCode converts an internal stock code (e.g. 600519) into Tushare's
// ts_code format (e.g. 600519.SH). Codes that already include a suffix are
// returned unchanged.
func toTushareCode(stockCode string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(stockCode))
	if strings.Contains(code, ".") {
		return code, nil
	}
	if len(code) != 6 {
		return "", fmt.Errorf("invalid stock code: %s", stockCode)
	}
	switch code[0] {
	case '6':
		return code + ".SH", nil
	case '0', '3':
		return code + ".SZ", nil
	case '8', '4':
		return code + ".BJ", nil
	default:
		return code + ".SZ", nil
	}
}

func stripSuffix(tsCode string) string {
	if i := strings.Index(tsCode, "."); i > 0 {
		return tsCode[:i]
	}
	return tsCode
}

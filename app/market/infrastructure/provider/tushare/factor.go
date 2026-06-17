package tushare

import (
	"context"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
)

// FetchFactors uses Tushare's `daily_basic` endpoint to derive PE/PB/PS/PE_TTM
// values, which is the MVP factor surface required by the PRD.
func (p *Provider) FetchFactors(ctx context.Context, q provider.FactorQuery) ([]*factor.Factor, error) {
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
	fields := "ts_code,trade_date,pe,pe_ttm,pb,ps,ps_ttm,turnover_rate,dv_ratio"
	resp, err := p.client.Call(ctx, "daily_basic", params, fields)
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	wanted := factorFilter(q.FactorNames)

	out := make([]*factor.Factor, 0, len(resp.Data.Items)*4)
	for _, row := range resp.Data.Items {
		date := parseTushareDate(cellString(row[idx("trade_date")]))
		if date.IsZero() {
			continue
		}
		stockCode := strings.ToUpper(stripSuffix(cellString(row[idx("ts_code")])))
		appendFactor(&out, wanted, stockCode, date, "PE", cellFloat(row[idx("pe")]))
		appendFactor(&out, wanted, stockCode, date, "PE_TTM", cellFloat(row[idx("pe_ttm")]))
		appendFactor(&out, wanted, stockCode, date, "PB", cellFloat(row[idx("pb")]))
		appendFactor(&out, wanted, stockCode, date, "PS", cellFloat(row[idx("ps")]))
		appendFactor(&out, wanted, stockCode, date, "PS_TTM", cellFloat(row[idx("ps_ttm")]))
		appendFactor(&out, wanted, stockCode, date, "TURNOVER", cellFloat(row[idx("turnover_rate")]))
		appendFactor(&out, wanted, stockCode, date, "DIV_YIELD", cellFloat(row[idx("dv_ratio")]))
	}
	return out, nil
}

func factorFilter(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		out[strings.ToUpper(strings.TrimSpace(n))] = struct{}{}
	}
	return out
}

func appendFactor(out *[]*factor.Factor, wanted map[string]struct{}, code string, date time.Time, name string, value float64) {
	if wanted != nil {
		if _, ok := wanted[name]; !ok {
			return
		}
	}
	*out = append(*out, &factor.Factor{
		StockCode:   code,
		TradeDate:   date,
		FactorName:  name,
		FactorValue: value,
	})
}

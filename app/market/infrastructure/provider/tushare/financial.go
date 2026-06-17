package tushare

import (
	"context"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FetchFinancials returns financial statements joined from Tushare income,
// balancesheet, and cashflow endpoints.
//
// For the MVP we return the income-statement-derived rows merged with the
// balance-sheet snapshot for the same period. A richer implementation can be
// added without breaking the interface.
func (p *Provider) FetchFinancials(ctx context.Context, q provider.FinancialQuery) ([]*financial.FinancialStatement, error) {
	tsCode, err := toTushareCode(q.StockCode)
	if err != nil {
		return nil, err
	}

	income, err := p.fetchIncome(ctx, tsCode, q)
	if err != nil {
		return nil, err
	}
	balance, err := p.fetchBalance(ctx, tsCode, q)
	if err != nil {
		return nil, err
	}
	cashflow, err := p.fetchCashflow(ctx, tsCode, q)
	if err != nil {
		return nil, err
	}

	out := make([]*financial.FinancialStatement, 0, len(income))
	for key, fs := range income {
		if b, ok := balance[key]; ok {
			fs.TotalAssets = b.TotalAssets
			fs.TotalLiabilities = b.TotalLiabilities
			fs.NetAssets = b.NetAssets
		}
		if c, ok := cashflow[key]; ok {
			fs.OperatingCashFlow = c.OperatingCashFlow
			fs.InvestingCashFlow = c.InvestingCashFlow
			fs.FinancingCashFlow = c.FinancingCashFlow
		}
		out = append(out, fs)
	}
	return out, nil
}

func (p *Provider) fetchIncome(ctx context.Context, tsCode string, q provider.FinancialQuery) (map[string]*financial.FinancialStatement, error) {
	params := financialParams(tsCode, q)
	resp, err := p.client.Call(ctx, "income", params, "ts_code,end_date,report_type,revenue,n_income")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make(map[string]*financial.FinancialStatement, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		fs := &financial.FinancialStatement{
			StockCode:  strings.ToUpper(stripSuffix(cellString(row[idx("ts_code")]))),
			ReportDate: parseTushareDate(cellString(row[idx("end_date")])),
			ReportType: mapReportType(cellString(row[idx("report_type")]), q.ReportType),
			Revenue:    cellFloat(row[idx("revenue")]),
			NetProfit:  cellFloat(row[idx("n_income")]),
		}
		out[fs.ReportDate.Format("20060102")] = fs
	}
	return out, nil
}

type balanceRow struct {
	TotalAssets      float64
	TotalLiabilities float64
	NetAssets        float64
}

func (p *Provider) fetchBalance(ctx context.Context, tsCode string, q provider.FinancialQuery) (map[string]balanceRow, error) {
	params := financialParams(tsCode, q)
	resp, err := p.client.Call(ctx, "balancesheet", params, "end_date,total_assets,total_liab,total_hldr_eqy_inc_min_int")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make(map[string]balanceRow, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		key := cellString(row[idx("end_date")])
		out[key] = balanceRow{
			TotalAssets:      cellFloat(row[idx("total_assets")]),
			TotalLiabilities: cellFloat(row[idx("total_liab")]),
			NetAssets:        cellFloat(row[idx("total_hldr_eqy_inc_min_int")]),
		}
	}
	return out, nil
}

type cashflowRow struct {
	OperatingCashFlow float64
	InvestingCashFlow float64
	FinancingCashFlow float64
}

func (p *Provider) fetchCashflow(ctx context.Context, tsCode string, q provider.FinancialQuery) (map[string]cashflowRow, error) {
	params := financialParams(tsCode, q)
	resp, err := p.client.Call(ctx, "cashflow", params, "end_date,n_cashflow_act,n_cashflow_inv_act,n_cash_flows_fnc_act")
	if err != nil {
		return nil, err
	}
	idx := indexer(resp.Data.Fields)
	out := make(map[string]cashflowRow, len(resp.Data.Items))
	for _, row := range resp.Data.Items {
		key := cellString(row[idx("end_date")])
		out[key] = cashflowRow{
			OperatingCashFlow: cellFloat(row[idx("n_cashflow_act")]),
			InvestingCashFlow: cellFloat(row[idx("n_cashflow_inv_act")]),
			FinancingCashFlow: cellFloat(row[idx("n_cash_flows_fnc_act")]),
		}
	}
	return out, nil
}

func financialParams(tsCode string, q provider.FinancialQuery) map[string]any {
	params := map[string]any{"ts_code": tsCode}
	if !q.Range.Start.IsZero() {
		params["start_date"] = q.Range.Start.Format("20060102")
	}
	if !q.Range.End.IsZero() {
		params["end_date"] = q.Range.End.Format("20060102")
	}
	return params
}

func mapReportType(raw string, fallback valueobject.ReportType) valueobject.ReportType {
	// Tushare report_type 1=年报、2=半年报、3=三季报、4=一季报。
	switch raw {
	case "1":
		return valueobject.ReportAnnual
	case "2":
		return valueobject.ReportInterim
	case "3":
		return valueobject.ReportQ3
	case "4":
		return valueobject.ReportQ1
	}
	if fallback != "" {
		return fallback
	}
	return valueobject.ReportAnnual
}

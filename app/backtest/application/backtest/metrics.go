package backtest

import (
	"math"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// metricsConfig wraps the inputs the metrics builder needs.
type metricsConfig struct {
	jobID          int64
	initialCapital float64
	startDate      time.Time
	endDate        time.Time
	now            time.Time
	// AnnualisationDays controls how the annualised return / Sharpe are
	// scaled. Defaults to 252 (A-share trading days per year) when zero.
	annualisationDays float64
	// RiskFreeRate is the daily risk-free rate used in the Sharpe ratio.
	// Defaults to 0 to keep the MVP simple; users can override via config.
	riskFreeRate float64
}

// buildReport collapses the snapshot+trade arrays into a PerformanceReport.
//
// The math is deliberately textbook so results are easy to validate:
//   total_return = final_asset / initial_capital - 1
//   annual_return = total_return * 252 / N where N = number of snapshots
//   volatility = stdev(daily_returns) * sqrt(252)
//   sharpe = (mean_daily_return - risk_free) / stdev * sqrt(252)
//   max_drawdown = max(1 - asset / running_peak)
//   win_rate = winning trades / total trades (round trips approximated by
//             counting SELL trades whose price exceeds the average cost
//             carried at the time of the fill).
func buildReport(cfg metricsConfig, snapshots []portfolio.Snapshot, trades []*trade.Trade) *report.PerformanceReport {
	rep := &report.PerformanceReport{
		JobID:          cfg.jobID,
		StartDate:      cfg.startDate,
		EndDate:        cfg.endDate,
		InitialCapital: cfg.initialCapital,
		GeneratedAt:    cfg.now,
		EquityCurve:    make([]report.EquityPoint, 0, len(snapshots)),
	}
	if cfg.initialCapital <= 0 || len(snapshots) == 0 {
		return rep
	}
	annualDays := cfg.annualisationDays
	if annualDays <= 0 {
		annualDays = 252
	}

	peak := snapshots[0].TotalAsset
	dailyReturns := make([]float64, 0, len(snapshots))
	prev := cfg.initialCapital
	maxDD := 0.0

	for _, s := range snapshots {
		if s.TotalAsset > peak {
			peak = s.TotalAsset
		}
		var dd float64
		if peak > 0 {
			dd = 1 - s.TotalAsset/peak
			if dd > maxDD {
				maxDD = dd
			}
		}
		ret := 0.0
		if prev > 0 {
			ret = s.TotalAsset/prev - 1
		}
		dailyReturns = append(dailyReturns, ret)
		rep.EquityCurve = append(rep.EquityCurve, report.EquityPoint{
			TradeDate:  s.TradeDate,
			TotalAsset: s.TotalAsset,
			Drawdown:   dd,
			Return:     s.TotalAsset/cfg.initialCapital - 1,
		})
		prev = s.TotalAsset
	}

	final := snapshots[len(snapshots)-1].TotalAsset
	rep.FinalAsset = final
	rep.TotalReturn = final/cfg.initialCapital - 1
	rep.MaxDrawdown = maxDD

	if len(dailyReturns) > 0 {
		rep.AnnualReturn = rep.TotalReturn * annualDays / float64(len(dailyReturns))
	}
	if len(dailyReturns) > 1 {
		mean := 0.0
		for _, r := range dailyReturns {
			mean += r
		}
		mean /= float64(len(dailyReturns))
		variance := 0.0
		for _, r := range dailyReturns {
			diff := r - mean
			variance += diff * diff
		}
		variance /= float64(len(dailyReturns) - 1)
		stdev := math.Sqrt(variance)
		rep.Volatility = stdev * math.Sqrt(annualDays)
		if stdev > 0 {
			rep.SharpeRatio = (mean - cfg.riskFreeRate) / stdev * math.Sqrt(annualDays)
		}
	}

	rep.TradeCount = len(trades)
	rep.WinRate = computeWinRate(trades)
	return rep
}

// computeWinRate returns the fraction of profitable round-trips.
//
// We track per-stock running average cost and only count a SELL as winning
// when its fill price strictly exceeds the carried cost. BUYs do not count
// toward win rate; this keeps the metric stable when a strategy holds a
// position for several days before exiting.
func computeWinRate(trades []*trade.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	type lot struct {
		qty  int64
		cost float64
	}
	holdings := make(map[string]*lot)
	wins := 0
	exits := 0
	for _, t := range trades {
		if t == nil {
			continue
		}
		switch t.Side {
		case valueobject.OrderSideBuy:
			pos, ok := holdings[t.StockCode]
			if !ok {
				holdings[t.StockCode] = &lot{qty: t.Quantity, cost: t.Price}
				continue
			}
			newQty := pos.qty + t.Quantity
			pos.cost = (pos.cost*float64(pos.qty) + t.Price*float64(t.Quantity)) / float64(newQty)
			pos.qty = newQty
		case valueobject.OrderSideSell:
			pos, ok := holdings[t.StockCode]
			if !ok {
				continue
			}
			exits++
			if t.Price > pos.cost {
				wins++
			}
			pos.qty -= t.Quantity
			if pos.qty <= 0 {
				delete(holdings, t.StockCode)
			}
		}
	}
	if exits == 0 {
		return 0
	}
	return float64(wins) / float64(exits)
}

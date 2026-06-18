package marketdata

import (
	"context"
	"time"

	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FromMarketService is a marketdata.Provider backed by the in-process
// Market Data application service. It is the default wiring when the
// monolith deploys Backtest and Market in the same binary; a gRPC adapter
// is the natural drop-in replacement once the services split.
//
// The adapter requests pre-adjusted prices so the engine never has to
// re-apply factors mid-replay; this matches the design rule that all bars
// hitting Backtest carry the same adjustment basis.
type FromMarketService struct {
	svc        appMarket.Service
	period     marketVO.Period
	adjustment marketVO.Adjustment
}

// NewFromMarketService wires the adapter using day bars and the configured
// adjustment mode. Empty / invalid adjustment falls back to the platform
// default ("pre", forward-adjusted).
func NewFromMarketService(svc appMarket.Service, adjustment marketVO.Adjustment) *FromMarketService {
	if !adjustment.IsValid() {
		adjustment = marketVO.AdjustmentPre
	}
	return &FromMarketService{
		svc:        svc,
		period:     marketVO.PeriodDay,
		adjustment: adjustment,
	}
}

// LoadBars implements marketdata.Provider.
func (a *FromMarketService) LoadBars(ctx context.Context, req dommarket.BarsRequest) (map[string][]dommarket.Bar, error) {
	out := make(map[string][]dommarket.Bar, len(req.StockCodes))
	rng := toMarketRange(req.Range)
	for _, code := range req.StockCodes {
		bars, err := a.svc.GetBars(ctx, appMarket.GetBarsQuery{
			StockCode:   code,
			Period:      a.period,
			Range:       rng,
			Adjustment:  a.adjustment,
			DataVersion: req.DataVersion,
		})
		if err != nil {
			out[code] = nil
			continue
		}
		converted := make([]dommarket.Bar, len(bars.Items))
		for i, b := range bars.Items {
			converted[i] = dommarket.Bar{
				StockCode: b.StockCode,
				TradeDate: b.TradeDate,
				Open:      b.Open,
				High:      b.High,
				Low:       b.Low,
				Close:     b.Close,
				Volume:    b.Volume,
				Amount:    b.Amount,
			}
		}
		out[code] = converted
	}
	return out, nil
}

// TradingDays implements marketdata.Provider.
func (a *FromMarketService) TradingDays(ctx context.Context, req dommarket.CalendarRequest) ([]time.Time, error) {
	rng := toMarketRange(req.Range)
	res, err := a.svc.GetCalendar(ctx, appMarket.CalendarQuery{Range: rng})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	days := make([]time.Time, 0, len(res.Days))
	for _, d := range res.Days {
		if !d.IsOpen {
			continue
		}
		days = append(days, d.TradeDate)
	}
	return days, nil
}

func toMarketRange(r valueobject.DateRange) marketVO.DateRange {
	return marketVO.DateRange{Start: r.Start, End: r.End}
}

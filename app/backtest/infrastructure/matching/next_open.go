// Package matching provides matching engine implementations.
package matching

import (
	"math"
	"time"

	dommatch "github.com/agoXQ/QuantLab/app/backtest/domain/matching"
	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// NextOpenEngine fills orders at the next bar's open price with a fixed
// slippage band: BUY pays Open*(1+slippage), SELL receives Open*(1-slippage).
//
// MVP semantics:
//   - Halted bars reject the order with reason "halted".
//   - LimitUp blocks BUY (we cannot chase the limit-up close); LimitDown
//     blocks SELL.
//   - Missing bars (no key in the map) reject with reason "no_bar".
//   - Volume is not enforced because daily replay treats Volume as
//     informational; this is fine for the MVP A-share use case where the
//     universe is liquid by construction.
type NextOpenEngine struct {
	commissionRate float64
	stampDutyRate  float64
	minCommission  float64
	slippageRate   float64
	now            func() time.Time
}

// EngineConfig bundles the parameters NextOpenEngine needs.
type EngineConfig struct {
	CommissionRate float64
	StampDutyRate  float64
	MinCommission  float64
	SlippageRate   float64
	Now            func() time.Time
}

// NewNextOpenEngine constructs the engine using the supplied parameters.
//
// Zero / negative rates fall back to A-share defaults so a misconfigured
// caller still gets a non-nonsensical result; the application service is
// responsible for surfacing the warning when it normalizes Config.
func NewNextOpenEngine(cfg EngineConfig) *NextOpenEngine {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.SlippageRate < 0 {
		cfg.SlippageRate = 0
	}
	if cfg.CommissionRate < 0 {
		cfg.CommissionRate = 0
	}
	if cfg.StampDutyRate < 0 {
		cfg.StampDutyRate = 0
	}
	if cfg.MinCommission < 0 {
		cfg.MinCommission = 0
	}
	return &NextOpenEngine{
		commissionRate: cfg.CommissionRate,
		stampDutyRate:  cfg.StampDutyRate,
		minCommission:  cfg.MinCommission,
		slippageRate:   cfg.SlippageRate,
		now:            cfg.Now,
	}
}

// Match implements matching.Engine.
func (e *NextOpenEngine) Match(orders []*order.Order, bars map[string]dommatch.BarSnapshot) ([]*trade.Trade, []*order.Order) {
	trades := make([]*trade.Trade, 0, len(orders))
	rejects := make([]*order.Order, 0)
	now := e.now()

	for _, ord := range orders {
		if ord == nil {
			continue
		}
		bar, ok := bars[ord.StockCode]
		if !ok || bar.Open <= 0 || math.IsNaN(bar.Open) {
			ord.Status = valueobject.OrderStatusRejected
			ord.Reason = "no_bar"
			rejects = append(rejects, ord)
			continue
		}
		if bar.Halted {
			ord.Status = valueobject.OrderStatusRejected
			ord.Reason = "halted"
			rejects = append(rejects, ord)
			continue
		}
		if ord.Side == valueobject.OrderSideBuy && bar.LimitUp {
			ord.Status = valueobject.OrderStatusRejected
			ord.Reason = "limit_up"
			rejects = append(rejects, ord)
			continue
		}
		if ord.Side == valueobject.OrderSideSell && bar.LimitDown {
			ord.Status = valueobject.OrderStatusRejected
			ord.Reason = "limit_down"
			rejects = append(rejects, ord)
			continue
		}

		fillPrice := bar.Open
		switch ord.Side {
		case valueobject.OrderSideBuy:
			fillPrice = bar.Open * (1 + e.slippageRate)
		case valueobject.OrderSideSell:
			fillPrice = bar.Open * (1 - e.slippageRate)
		}

		gross := fillPrice * float64(ord.Quantity)
		commission := math.Max(gross*e.commissionRate, e.minCommission)
		stampDuty := 0.0
		if ord.Side == valueobject.OrderSideSell {
			stampDuty = gross * e.stampDutyRate
		}
		slippageCost := math.Abs(fillPrice-bar.Open) * float64(ord.Quantity)

		ord.Status = valueobject.OrderStatusFilled
		filled := now
		ord.FilledAt = &filled
		ord.FilledPrice = fillPrice
		ord.FilledQty = ord.Quantity

		trades = append(trades, &trade.Trade{
			JobID:      ord.JobID,
			OrderID:    ord.ID,
			StockCode:  ord.StockCode,
			Side:       ord.Side,
			Quantity:   ord.Quantity,
			Price:      fillPrice,
			Commission: commission,
			StampDuty:  stampDuty,
			Slippage:   slippageCost,
			TradeTime:  now,
		})
	}

	return trades, rejects
}

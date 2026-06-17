// Package adjustment implements the price adjustment domain service.
//
// We use the upstream-provided AdjFactor on each MarketBar (Tushare/AkShare
// publishes it) and apply forward-adjustment (前复权) by default, which is the
// MVP requirement called out in the PRD and TD documents.
package adjustment

import (
	"context"

	domainAdj "github.com/agoXQ/QuantLab/app/market/domain/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// FactorAdjuster computes adjusted prices using the AdjFactor stored on each
// raw bar.
//
// Pre-adjustment formula:
//
//	adjusted_price = raw_price * adj_factor / latest_adj_factor
//
// Post-adjustment formula:
//
//	adjusted_price = raw_price * adj_factor
type FactorAdjuster struct{}

// NewFactorAdjuster returns the default Adjuster implementation.
func NewFactorAdjuster() domainAdj.Adjuster {
	return &FactorAdjuster{}
}

// Apply implements domainAdj.Adjuster.
func (a *FactorAdjuster) Apply(ctx context.Context, bars []*marketbar.MarketBar, mode valueobject.Adjustment) ([]*marketbar.MarketBar, error) {
	if len(bars) == 0 || mode == valueobject.AdjustmentNone {
		return bars, nil
	}

	out := make([]*marketbar.MarketBar, len(bars))
	switch mode {
	case valueobject.AdjustmentPre:
		latest := bars[len(bars)-1].AdjFactor
		if latest == 0 {
			latest = 1
		}
		for i, b := range bars {
			adjusted := *b
			ratio := safeRatio(b.AdjFactor, latest)
			adjusted.Open = b.Open * ratio
			adjusted.High = b.High * ratio
			adjusted.Low = b.Low * ratio
			adjusted.Close = b.Close * ratio
			out[i] = &adjusted
		}
	case valueobject.AdjustmentPost:
		for i, b := range bars {
			adjusted := *b
			factor := b.AdjFactor
			if factor == 0 {
				factor = 1
			}
			adjusted.Open = b.Open * factor
			adjusted.High = b.High * factor
			adjusted.Low = b.Low * factor
			adjusted.Close = b.Close * factor
			out[i] = &adjusted
		}
	default:
		return bars, nil
	}
	return out, nil
}

func safeRatio(num, den float64) float64 {
	if num == 0 {
		num = 1
	}
	if den == 0 {
		return 1
	}
	return num / den
}

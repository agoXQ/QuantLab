package tests

import (
	"context"
	"math"
	"testing"
	"time"

	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

func TestForwardAdjustmentRescalesByLatestFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := []*marketbar.MarketBar{
		{StockCode: "600519", TradeDate: now, Period: valueobject.PeriodDay, Close: 100, AdjFactor: 1},
		{StockCode: "600519", TradeDate: now.AddDate(0, 0, 1), Period: valueobject.PeriodDay, Close: 200, AdjFactor: 2},
	}

	adj := infraAdj.NewFactorAdjuster()
	out, err := adj.Apply(context.Background(), bars, valueobject.AdjustmentPre)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(out))
	}
	if math.Abs(out[0].Close-50) > 1e-6 {
		t.Fatalf("expected close=50, got %v", out[0].Close)
	}
	if math.Abs(out[1].Close-200) > 1e-6 {
		t.Fatalf("latest bar should remain 200, got %v", out[1].Close)
	}
}

func TestPostAdjustmentMultipliesByFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := []*marketbar.MarketBar{
		{Close: 10, AdjFactor: 1, TradeDate: now},
		{Close: 20, AdjFactor: 2, TradeDate: now.AddDate(0, 0, 1)},
	}
	out, err := infraAdj.NewFactorAdjuster().Apply(context.Background(), bars, valueobject.AdjustmentPost)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if math.Abs(out[1].Close-40) > 1e-6 {
		t.Fatalf("post adjustment expected 40, got %v", out[1].Close)
	}
}

func TestNoneReturnsInputUnchanged(t *testing.T) {
	bars := []*marketbar.MarketBar{{Close: 5, AdjFactor: 1}}
	out, err := infraAdj.NewFactorAdjuster().Apply(context.Background(), bars, valueobject.AdjustmentNone)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if &out[0] == &bars[0] {
		// Pointer identity is fine; we just verify content equality.
	}
	if out[0].Close != 5 {
		t.Fatalf("AdjustmentNone should not modify prices, got %v", out[0].Close)
	}
}

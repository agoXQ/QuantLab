package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
)

type fixture struct {
	svc       appMarket.Service
	bars      *fakeBarRepo
	secs      *fakeSecurityRepo
	versions  *fakeVersionRepo
}

func newFixture() *fixture {
	secs := newFakeSecurityRepo()
	bars := newFakeBarRepo()
	versions := &fakeVersionRepo{}

	deps := appMarket.Dependencies{
		Securities:   secs,
		Bars:         bars,
		Financials:   &fakeFinancialRepo{},
		Factors:      &fakeFactorRepo{},
		Indexes:      &fakeIndexRepo{},
		Calendar:     &fakeCalendarRepo{},
		DataVersions: versions,
		Adjuster:     infraAdj.NewFactorAdjuster(),
	}
	return &fixture{
		svc:      appMarket.NewService(deps),
		bars:     bars,
		secs:     secs,
		versions: versions,
	}
}

func TestService_GetSecurityNotFound(t *testing.T) {
	f := newFixture()
	_, err := f.svc.GetSecurity(context.Background(), "600519")
	if !errors.Is(err, domainErr.ErrSecurityNotFound) {
		t.Fatalf("expected ErrSecurityNotFound, got %v", err)
	}
}

func TestService_GetSecurityNormalized(t *testing.T) {
	f := newFixture()
	_ = f.secs.Upsert(context.Background(), &security.Security{
		StockCode: "600519",
		StockName: "贵州茅台",
		Market:    valueobject.MarketCN,
	})
	sec, err := f.svc.GetSecurity(context.Background(), "  600519  ")
	if err != nil {
		t.Fatalf("GetSecurity: %v", err)
	}
	if sec.StockCode != "600519" {
		t.Fatalf("expected normalized code, got %s", sec.StockCode)
	}
}

func TestService_GetBarsValidatesInput(t *testing.T) {
	f := newFixture()
	_, err := f.svc.GetBars(context.Background(), appMarket.GetBarsQuery{StockCode: ""})
	if !errors.Is(err, domainErr.ErrInvalidStockCode) {
		t.Fatalf("expected ErrInvalidStockCode, got %v", err)
	}
}

func TestService_GetBarsAppliesAdjustment(t *testing.T) {
	f := newFixture()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = f.bars.BulkUpsert(context.Background(), []*marketbar.MarketBar{
		{StockCode: "600519", Period: valueobject.PeriodDay, TradeDate: now, Close: 100, AdjFactor: 1},
		{StockCode: "600519", Period: valueobject.PeriodDay, TradeDate: now.AddDate(0, 0, 1), Close: 200, AdjFactor: 2},
	})
	res, err := f.svc.GetBars(context.Background(), appMarket.GetBarsQuery{
		StockCode: "600519",
		Period:    valueobject.PeriodDay,
	})
	if err != nil {
		t.Fatalf("GetBars: %v", err)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(res.Items))
	}
	if res.Items[0].Close == 100 {
		t.Fatalf("expected first bar to be forward-adjusted, got 100")
	}
}

func TestService_GetBarsResolvesLatestVersion(t *testing.T) {
	f := newFixture()
	_ = f.versions.Create(context.Background(), &dataversion.DataVersion{Version: "2026.01", CreatedAt: time.Now()})
	_ = f.versions.Create(context.Background(), &dataversion.DataVersion{Version: "2026.02", CreatedAt: time.Now()})
	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	_ = f.bars.BulkUpsert(context.Background(), []*marketbar.MarketBar{
		{StockCode: "600519", Period: valueobject.PeriodDay, TradeDate: now, Close: 100, AdjFactor: 1, DataVersion: "2026.02"},
	})
	res, err := f.svc.GetBars(context.Background(), appMarket.GetBarsQuery{
		StockCode: "600519",
		Period:    valueobject.PeriodDay,
	})
	if err != nil {
		t.Fatalf("GetBars: %v", err)
	}
	if res.DataVersion != "2026.02" {
		t.Fatalf("expected latest version 2026.02, got %s", res.DataVersion)
	}
}

func TestService_ListVersions(t *testing.T) {
	f := newFixture()
	_ = f.versions.Create(context.Background(), &dataversion.DataVersion{Version: "2026.01"})
	res, err := f.svc.ListVersions(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected 1 version, got %d", len(res.Items))
	}
}

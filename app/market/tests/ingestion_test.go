package tests

import (
	"context"
	"testing"
	"time"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	domainEvent "github.com/agoXQ/QuantLab/app/market/domain/event"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/provider/faketushare"
)

type capturingPublisher struct {
	events []domainEvent.Event
}

func (p *capturingPublisher) Publish(_ context.Context, e domainEvent.Event) error {
	p.events = append(p.events, e)
	return nil
}

func TestIngestion_BarsEmitsEvent(t *testing.T) {
	prov := faketushare.NewProvider()
	prov.Bars = []*marketbar.MarketBar{
		{StockCode: "600519", Period: valueobject.PeriodDay, TradeDate: time.Now().UTC(), Close: 100, AdjFactor: 1},
	}
	bars := newFakeBarRepo()
	versions := &fakeVersionRepo{}
	publisher := &capturingPublisher{}

	svc := appMarket.NewIngestionService(appMarket.IngestionDeps{
		Provider:     prov,
		Securities:   newFakeSecurityRepo(),
		Bars:         bars,
		Financials:   &fakeFinancialRepo{},
		Factors:      &fakeFactorRepo{},
		Indexes:      &fakeIndexRepo{},
		Calendar:     &fakeCalendarRepo{},
		DataVersions: versions,
		Publisher:    publisher,
		Clock:        time.Now,
	})

	dv, err := svc.CreateVersion(context.Background(), "test")
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	count, err := svc.IngestBars(context.Background(), provider.BarQuery{
		StockCode: "600519",
		Period:    valueobject.PeriodDay,
	}, dv.Version)
	if err != nil {
		t.Fatalf("IngestBars: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 bar ingested, got %d", count)
	}
	if len(publisher.events) < 2 {
		t.Fatalf("expected DataVersionCreated + MarketDataUpdated events, got %d", len(publisher.events))
	}
}

func TestIngestion_SecuritiesNormalizes(t *testing.T) {
	prov := faketushare.NewProvider()
	prov.Securities = []*security.Security{{StockCode: " 600519 ", Market: valueobject.MarketCN}}
	repo := newFakeSecurityRepo()
	svc := appMarket.NewIngestionService(appMarket.IngestionDeps{
		Provider:     prov,
		Securities:   repo,
		Bars:         newFakeBarRepo(),
		Financials:   &fakeFinancialRepo{},
		Factors:      &fakeFactorRepo{},
		Indexes:      &fakeIndexRepo{},
		Calendar:     &fakeCalendarRepo{},
		DataVersions: &fakeVersionRepo{},
	})
	if _, err := svc.IngestSecurities(context.Background(), valueobject.MarketCN); err != nil {
		t.Fatalf("IngestSecurities: %v", err)
	}
	got, err := repo.GetByCode(context.Background(), "600519")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got.StockCode != "600519" {
		t.Fatalf("expected trimmed code, got %q", got.StockCode)
	}
}

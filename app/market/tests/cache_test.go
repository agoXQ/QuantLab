package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
)

type memoryCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMemoryCache() *memoryCache { return &memoryCache{data: map[string][]byte{}} }

func (c *memoryCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key], nil
}

func (c *memoryCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	c.data[key] = cp
	return nil
}

func (c *memoryCache) Del(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.data, k)
	}
	return nil
}

func TestCachedService_GetSecurityCachesValue(t *testing.T) {
	secs := newFakeSecurityRepo()
	_ = secs.Upsert(context.Background(), &security.Security{StockCode: "600519", StockName: "Moutai", Market: valueobject.MarketCN})

	deps := appMarket.Dependencies{
		Securities:   secs,
		Bars:         newFakeBarRepo(),
		Financials:   &fakeFinancialRepo{},
		Factors:      &fakeFactorRepo{},
		Indexes:      &fakeIndexRepo{},
		Calendar:     &fakeCalendarRepo{},
		DataVersions: &fakeVersionRepo{},
		Adjuster:     infraAdj.NewFactorAdjuster(),
	}
	base := appMarket.NewService(deps)
	cache := newMemoryCache()
	svc := appMarket.NewCachedService(base, cache, 0)

	if _, err := svc.GetSecurity(context.Background(), "600519"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	// Wipe the underlying repo; the cached service should still return data.
	secs.data = map[string]*security.Security{}
	got, err := svc.GetSecurity(context.Background(), "600519")
	if err != nil {
		t.Fatalf("cached call: %v", err)
	}
	if got.StockCode != "600519" {
		t.Fatalf("expected cached security, got %+v", got)
	}
}

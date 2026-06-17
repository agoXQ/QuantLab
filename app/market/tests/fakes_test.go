package tests

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// --- in-memory repositories used by application tests ---

type fakeSecurityRepo struct {
	mu   sync.Mutex
	data map[string]*security.Security
}

func newFakeSecurityRepo() *fakeSecurityRepo {
	return &fakeSecurityRepo{data: map[string]*security.Security{}}
}

func (r *fakeSecurityRepo) GetByCode(_ context.Context, code string) (*security.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sec, ok := r.data[strings.ToUpper(code)]
	if !ok {
		return nil, domainErr.ErrSecurityNotFound
	}
	cp := *sec
	return &cp, nil
}

func (r *fakeSecurityRepo) List(_ context.Context, q security.ListQuery) ([]*security.Security, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*security.Security, 0, len(r.data))
	for _, s := range r.data {
		if q.Market != "" && s.Market != q.Market {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StockCode < out[j].StockCode })
	return out, "", nil
}

func (r *fakeSecurityRepo) Upsert(_ context.Context, sec *security.Security) error {
	return r.BulkUpsert(context.Background(), []*security.Security{sec})
}

func (r *fakeSecurityRepo) BulkUpsert(_ context.Context, list []*security.Security) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sec := range list {
		sec.Normalize()
		r.data[sec.StockCode] = sec
	}
	return nil
}

type fakeBarRepo struct {
	mu   sync.Mutex
	data []*marketbar.MarketBar
}

func newFakeBarRepo() *fakeBarRepo { return &fakeBarRepo{} }

func (r *fakeBarRepo) Range(_ context.Context, q marketbar.RangeQuery) ([]*marketbar.MarketBar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*marketbar.MarketBar, 0, len(r.data))
	for _, b := range r.data {
		if !strings.EqualFold(b.StockCode, q.StockCode) {
			continue
		}
		if q.Period != "" && b.Period != q.Period {
			continue
		}
		if !q.Range.Start.IsZero() && b.TradeDate.Before(q.Range.Start) {
			continue
		}
		if !q.Range.End.IsZero() && b.TradeDate.After(q.Range.End) {
			continue
		}
		cp := *b
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TradeDate.Before(out[j].TradeDate) })
	return out, nil
}

func (r *fakeBarRepo) Latest(_ context.Context, _ string, _ valueobject.Period, _ time.Time) (*marketbar.MarketBar, error) {
	return nil, nil
}

func (r *fakeBarRepo) BulkUpsert(_ context.Context, bars []*marketbar.MarketBar) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range bars {
		cp := *b
		r.data = append(r.data, &cp)
	}
	return nil
}

type fakeFinancialRepo struct{ data []*financial.FinancialStatement }

func (r *fakeFinancialRepo) List(_ context.Context, q financial.ListQuery) ([]*financial.FinancialStatement, error) {
	out := make([]*financial.FinancialStatement, 0, len(r.data))
	for _, f := range r.data {
		if !strings.EqualFold(f.StockCode, q.StockCode) {
			continue
		}
		if q.ReportType != "" && f.ReportType != q.ReportType {
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *fakeFinancialRepo) BulkUpsert(_ context.Context, list []*financial.FinancialStatement) error {
	r.data = append(r.data, list...)
	return nil
}

type fakeFactorRepo struct{ data []*factor.Factor }

func (r *fakeFactorRepo) List(_ context.Context, q factor.ListQuery) ([]*factor.Factor, error) {
	out := make([]*factor.Factor, 0, len(r.data))
	for _, f := range r.data {
		if !strings.EqualFold(f.StockCode, q.StockCode) {
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *fakeFactorRepo) BulkUpsert(_ context.Context, factors []*factor.Factor) error {
	r.data = append(r.data, factors...)
	return nil
}

type fakeIndexRepo struct{ data []*indexbar.IndexBar }

func (r *fakeIndexRepo) List(_ context.Context, q indexbar.RangeQuery) ([]*indexbar.IndexBar, error) {
	out := make([]*indexbar.IndexBar, 0, len(r.data))
	for _, b := range r.data {
		if !strings.EqualFold(b.IndexCode, q.IndexCode) {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *fakeIndexRepo) BulkUpsert(_ context.Context, bars []*indexbar.IndexBar) error {
	r.data = append(r.data, bars...)
	return nil
}

type fakeCalendarRepo struct{ data []calendar.TradingDay }

func (r *fakeCalendarRepo) Range(_ context.Context, rg valueobject.DateRange) ([]calendar.TradingDay, error) {
	out := make([]calendar.TradingDay, 0, len(r.data))
	for _, d := range r.data {
		if !rg.Start.IsZero() && d.TradeDate.Before(rg.Start) {
			continue
		}
		if !rg.End.IsZero() && d.TradeDate.After(rg.End) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *fakeCalendarRepo) BulkUpsert(_ context.Context, days []calendar.TradingDay) error {
	r.data = append(r.data, days...)
	return nil
}

type fakeVersionRepo struct {
	mu   sync.Mutex
	data []*dataversion.DataVersion
}

func (r *fakeVersionRepo) Get(_ context.Context, version string) (*dataversion.DataVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range r.data {
		if v.Version == version {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainErr.ErrDataVersionNotFound
}

func (r *fakeVersionRepo) List(_ context.Context, _ int) ([]*dataversion.DataVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*dataversion.DataVersion, len(r.data))
	for i, v := range r.data {
		cp := *v
		out[i] = &cp
	}
	return out, nil
}

func (r *fakeVersionRepo) Latest(_ context.Context) (*dataversion.DataVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.data) == 0 {
		return nil, nil
	}
	cp := *r.data[len(r.data)-1]
	return &cp, nil
}

func (r *fakeVersionRepo) Create(_ context.Context, dv *dataversion.DataVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range r.data {
		if v.Version == dv.Version {
			return errors.New("duplicate version")
		}
	}
	cp := *dv
	r.data = append(r.data, &cp)
	return nil
}

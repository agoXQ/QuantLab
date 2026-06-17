package market

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/cache"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

const (
	cacheKeyPrefixSecurity = "market:security:"
	cacheKeyPrefixBars     = "market:bars:"
	cacheKeyPrefixCalendar = "market:calendar:"

	defaultCacheTTL  = 1 * time.Hour
	securityCacheTTL = 24 * time.Hour
	calendarCacheTTL = 24 * time.Hour
)

// CachedService wraps a Service with a Redis cache for the read-heavy queries.
//
// Only safe-to-cache and idempotent reads are intercepted. Mutations and
// queries with non-deterministic inputs (e.g. cursor pagination) are passed
// through to the underlying service.
type CachedService struct {
	inner Service
	cache cache.Cache
	ttl   time.Duration
}

// NewCachedService returns a CachedService decorator.
func NewCachedService(inner Service, c cache.Cache, ttl time.Duration) Service {
	if ttl == 0 {
		ttl = defaultCacheTTL
	}
	return &CachedService{inner: inner, cache: c, ttl: ttl}
}

func (s *CachedService) GetSecurity(ctx context.Context, stockCode string) (*security.Security, error) {
	key := cacheKeyPrefixSecurity + strings.ToUpper(strings.TrimSpace(stockCode))
	if data, err := s.cache.Get(ctx, key); err == nil && len(data) > 0 {
		var sec security.Security
		if jsonErr := json.Unmarshal(data, &sec); jsonErr == nil {
			return &sec, nil
		}
	}

	sec, err := s.inner.GetSecurity(ctx, stockCode)
	if err != nil {
		return nil, err
	}
	if data, marshalErr := json.Marshal(sec); marshalErr == nil {
		_ = s.cache.Set(ctx, key, data, securityCacheTTL)
	}
	return sec, nil
}

func (s *CachedService) ListSecurities(ctx context.Context, q ListSecuritiesQuery) (*SecurityList, error) {
	return s.inner.ListSecurities(ctx, q)
}

func (s *CachedService) GetBars(ctx context.Context, q GetBarsQuery) (*BarList, error) {
	if q.Range.IsZero() || q.Limit > 0 {
		return s.inner.GetBars(ctx, q)
	}
	period := q.Period
	if period == "" {
		period = valueobject.PeriodDay
	}
	mode := q.Adjustment
	if mode == "" {
		mode = valueobject.AdjustmentPre
	}
	digest := s.inner.CacheKey(
		strings.ToUpper(q.StockCode),
		string(period),
		string(mode),
		valueobject.FormatDate(q.Range.Start),
		valueobject.FormatDate(q.Range.End),
		q.DataVersion,
	)
	key := cacheKeyPrefixBars + digest
	if data, err := s.cache.Get(ctx, key); err == nil && len(data) > 0 {
		var resp BarList
		if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil {
			return &resp, nil
		}
	}
	resp, err := s.inner.GetBars(ctx, q)
	if err != nil {
		return nil, err
	}
	if data, marshalErr := json.Marshal(resp); marshalErr == nil {
		_ = s.cache.Set(ctx, key, data, s.ttl)
	}
	return resp, nil
}

func (s *CachedService) GetFinancials(ctx context.Context, q GetFinancialsQuery) (*FinancialList, error) {
	return s.inner.GetFinancials(ctx, q)
}

func (s *CachedService) GetFactors(ctx context.Context, q GetFactorsQuery) (*FactorList, error) {
	return s.inner.GetFactors(ctx, q)
}

func (s *CachedService) GetIndex(ctx context.Context, q GetIndexQuery) (*IndexList, error) {
	return s.inner.GetIndex(ctx, q)
}

func (s *CachedService) GetCalendar(ctx context.Context, q CalendarQuery) (*CalendarResult, error) {
	if q.Range.IsZero() {
		return s.inner.GetCalendar(ctx, q)
	}
	digest := s.inner.CacheKey(
		valueobject.FormatDate(q.Range.Start),
		valueobject.FormatDate(q.Range.End),
	)
	key := cacheKeyPrefixCalendar + digest
	if data, err := s.cache.Get(ctx, key); err == nil && len(data) > 0 {
		var resp CalendarResult
		if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil {
			return &resp, nil
		}
	}
	resp, err := s.inner.GetCalendar(ctx, q)
	if err != nil {
		return nil, err
	}
	if data, marshalErr := json.Marshal(resp); marshalErr == nil {
		_ = s.cache.Set(ctx, key, data, calendarCacheTTL)
	}
	return resp, nil
}

func (s *CachedService) ListVersions(ctx context.Context, limit int) (*VersionsResult, error) {
	return s.inner.ListVersions(ctx, limit)
}

func (s *CachedService) CacheKey(parts ...string) string {
	return s.inner.CacheKey(parts...)
}

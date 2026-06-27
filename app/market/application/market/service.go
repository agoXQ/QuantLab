package market

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/adjustment"
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

// Service is the application-level interface for the Market Data service.
//
// It encapsulates use cases consumed by the gRPC and REST interface layers,
// and is intentionally kept free of transport concerns.
type Service interface {
	GetSecurity(ctx context.Context, stockCode string) (*security.Security, error)
	ListSecurities(ctx context.Context, q ListSecuritiesQuery) (*SecurityList, error)

	GetBars(ctx context.Context, q GetBarsQuery) (*BarList, error)

	GetFinancials(ctx context.Context, q GetFinancialsQuery) (*FinancialList, error)

	GetFactors(ctx context.Context, q GetFactorsQuery) (*FactorList, error)

	GetIndex(ctx context.Context, q GetIndexQuery) (*IndexList, error)

	GetCalendar(ctx context.Context, q CalendarQuery) (*CalendarResult, error)

	ListVersions(ctx context.Context, limit int) (*VersionsResult, error)

	// CacheKey produces a deterministic cache key fragment for the given query
	// payload. Decorators use it to compose Redis keys.
	CacheKey(parts ...string) string
}

// Dependencies bundles the repositories and domain services needed by the
// application service. Wiring is done in svc/servicecontext.go.
type Dependencies struct {
	Securities   security.Repository
	Bars         marketbar.Repository
	Financials   financial.Repository
	Factors      factor.Repository
	Indexes      indexbar.Repository
	Calendar     calendar.Repository
	DataVersions dataversion.Repository
	Adjuster     adjustment.Adjuster
}

type service struct {
	deps Dependencies
}

// NewService returns the default application service implementation.
func NewService(deps Dependencies) Service {
	return &service{deps: deps}
}

// --- Securities ---

func (s *service) GetSecurity(ctx context.Context, stockCode string) (*security.Security, error) {
	code := strings.TrimSpace(stockCode)
	if code == "" {
		return nil, domainErr.ErrInvalidStockCode
	}
	sec, err := s.deps.Securities.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, domainErr.ErrSecurityNotFound
	}
	return sec, nil
}

func (s *service) ListSecurities(ctx context.Context, q ListSecuritiesQuery) (*SecurityList, error) {
	limit := normalizeLimit(q.Limit, 50, 500)
	items, next, err := s.deps.Securities.List(ctx, security.ListQuery{
		Market:    q.Market,
		Exchange:  q.Exchange,
		AssetType: q.AssetType,
		Industry:  q.Industry,
		Status:    q.Status,
		Cursor:    q.Cursor,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	return &SecurityList{
		Items:      items,
		NextCursor: next,
		HasMore:    next != "",
	}, nil
}

// --- Bars ---

func (s *service) GetBars(ctx context.Context, q GetBarsQuery) (*BarList, error) {
	if strings.TrimSpace(q.StockCode) == "" {
		return nil, domainErr.ErrInvalidStockCode
	}
	period := q.Period
	if period == "" {
		period = valueobject.PeriodDay
	}
	if !period.IsValid() {
		return nil, domainErr.ErrInvalidPeriod
	}

	mode := q.Adjustment
	if mode == "" {
		mode = valueobject.AdjustmentPre
	}
	if !mode.IsValid() {
		return nil, domainErr.ErrInvalidAdjustment
	}
	if err := q.Range.Validate(); err != nil {
		return nil, domainErr.ErrInvalidDateRange
	}

	version, err := s.resolveVersion(ctx, q.DataVersion)
	if err != nil {
		return nil, err
	}

	bars, err := s.deps.Bars.Range(ctx, marketbar.RangeQuery{
		StockCode:   q.StockCode,
		Period:      period,
		Range:       q.Range,
		DataVersion: version,
		Limit:       normalizeLimit(q.Limit, 1000, 10000),
	})
	if err != nil {
		return nil, err
	}

	if mode != valueobject.AdjustmentNone && s.deps.Adjuster != nil {
		bars, err = s.deps.Adjuster.Apply(ctx, bars, mode)
		if err != nil {
			return nil, fmt.Errorf("apply adjustment: %w", err)
		}
	}

	return &BarList{Items: bars, Adjustment: mode, DataVersion: version}, nil
}

// --- Financials ---

func (s *service) GetFinancials(ctx context.Context, q GetFinancialsQuery) (*FinancialList, error) {
	if strings.TrimSpace(q.StockCode) == "" {
		return nil, domainErr.ErrInvalidStockCode
	}
	if q.ReportType != "" && !q.ReportType.IsValid() {
		return nil, domainErr.ErrInvalidReportType
	}
	if err := q.Range.Validate(); err != nil {
		return nil, domainErr.ErrInvalidDateRange
	}

	version, err := s.resolveVersion(ctx, q.DataVersion)
	if err != nil {
		return nil, err
	}

	limit := normalizeLimit(q.Limit, 40, 200)
	items, err := s.deps.Financials.List(ctx, financial.ListQuery{
		StockCode:   q.StockCode,
		ReportType:  q.ReportType,
		Range:       q.Range,
		DataVersion: version,
		Cursor:      q.Cursor,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}

	return &FinancialList{Items: items, DataVersion: version}, nil
}

// --- Factors ---

func (s *service) GetFactors(ctx context.Context, q GetFactorsQuery) (*FactorList, error) {
	if strings.TrimSpace(q.StockCode) == "" {
		return nil, domainErr.ErrInvalidStockCode
	}
	if err := q.Range.Validate(); err != nil {
		return nil, domainErr.ErrInvalidDateRange
	}

	version, err := s.resolveVersion(ctx, q.DataVersion)
	if err != nil {
		return nil, err
	}

	items, err := s.deps.Factors.List(ctx, factor.ListQuery{
		StockCode:   q.StockCode,
		FactorNames: q.FactorNames,
		Range:       q.Range,
		DataVersion: version,
		Limit:       normalizeLimit(q.Limit, 1000, 10000),
	})
	if err != nil {
		return nil, err
	}

	return &FactorList{Items: items, DataVersion: version}, nil
}

// --- Index ---

func (s *service) GetIndex(ctx context.Context, q GetIndexQuery) (*IndexList, error) {
	if strings.TrimSpace(q.IndexCode) == "" {
		return nil, domainErr.ErrIndexNotFound
	}
	if err := q.Range.Validate(); err != nil {
		return nil, domainErr.ErrInvalidDateRange
	}
	version, err := s.resolveVersion(ctx, q.DataVersion)
	if err != nil {
		return nil, err
	}

	items, err := s.deps.Indexes.List(ctx, indexbar.RangeQuery{
		IndexCode:   q.IndexCode,
		Range:       q.Range,
		DataVersion: version,
		Limit:       normalizeLimit(0, 1000, 10000),
	})
	if err != nil {
		return nil, err
	}

	return &IndexList{Items: items, DataVersion: version}, nil
}

// --- Calendar / Versions ---

func (s *service) GetCalendar(ctx context.Context, q CalendarQuery) (*CalendarResult, error) {
	if err := q.Range.Validate(); err != nil {
		return nil, domainErr.ErrInvalidDateRange
	}
	days, err := s.deps.Calendar.Range(ctx, q.Range)
	if err != nil {
		return nil, err
	}
	return &CalendarResult{Days: days}, nil
}

func (s *service) ListVersions(ctx context.Context, limit int) (*VersionsResult, error) {
	items, err := s.deps.DataVersions.List(ctx, normalizeLimit(limit, 50, 500))
	if err != nil {
		return nil, err
	}
	return &VersionsResult{Items: items}, nil
}

// --- Helpers ---

func (s *service) resolveVersion(ctx context.Context, requested string) (string, error) {
	if requested != "" {
		dv, err := s.deps.DataVersions.Get(ctx, requested)
		if err != nil {
			return "", err
		}
		if dv == nil {
			return "", domainErr.ErrDataVersionNotFound
		}
		return dv.Version, nil
	}
	latest, err := s.deps.DataVersions.Latest(ctx)
	if err != nil {
		if errors.Is(err, domainErr.ErrDataVersionNotFound) {
			return "", nil
		}
		return "", err
	}
	if latest == nil {
		return "", nil
	}
	return latest.Version, nil
}

func (s *service) CacheKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func normalizeLimit(raw, def, max int) int {
	if raw <= 0 {
		return def
	}
	if raw > max {
		return max
	}
	return raw
}

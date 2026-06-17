package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	domainEvent "github.com/agoXQ/QuantLab/app/market/domain/event"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	"github.com/google/uuid"
)

// IngestionService orchestrates the data acquisition pipeline:
//
//	Provider -> (Cleaner) -> Storage -> Event
//
// The service is intentionally thin: each step is delegated to a domain
// abstraction, and a fresh DataVersion is allocated per ingestion run so the
// resulting facts are reproducible.
type IngestionService interface {
	// IngestSecurities pulls the security master from the provider and upserts it.
	IngestSecurities(ctx context.Context, market valueobject.Market) (int, error)

	// IngestCalendar pulls the trading calendar for the given range.
	IngestCalendar(ctx context.Context, market valueobject.Market, r valueobject.DateRange) (int, error)

	// IngestBars pulls daily bars for one security.
	IngestBars(ctx context.Context, q provider.BarQuery, version string) (int, error)

	// IngestFinancials pulls financial statements for one security.
	IngestFinancials(ctx context.Context, q provider.FinancialQuery, version string) (int, error)

	// IngestFactors pulls factor values for one security.
	IngestFactors(ctx context.Context, q provider.FactorQuery, version string) (int, error)

	// IngestIndex pulls index bars for one index.
	IngestIndex(ctx context.Context, q provider.IndexQuery, version string) (int, error)

	// CreateVersion creates a new DataVersion, returning the version string.
	CreateVersion(ctx context.Context, description string) (*dataversion.DataVersion, error)
}

// IngestionDeps groups the dependencies required by the ingestion service.
type IngestionDeps struct {
	Provider     provider.DataProvider
	Securities   security.Repository
	Bars         marketbar.Repository
	Financials   financial.Repository
	Factors      factor.Repository
	Indexes      indexbar.Repository
	Calendar     calendar.Repository
	DataVersions dataversion.Repository
	Publisher    domainEvent.Publisher
	Clock        func() time.Time
}

type ingestionService struct {
	deps IngestionDeps
}

// NewIngestionService constructs the default ingestion service.
func NewIngestionService(deps IngestionDeps) IngestionService {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	return &ingestionService{deps: deps}
}

func (s *ingestionService) IngestSecurities(ctx context.Context, market valueobject.Market) (int, error) {
	list, err := s.deps.Provider.FetchSecurities(ctx, market)
	if err != nil {
		return 0, fmt.Errorf("fetch securities: %w", err)
	}
	for _, sec := range list {
		sec.Normalize()
	}
	if err := s.deps.Securities.BulkUpsert(ctx, list); err != nil {
		return 0, fmt.Errorf("upsert securities: %w", err)
	}
	return len(list), nil
}

func (s *ingestionService) IngestCalendar(ctx context.Context, market valueobject.Market, r valueobject.DateRange) (int, error) {
	days, err := s.deps.Provider.FetchCalendar(ctx, market, r)
	if err != nil {
		return 0, fmt.Errorf("fetch calendar: %w", err)
	}
	if err := s.deps.Calendar.BulkUpsert(ctx, days); err != nil {
		return 0, fmt.Errorf("upsert calendar: %w", err)
	}
	return len(days), nil
}

func (s *ingestionService) IngestBars(ctx context.Context, q provider.BarQuery, version string) (int, error) {
	bars, err := s.deps.Provider.FetchBars(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("fetch bars: %w", err)
	}
	for _, b := range bars {
		b.DataVersion = version
		b.StockCode = strings.ToUpper(strings.TrimSpace(b.StockCode))
	}
	if err := s.deps.Bars.BulkUpsert(ctx, bars); err != nil {
		return 0, fmt.Errorf("upsert bars: %w", err)
	}
	if s.deps.Publisher != nil && len(bars) > 0 {
		_ = s.deps.Publisher.Publish(ctx, s.newEvent(domainEvent.EventMarketDataUpdated,
			domainEvent.MarketDataUpdatedPayload{
				TradeDate:   valueobject.FormatDate(bars[len(bars)-1].TradeDate),
				Period:      string(q.Period),
				Count:       len(bars),
				DataVersion: version,
			}))
	}
	return len(bars), nil
}

func (s *ingestionService) IngestFinancials(ctx context.Context, q provider.FinancialQuery, version string) (int, error) {
	items, err := s.deps.Provider.FetchFinancials(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("fetch financials: %w", err)
	}
	for _, f := range items {
		f.DataVersion = version
	}
	if err := s.deps.Financials.BulkUpsert(ctx, items); err != nil {
		return 0, fmt.Errorf("upsert financials: %w", err)
	}
	return len(items), nil
}

func (s *ingestionService) IngestFactors(ctx context.Context, q provider.FactorQuery, version string) (int, error) {
	items, err := s.deps.Provider.FetchFactors(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("fetch factors: %w", err)
	}
	for _, f := range items {
		f.DataVersion = version
	}
	if err := s.deps.Factors.BulkUpsert(ctx, items); err != nil {
		return 0, fmt.Errorf("upsert factors: %w", err)
	}
	if s.deps.Publisher != nil && len(items) > 0 {
		_ = s.deps.Publisher.Publish(ctx, s.newEvent(domainEvent.EventFactorUpdated,
			domainEvent.FactorUpdatedPayload{
				TradeDate:   valueobject.FormatDate(items[len(items)-1].TradeDate),
				Count:       len(items),
				DataVersion: version,
			}))
	}
	return len(items), nil
}

func (s *ingestionService) IngestIndex(ctx context.Context, q provider.IndexQuery, version string) (int, error) {
	items, err := s.deps.Provider.FetchIndexBars(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("fetch index: %w", err)
	}
	for _, b := range items {
		b.DataVersion = version
	}
	if err := s.deps.Indexes.BulkUpsert(ctx, items); err != nil {
		return 0, fmt.Errorf("upsert index: %w", err)
	}
	return len(items), nil
}

func (s *ingestionService) CreateVersion(ctx context.Context, description string) (*dataversion.DataVersion, error) {
	now := s.deps.Clock().UTC()
	dv := &dataversion.DataVersion{
		Version:     now.Format("2006.01.02.150405"),
		Description: description,
		CreatedAt:   now,
	}
	if err := s.deps.DataVersions.Create(ctx, dv); err != nil {
		return nil, fmt.Errorf("create data version: %w", err)
	}
	if s.deps.Publisher != nil {
		_ = s.deps.Publisher.Publish(ctx, s.newEvent(domainEvent.EventDataVersionCreated,
			domainEvent.DataVersionCreatedPayload{
				Version:     dv.Version,
				Description: dv.Description,
			}))
	}
	return dv, nil
}

func (s *ingestionService) newEvent(t domainEvent.EventType, payload any) domainEvent.Event {
	return domainEvent.Event{
		EventID:       uuid.NewString(),
		EventType:     t,
		EventVersion:  domainEvent.EventVersionV1,
		OccurredAt:    s.deps.Clock().UTC(),
		AggregateType: domainEvent.AggregateTypeSecurity,
		Producer:      domainEvent.ProducerMarketService,
		Payload:       payload,
	}
}

package dataport

import (
	"context"
	"fmt"
	"sync"
	"time"

	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"

	domainAdj "github.com/agoXQ/QuantLab/app/market/domain/adjustment"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// RepositoryDataPort is the in-process DataPort that reads directly from the
// Market Data domain repositories.
//
// It is the recommended adapter while QuantLab runs as a single binary
// (Roadmap, Phase 1: Monolith first). When the formula service is split off
// into its own process, swap this for the gRPC-backed adapter without
// touching the evaluator or the application layer.
type RepositoryDataPort struct {
	bars       marketbar.Repository
	financials financial.Repository
	factors    factor.Repository
	adjuster   domainAdj.Adjuster

	// Adjustment defaults to AdjustmentPre to match the Market Data service
	// default; callers can override per RepositoryDataPort instance.
	adjustment valueobject.Adjustment

	// LookbackPadding is added to req.LookbackBars when calling the bar
	// repository, so the rolling-window indicators inside the evaluator have
	// room to warm up. 30 trading days covers the majority of MVP MAs.
	lookbackPadding int
}

// RepositoryConfig configures a RepositoryDataPort.
//
// Adjuster is optional; when nil, raw bars are returned (which only matters
// for backtests that explicitly want unadjusted prices).
type RepositoryConfig struct {
	Bars       marketbar.Repository
	Financials financial.Repository
	Factors    factor.Repository
	Adjuster   domainAdj.Adjuster

	Adjustment      valueobject.Adjustment
	LookbackPadding int
}

// NewRepository builds a RepositoryDataPort. Bars and Factors are required;
// Financials is optional and may be nil for stripped-down deployments.
func NewRepository(c RepositoryConfig) (*RepositoryDataPort, error) {
	if c.Bars == nil {
		return nil, fmt.Errorf("repository data port: bars repository is required")
	}
	if c.Factors == nil {
		return nil, fmt.Errorf("repository data port: factors repository is required")
	}
	adjustment := c.Adjustment
	if adjustment == "" {
		adjustment = valueobject.AdjustmentPre
	}
	padding := c.LookbackPadding
	if padding < 0 {
		padding = 0
	}
	if padding == 0 {
		padding = 30
	}
	return &RepositoryDataPort{
		bars:            c.Bars,
		financials:      c.Financials,
		factors:         c.Factors,
		adjuster:        c.Adjuster,
		adjustment:      adjustment,
		lookbackPadding: padding,
	}, nil
}

// LoadBars implements domainEval.DataPort.
//
// We issue one Range query per stock. The market repository already enforces
// per-stock indexing; batching across stocks would require schema-level
// changes, which the MVP does not justify. Concurrency is bounded so a
// thousand-stock universe does not flood the connection pool.
func (p *RepositoryDataPort) LoadBars(ctx context.Context, req domainEval.BarsRequest) (map[string][]series.Bar, error) {
	if len(req.StockCodes) == 0 {
		return map[string][]series.Bar{}, nil
	}

	rng := p.barsRange(req)
	limit := req.LookbackBars + p.lookbackPadding
	if limit < 60 {
		limit = 60
	}

	out := make(map[string][]series.Bar, len(req.StockCodes))
	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		errCh = make(chan error, 1)
		sem   = make(chan struct{}, loadConcurrency)
	)

	for _, code := range req.StockCodes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		stockCode := code
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			bars, err := p.loadStockBars(ctx, stockCode, rng, limit, req.DataVersion)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("load bars for %s: %w", stockCode, err):
				default:
				}
				return
			}
			mu.Lock()
			out[stockCode] = bars
			mu.Unlock()
		}()
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return out, nil
}

func (p *RepositoryDataPort) loadStockBars(
	ctx context.Context,
	stockCode string,
	rng valueobject.DateRange,
	limit int,
	version string,
) ([]series.Bar, error) {
	rawBars, err := p.bars.Range(ctx, marketbar.RangeQuery{
		StockCode:   stockCode,
		Period:      valueobject.PeriodDay,
		Range:       rng,
		DataVersion: version,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	if len(rawBars) == 0 {
		return nil, nil
	}

	if p.adjuster != nil && p.adjustment != valueobject.AdjustmentNone {
		adjusted, err := p.adjuster.Apply(ctx, rawBars, p.adjustment)
		if err != nil {
			return nil, fmt.Errorf("apply adjustment: %w", err)
		}
		rawBars = adjusted
	}

	out := make([]series.Bar, len(rawBars))
	for i, b := range rawBars {
		out[i] = series.Bar{
			Timestamp: b.TradeDate,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    float64(b.Volume),
			Amount:    b.Amount,
		}
	}
	return out, nil
}

func (p *RepositoryDataPort) barsRange(req domainEval.BarsRequest) valueobject.DateRange {
	end := req.AsOfDate
	if end.IsZero() {
		end = time.Now().UTC()
	}
	if req.LookbackBars <= 0 {
		return valueobject.DateRange{End: end}
	}
	// Pad lookback with weekends/holidays. Trading days per calendar day is
	// ~5/7, so we multiply by 2 and add a few days of safety; the Limit on
	// the SQL query keeps the result bounded.
	calendarDays := (req.LookbackBars+p.lookbackPadding)*2 + 7
	start := end.AddDate(0, 0, -calendarDays)
	return valueobject.DateRange{Start: start, End: end}
}

// loadConcurrency caps the number of parallel repository requests we issue.
// The default value matches the MaxOpenConns budget for a single service.
const loadConcurrency = 8

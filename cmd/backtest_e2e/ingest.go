package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	"github.com/agoXQ/QuantLab/app/market/domain/provider"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// runIngest drives the Market Data ingestion service over the scenario's
// universe. It is idempotent at the storage layer (BulkUpsert) so reruns
// against the same range are safe.
//
// The function returns the DataVersion the ingestion ran under so the
// caller can pin the backtest to the same snapshot. Returning an empty
// string means ingestion was skipped.
func runIngest(ctx context.Context, sc *Scenario, ing appMarket.IngestionService) (string, error) {
	if sc.Ingest.Skip {
		log.Printf("[ingest] skipped by scenario")
		return "", nil
	}
	if ing == nil {
		return "", fmt.Errorf("ingestion service is not configured")
	}
	market, err := sc.IngestMarket()
	if err != nil {
		return "", err
	}

	desc := strings.TrimSpace(sc.Ingest.Description)
	if desc == "" {
		desc = "backtest_e2e " + sc.Name
	}
	dv, err := ing.CreateVersion(ctx, desc)
	if err != nil {
		return "", fmt.Errorf("create data version: %w", err)
	}
	log.Printf("[ingest] data_version=%s", dv.Version)

	if err := ingestSecurities(ctx, ing, market); err != nil {
		return "", err
	}
	if err := ingestCalendar(ctx, sc, ing, market); err != nil {
		return "", err
	}
	if err := ingestBars(ctx, sc, ing, dv); err != nil {
		return "", err
	}
	if err := ingestFinancials(ctx, sc, ing, dv); err != nil {
		return "", err
	}
	if err := ingestFactors(ctx, sc, ing, dv); err != nil {
		return "", err
	}

	return dv.Version, nil
}

func ingestSecurities(ctx context.Context, ing appMarket.IngestionService, m marketVO.Market) error {
	n, err := ing.IngestSecurities(ctx, m)
	if err != nil {
		return fmt.Errorf("ingest securities: %w", err)
	}
	log.Printf("[ingest] securities=%d", n)
	return nil
}

func ingestCalendar(ctx context.Context, sc *Scenario, ing appMarket.IngestionService, m marketVO.Market) error {
	if sc.Ingest.Calendar == nil {
		return nil
	}
	rng, err := sc.Ingest.Calendar.Range.Parse()
	if err != nil {
		return fmt.Errorf("calendar range: %w", err)
	}
	n, err := ing.IngestCalendar(ctx, m, rng)
	if err != nil {
		return fmt.Errorf("ingest calendar: %w", err)
	}
	log.Printf("[ingest] calendar=%d", n)
	return nil
}

func ingestBars(ctx context.Context, sc *Scenario, ing appMarket.IngestionService, dv *dataversion.DataVersion) error {
	if sc.Ingest.Bars == nil {
		return nil
	}
	period, err := marketVO.ParsePeriod(sc.Ingest.Bars.Period)
	if err != nil {
		return fmt.Errorf("parse bars.period: %w", err)
	}
	rng, err := sc.Ingest.Bars.Range.Parse()
	if err != nil {
		return fmt.Errorf("bars range: %w", err)
	}
	total := 0
	for _, code := range sc.Backtest.Universe {
		n, err := ing.IngestBars(ctx, provider.BarQuery{
			StockCode: code,
			Period:    period,
			Range:     rng,
		}, dv.Version)
		if err != nil {
			return fmt.Errorf("ingest bars %s: %w", code, err)
		}
		total += n
	}
	log.Printf("[ingest] bars=%d (period=%s)", total, period)
	return nil
}

func ingestFinancials(ctx context.Context, sc *Scenario, ing appMarket.IngestionService, dv *dataversion.DataVersion) error {
	if sc.Ingest.Financials == nil {
		return nil
	}
	reportType := marketVO.ReportType(strings.ToLower(strings.TrimSpace(sc.Ingest.Financials.ReportType)))
	if reportType == "" {
		reportType = marketVO.ReportAnnual
	}
	if !reportType.IsValid() {
		return fmt.Errorf("financials.report_type %q invalid", sc.Ingest.Financials.ReportType)
	}
	rng, err := sc.Ingest.Financials.Range.Parse()
	if err != nil {
		return fmt.Errorf("financials range: %w", err)
	}
	total := 0
	for _, code := range sc.Backtest.Universe {
		n, err := ing.IngestFinancials(ctx, provider.FinancialQuery{
			StockCode:  code,
			ReportType: reportType,
			Range:      rng,
		}, dv.Version)
		if err != nil {
			return fmt.Errorf("ingest financials %s: %w", code, err)
		}
		total += n
	}
	log.Printf("[ingest] financials=%d", total)
	return nil
}

func ingestFactors(ctx context.Context, sc *Scenario, ing appMarket.IngestionService, dv *dataversion.DataVersion) error {
	if sc.Ingest.Factors == nil {
		return nil
	}
	rng, err := sc.Ingest.Factors.Range.Parse()
	if err != nil {
		return fmt.Errorf("factors range: %w", err)
	}
	names := sc.Ingest.Factors.Names
	total := 0
	for _, code := range sc.Backtest.Universe {
		n, err := ing.IngestFactors(ctx, provider.FactorQuery{
			StockCode:   code,
			FactorNames: names,
			Range:       rng,
		}, dv.Version)
		if err != nil {
			return fmt.Errorf("ingest factors %s: %w", code, err)
		}
		total += n
	}
	log.Printf("[ingest] factors=%d", total)
	return nil
}

package svc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraMarketAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	infraMarketPg "github.com/agoXQ/QuantLab/app/market/infrastructure/postgres"

	infraDataport "github.com/agoXQ/QuantLab/app/formula/infrastructure/dataport"

	"github.com/agoXQ/QuantLab/app/backtest/internal/config"
	infraMarketData "github.com/agoXQ/QuantLab/app/backtest/infrastructure/marketdata"
)

// marketStack groups the optional data-access pieces that depend on the
// Market Data Postgres database. When MarketData.DSN is empty (CI smoke
// tests, local exploration without the platform DB), every field on the
// returned stack is nil and the caller falls back to the in-memory
// adapters.
//
// Both Backtest and Formula need the same underlying tables, so we build
// one Market application service and one RepositoryDataPort and let each
// consumer pick the surface it cares about:
//
//   * Backtest's MarketData.Provider goes through FromMarketService so
//     the engine reuses the calendar / adjustment plumbing inside
//     market.Service.
//   * Formula's evaluator.DataPort goes through RepositoryDataPort so the
//     in-process formula stack reads the same bars without taking a
//     dependency on the application package.
type marketStack struct {
	db        *sql.DB
	provider  *infraMarketData.FromMarketService
	dataPort  *infraDataport.RepositoryDataPort
}

// buildMarketStack opens the Market Data database, wires the seven
// repositories Market Data ships with, and returns the Backtest provider
// plus the Formula data port. Returning a zero stack and nil error means
// the caller should keep using the in-memory adapters; a non-nil error
// signals a configuration mistake worth surfacing.
func buildMarketStack(c config.MarketDataConfig) (marketStack, error) {
	if c.DSN == "" {
		return marketStack{}, nil
	}

	db, err := openMarketDataDB(c)
	if err != nil {
		return marketStack{}, err
	}

	deps := appMarket.Dependencies{
		Securities:   infraMarketPg.NewSecurityRepository(db),
		Bars:         infraMarketPg.NewMarketBarRepository(db),
		Financials:   infraMarketPg.NewFinancialRepository(db),
		Factors:      infraMarketPg.NewFactorRepository(db),
		Indexes:      infraMarketPg.NewIndexBarRepository(db),
		Calendar:     infraMarketPg.NewCalendarRepository(db),
		DataVersions: infraMarketPg.NewDataVersionRepository(db),
		Adjuster:     infraMarketAdj.NewFactorAdjuster(),
	}

	adjustment := marketVO.Adjustment(c.Adjustment)
	if !adjustment.IsValid() {
		adjustment = marketVO.AdjustmentPre
	}

	svc := appMarket.NewService(deps)
	provider := infraMarketData.NewFromMarketService(svc, adjustment)

	dataPort, err := infraDataport.NewRepository(infraDataport.RepositoryConfig{
		Bars:       deps.Bars,
		Financials: deps.Financials,
		Factors:    deps.Factors,
		Adjuster:   deps.Adjuster,
		Adjustment: adjustment,
	})
	if err != nil {
		_ = db.Close()
		return marketStack{}, fmt.Errorf("build repository data port: %w", err)
	}

	log.Printf("[backtest] market data stack wired (adjustment=%s)", adjustment)
	return marketStack{db: db, provider: provider, dataPort: dataPort}, nil
}

func openMarketDataDB(c config.MarketDataConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open market data db: %w", err)
	}
	if c.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.MaxIdleConns)
	}
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping market data db: %w", err)
	}
	return db, nil
}

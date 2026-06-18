// Command backtest_e2e is the end-to-end smoke harness for the Backtest
// engine. The binary glues Tushare ingestion, the Market Data Postgres
// repositories, and a single backtest replay together so a regression in
// any layer fails one CI step. See README.md for the full story.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
	"github.com/agoXQ/QuantLab/app/market/infrastructure/env"
)

func main() {
	scenarioPath := flag.String("scenario", "", "path to a scenario YAML file (required)")
	dsn := flag.String("dsn", "", "Market Data Postgres DSN (defaults to MARKET_DATA_DSN env)")
	skipIngest := flag.Bool("skip-ingest", false, "do not call Tushare; assume the DB is already populated")
	updateBaseline := flag.Bool("update-baseline", false, "overwrite the baseline JSON instead of failing on drift")
	flag.Parse()

	if strings.TrimSpace(*scenarioPath) == "" {
		fail("missing -scenario flag")
	}
	loadDotenv()

	sc, err := LoadScenario(*scenarioPath)
	if err != nil {
		fail(err.Error())
	}
	if *skipIngest {
		sc.Ingest.Skip = true
	}
	log.Printf("[scenario] %s formula=%q universe=%v", sc.Name, sc.Backtest.Formula, sc.Backtest.Universe)

	cfg := platformConfig{
		dsn:          firstNonEmpty(*dsn, os.Getenv("MARKET_DATA_DSN")),
		tushareToken: os.Getenv("TUSHARE_TOKEN"),
	}
	if sc.Ingest.Skip {
		cfg.tushareToken = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	pf, err := buildPlatform(ctx, cfg)
	if err != nil {
		fail(err.Error())
	}
	defer func() {
		if pf != nil && pf.close != nil {
			_ = pf.close()
		}
	}()

	dataVersion, err := runIngest(ctx, sc, pf.ingest)
	if err != nil {
		fail("ingest: " + err.Error())
	}

	res, err := runBacktest(ctx, pf, sc, dataVersion)
	if err != nil {
		fail("run: " + err.Error())
	}
	if res.Report == nil {
		fail("backtest produced no report")
	}
	printReport(res.Report)

	os.Exit(handleBaseline(*scenarioPath, sc, res.Report, *updateBaseline))
}

// handleBaseline writes a fresh baseline when none exists, otherwise
// diffs against it and returns a non-zero exit code on drift.
func handleBaseline(scenarioPath string, sc *Scenario, rep *report.PerformanceReport, force bool) int {
	baselinePath := BaselinePath(scenarioPath)
	current := snapshotFrom(rep, sc)

	prev, err := readBaseline(baselinePath)
	if err != nil {
		log.Printf("[baseline] read failed: %v", err)
		return 2
	}
	if prev == nil || force {
		if err := writeBaseline(baselinePath, &current); err != nil {
			log.Printf("[baseline] write failed: %v", err)
			return 2
		}
		if prev == nil {
			log.Printf("[baseline] wrote initial baseline at %s", baselinePath)
		} else {
			log.Printf("[baseline] forced overwrite of %s", baselinePath)
		}
		return 0
	}

	passed, rows := compareMetrics(prev, &current, sc.Baseline.Tolerance)
	fmt.Println()
	fmt.Println("baseline diff (baseline -> current, tolerance):")
	for _, r := range rows {
		mark := "ok"
		if !r.pass {
			mark = "FAIL"
		}
		fmt.Printf("  %-14s %12.6f -> %-12.6f tol=%-10.6f %s\n", r.field, r.baseline, r.current, r.tolerance, mark)
	}
	if !passed {
		fmt.Println()
		fmt.Println("baseline diff exceeded tolerance; rerun with -update-baseline to accept the new metrics")
		return 1
	}
	log.Printf("[baseline] within tolerance (%s)", baselinePath)
	return 0
}

// printReport summarises the run for human readers. The numbers are the
// same ones the baseline pins.
func printReport(rep *report.PerformanceReport) {
	fmt.Println()
	fmt.Println("performance report:")
	fmt.Printf("  range          %s -> %s\n", rep.StartDate.Format("2006-01-02"), rep.EndDate.Format("2006-01-02"))
	fmt.Printf("  initial_capital %12.2f\n", rep.InitialCapital)
	fmt.Printf("  final_asset     %12.2f\n", rep.FinalAsset)
	fmt.Printf("  total_return    %12.6f\n", rep.TotalReturn)
	fmt.Printf("  annual_return   %12.6f\n", rep.AnnualReturn)
	fmt.Printf("  volatility      %12.6f\n", rep.Volatility)
	fmt.Printf("  sharpe_ratio    %12.6f\n", rep.SharpeRatio)
	fmt.Printf("  max_drawdown    %12.6f\n", rep.MaxDrawdown)
	fmt.Printf("  win_rate        %12.6f\n", rep.WinRate)
	fmt.Printf("  trade_count     %12d\n", rep.TradeCount)
}

// loadDotenv loads .env from common locations so the harness picks up
// TUSHARE_TOKEN / MARKET_DATA_DSN without an explicit shell export.
func loadDotenv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd+"/.env")
	}
	if err := env.Load(candidates...); err != nil {
		log.Printf("[env] load: %v", err)
	}
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "backtest_e2e: %s\n", msg)
	os.Exit(2)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

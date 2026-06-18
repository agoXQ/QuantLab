package tests

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	infraPg "github.com/agoXQ/QuantLab/app/backtest/infrastructure/repository/postgres"
)

// pgFixture lazily opens a postgres connection backed by the BACKTEST_TEST_DSN
// environment variable. When the DSN is empty or the connection fails the
// caller skips the test rather than failing CI on machines without docker.
type pgFixture struct {
	db *sql.DB
}

func openPg(t *testing.T) *pgFixture {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("BACKTEST_TEST_DSN"))
	if dsn == "" {
		t.Skip("BACKTEST_TEST_DSN not set; skipping postgres repo tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("open postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("ping postgres: %v", err)
	}
	if err := infraPg.EnsureSchema(ctx, db); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// Truncate the per-test surface so reruns stay deterministic. CASCADE
	// is required because backtest_order / backtest_trade / snapshots /
	// reports all reference backtest_job.
	cleanCtx, cancelClean := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelClean()
	if _, err := db.ExecContext(cleanCtx, `TRUNCATE backtest_job, backtest_order, backtest_trade,
		backtest_portfolio_snapshot, backtest_report RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return &pgFixture{db: db}
}

func TestPostgres_JobCRUD(t *testing.T) {
	fx := openPg(t)
	repo := infraPg.NewJobRepository(fx.db)
	ctx := context.Background()
	now := time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC)

	job := &backtestjob.BacktestJob{
		UserID:         42,
		StrategyID:     7,
		Formula:        "ROE > 15",
		Universe:       []string{"000001", "000002"},
		Benchmark:      "000300",
		DataVersion:    "v1",
		InitialCapital: 1_000_000,
		Range: valueobject.DateRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		Config:    backtestjob.DefaultConfig(),
		Status:    valueobject.JobStatusCreated,
		CreatedAt: now,
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}
	if job.ID == 0 {
		t.Fatal("expected generated id")
	}

	if err := job.MarkRunning(now.Add(time.Minute)); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := repo.Update(ctx, job); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != valueobject.JobStatusRunning {
		t.Errorf("expected RUNNING, got %s", got.Status)
	}
	if got.UserID != 42 || got.StrategyID != 7 {
		t.Errorf("ids round-tripped wrong: %+v", got)
	}
	if len(got.Universe) != 2 || got.Universe[0] != "000001" {
		t.Errorf("universe round-trip mismatch: %v", got.Universe)
	}
	if got.Config.MaxPositionCount == 0 {
		t.Errorf("expected default config to round-trip, got %+v", got.Config)
	}

	listed, err := repo.List(ctx, backtestjob.ListQuery{UserID: 42, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one job, got %d", len(listed))
	}
}

func TestPostgres_OrdersTradesSnapshotsReport(t *testing.T) {
	fx := openPg(t)
	jobs := infraPg.NewJobRepository(fx.db)
	orderRepo := infraPg.NewOrderRepository(fx.db)
	tradeRepo := infraPg.NewTradeRepository(fx.db)
	portRepo := infraPg.NewPortfolioRepository(fx.db)
	reportRepo := infraPg.NewReportRepository(fx.db)
	ctx := context.Background()
	now := time.Date(2024, 5, 6, 9, 30, 0, 0, time.UTC)

	job := &backtestjob.BacktestJob{
		Formula:        "ROE > 0",
		Universe:       []string{"000001"},
		InitialCapital: 100_000,
		Range: valueobject.DateRange{
			Start: time.Date(2024, 5, 6, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 5, 7, 0, 0, 0, 0, time.UTC),
		},
		Config:    backtestjob.DefaultConfig(),
		Status:    valueobject.JobStatusRunning,
		CreatedAt: now,
	}
	if err := jobs.Create(ctx, job); err != nil {
		t.Fatalf("create job: %v", err)
	}

	ord := &order.Order{
		JobID:       job.ID,
		StockCode:   "000001",
		Side:        valueobject.OrderSideBuy,
		Quantity:    100,
		Status:      valueobject.OrderStatusFilled,
		SubmittedAt: now,
	}
	if err := orderRepo.BulkInsert(ctx, []*order.Order{ord}); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if ord.ID == 0 {
		t.Fatal("expected order id")
	}
	listedOrders, err := orderRepo.ListByJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(listedOrders) != 1 || listedOrders[0].StockCode != "000001" {
		t.Fatalf("orders mismatch: %+v", listedOrders)
	}

	tr := &trade.Trade{
		JobID:      job.ID,
		OrderID:    ord.ID,
		StockCode:  "000001",
		Side:       valueobject.OrderSideBuy,
		Quantity:   100,
		Price:      12.5,
		Commission: 5,
		StampDuty:  0,
		Slippage:   0.01,
		TradeTime:  now.Add(time.Minute),
	}
	if err := tradeRepo.BulkInsert(ctx, []*trade.Trade{tr}); err != nil {
		t.Fatalf("insert trade: %v", err)
	}
	listedTrades, err := tradeRepo.ListByJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("list trades: %v", err)
	}
	if len(listedTrades) != 1 || listedTrades[0].Price != 12.5 {
		t.Fatalf("trades mismatch: %+v", listedTrades)
	}

	snap := portfolio.Snapshot{
		JobID:       job.ID,
		TradeDate:   time.Date(2024, 5, 6, 0, 0, 0, 0, time.UTC),
		Cash:        99_000,
		MarketValue: 1_000,
		TotalAsset:  100_000,
		Positions: []portfolio.Position{{
			StockCode: "000001", Quantity: 100, CostPrice: 10, MarketPrice: 10, MarketValue: 1000,
		}},
	}
	if err := portRepo.BulkInsertSnapshots(ctx, []portfolio.Snapshot{snap}); err != nil {
		t.Fatalf("insert snapshot: %v", err)
	}
	// Upsert path: rerun with the same trade_date and assert overwrite.
	snap.Cash = 98_500
	if err := portRepo.BulkInsertSnapshots(ctx, []portfolio.Snapshot{snap}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}
	listedSnaps, err := portRepo.ListSnapshots(ctx, job.ID)
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(listedSnaps) != 1 || listedSnaps[0].Cash != 98_500 {
		t.Fatalf("snapshots mismatch: %+v", listedSnaps)
	}
	if len(listedSnaps[0].Positions) != 1 || listedSnaps[0].Positions[0].Quantity != 100 {
		t.Fatalf("positions json round-trip mismatch: %+v", listedSnaps[0].Positions)
	}

	rep := &report.PerformanceReport{
		JobID:          job.ID,
		StartDate:      job.Range.Start,
		EndDate:        job.Range.End,
		InitialCapital: 100_000,
		FinalAsset:     105_000,
		TotalReturn:    0.05,
		AnnualReturn:   0.45,
		Volatility:     0.18,
		SharpeRatio:    1.5,
		MaxDrawdown:    0.03,
		WinRate:        0.6,
		TradeCount:     1,
		EquityCurve: []report.EquityPoint{{
			TradeDate: snap.TradeDate, TotalAsset: 100_000, Drawdown: 0, Return: 0,
		}},
		GeneratedAt: now,
	}
	if err := reportRepo.Save(ctx, rep); err != nil {
		t.Fatalf("save report: %v", err)
	}
	got, err := reportRepo.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if got.SharpeRatio != 1.5 || got.TradeCount != 1 {
		t.Errorf("report round-trip mismatch: %+v", got)
	}
	if len(got.EquityCurve) != 1 {
		t.Errorf("equity curve length mismatch: %+v", got.EquityCurve)
	}

	// Save again with new numbers to exercise the upsert branch.
	rep.SharpeRatio = 2.0
	if err := reportRepo.Save(ctx, rep); err != nil {
		t.Fatalf("save report (update): %v", err)
	}
	got, err = reportRepo.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("get report (update): %v", err)
	}
	if got.SharpeRatio != 2.0 {
		t.Errorf("expected updated sharpe 2.0, got %v", got.SharpeRatio)
	}
}

package tests

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
	infraPg "github.com/agoXQ/QuantLab/app/strategy/infrastructure/repository/postgres"
)

// pgFixture lazily opens a postgres connection backed by the
// STRATEGY_TEST_DSN environment variable. When the DSN is empty or the
// connection fails the caller skips the test rather than failing CI on
// machines without docker. Mirrors the Backtest service's pattern so
// the rule is consistent across services.
type pgFixture struct {
	db *sql.DB
}

func openPg(t *testing.T) *pgFixture {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("STRATEGY_TEST_DSN"))
	if dsn == "" {
		t.Skip("STRATEGY_TEST_DSN not set; skipping postgres repo tests")
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
	cleanCtx, cancelClean := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelClean()
	if _, err := db.ExecContext(cleanCtx,
		`TRUNCATE strategy_fork, strategy_version, strategy RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return &pgFixture{db: db}
}

func TestPostgres_StrategyCRUDAndIncrement(t *testing.T) {
	fx := openPg(t)
	repo := infraPg.NewStrategyRepository(fx.db)
	ctx := context.Background()
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)

	st := &domstrategy.Strategy{
		AuthorID:   42,
		Title:      "PG roundtrip",
		Tags:       []string{"alpha", "factor"},
		Status:     valueobject.LifecycleStatusDraft,
		Visibility: valueobject.VisibilityPrivate,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.Create(ctx, st); err != nil {
		t.Fatalf("create: %v", err)
	}
	if st.ID == 0 {
		t.Fatal("expected generated id")
	}

	got, err := repo.Get(ctx, st.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.AuthorID != 42 || got.Title != "PG roundtrip" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "alpha" {
		t.Errorf("tags roundtrip: %v", got.Tags)
	}

	// Update flips visibility
	got.Visibility = valueobject.VisibilityPublic
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	again, err := repo.Get(ctx, st.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if again.Visibility != valueobject.VisibilityPublic {
		t.Errorf("expected PUBLIC, got %s", again.Visibility)
	}

	// IncrementForkCount
	if err := repo.IncrementForkCount(ctx, st.ID); err != nil {
		t.Fatalf("increment: %v", err)
	}
	if err := repo.IncrementForkCount(ctx, st.ID); err != nil {
		t.Fatalf("increment 2: %v", err)
	}
	final, err := repo.Get(ctx, st.ID)
	if err != nil {
		t.Fatalf("get final: %v", err)
	}
	if final.ForkCount != 2 {
		t.Errorf("expected ForkCount=2, got %d", final.ForkCount)
	}

	listed, err := repo.List(ctx, domstrategy.ListQuery{AuthorID: 42, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one row, got %d", len(listed))
	}

	if _, err := repo.Get(ctx, 999999); err != stratErr.ErrStrategyNotFound {
		t.Errorf("expected ErrStrategyNotFound, got %v", err)
	}
}

func TestPostgres_VersionAppendOnly(t *testing.T) {
	fx := openPg(t)
	stRepo := infraPg.NewStrategyRepository(fx.db)
	verRepo := infraPg.NewVersionRepository(fx.db)
	ctx := context.Background()
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)

	st := &domstrategy.Strategy{
		AuthorID:   1,
		Title:      "version host",
		Status:     valueobject.LifecycleStatusDraft,
		Visibility: valueobject.VisibilityPrivate,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := stRepo.Create(ctx, st); err != nil {
		t.Fatalf("create strategy: %v", err)
	}

	for i := 1; i <= 3; i++ {
		v := &domversion.StrategyVersion{
			StrategyID:  st.ID,
			VersionNo:   "v" + intToStr(i),
			FormulaText: "ROE>0",
			CreatedBy:   1,
			CreatedAt:   now.Add(time.Duration(i) * time.Minute),
		}
		if err := verRepo.Create(ctx, v); err != nil {
			t.Fatalf("create version %d: %v", i, err)
		}
	}
	latest, err := verRepo.LatestNumber(ctx, st.ID)
	if err != nil {
		t.Fatalf("latest number: %v", err)
	}
	if latest != 3 {
		t.Errorf("expected latest=3, got %d", latest)
	}
	listed, err := verRepo.ListByStrategy(ctx, st.ID, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(listed))
	}
	// Newest first.
	if listed[0].VersionNo != "v3" {
		t.Errorf("expected v3 first, got %s", listed[0].VersionNo)
	}
}

func TestPostgres_ForkLedger(t *testing.T) {
	fx := openPg(t)
	stRepo := infraPg.NewStrategyRepository(fx.db)
	forkRepo := infraPg.NewForkRepository(fx.db)
	ctx := context.Background()
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)

	src := &domstrategy.Strategy{AuthorID: 1, Title: "src", Status: valueobject.LifecycleStatusPublished, Visibility: valueobject.VisibilityPublic, CreatedAt: now, UpdatedAt: now}
	target := &domstrategy.Strategy{AuthorID: 2, Title: "target", Status: valueobject.LifecycleStatusDraft, Visibility: valueobject.VisibilityPrivate, CreatedAt: now, UpdatedAt: now}
	if err := stRepo.Create(ctx, src); err != nil {
		t.Fatalf("create src: %v", err)
	}
	if err := stRepo.Create(ctx, target); err != nil {
		t.Fatalf("create target: %v", err)
	}
	fork := &domfork.StrategyFork{SourceStrategyID: src.ID, TargetStrategyID: target.ID, CreatorID: 2, CreatedAt: now}
	if err := forkRepo.Create(ctx, fork); err != nil {
		t.Fatalf("create fork: %v", err)
	}
	if fork.ID == 0 {
		t.Fatal("expected fork id assigned")
	}
	listed, err := forkRepo.ListBySource(ctx, src.ID, 10)
	if err != nil {
		t.Fatalf("list by source: %v", err)
	}
	if len(listed) != 1 || listed[0].TargetStrategyID != target.ID {
		t.Fatalf("fork ledger mismatch: %+v", listed)
	}
}

func intToStr(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return ""
}

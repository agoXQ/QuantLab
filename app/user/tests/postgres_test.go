package tests

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
	infraPg "github.com/agoXQ/QuantLab/app/user/infrastructure/repository/postgres"
)

// openPg lazily opens a postgres connection backed by USER_TEST_DSN.
// Empty / unreachable DSN skips the test rather than failing CI on
// machines without docker. Mirrors the Strategy / Backtest pattern.
func openPg(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("USER_TEST_DSN"))
	if dsn == "" {
		t.Skip("USER_TEST_DSN not set; skipping postgres repo tests")
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
		`TRUNCATE user_follow, app_user RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return db
}

func TestPostgres_UserCRUD(t *testing.T) {
	db := openPg(t)
	repo := infraPg.NewUserRepository(db)
	ctx := context.Background()
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)

	u := &domuser.User{
		Username:       "alice",
		Email:          "alice@example.com",
		PasswordHash:   "hash-1",
		Status:         valueobject.AccountStatusActive,
		MembershipTier: valueobject.MembershipTierFree,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected generated id")
	}

	got, err := repo.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Username != "alice" || got.Email != "alice@example.com" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}

	got.Bio = "moved to Tokyo"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	again, err := repo.GetByEmail(ctx, "ALICE@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if again.Bio != "moved to Tokyo" {
		t.Errorf("expected bio updated, got %q", again.Bio)
	}

	dup := &domuser.User{
		Username:       "alice2",
		Email:          "alice@example.com",
		PasswordHash:   "hash-2",
		Status:         valueobject.AccountStatusActive,
		MembershipTier: valueobject.MembershipTierFree,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.Create(ctx, dup); !errors.Is(err, userErr.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}

	if _, err := repo.Get(ctx, 999999); !errors.Is(err, userErr.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPostgres_FollowFlow(t *testing.T) {
	db := openPg(t)
	users := infraPg.NewUserRepository(db)
	follows := infraPg.NewFollowRepository(db)
	ctx := context.Background()
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)

	a := &domuser.User{
		Username: "alice", Email: "alice@example.com", PasswordHash: "h",
		Status: valueobject.AccountStatusActive, MembershipTier: valueobject.MembershipTierFree,
		CreatedAt: now, UpdatedAt: now,
	}
	b := &domuser.User{
		Username: "bob", Email: "bob@example.com", PasswordHash: "h",
		Status: valueobject.AccountStatusActive, MembershipTier: valueobject.MembershipTierFree,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := users.Create(ctx, a); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := users.Create(ctx, b); err != nil {
		t.Fatalf("create b: %v", err)
	}

	if err := follows.Create(ctx, &domfollow.Follow{FollowerID: a.ID, FolloweeID: b.ID, CreatedAt: now}); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if err := follows.Create(ctx, &domfollow.Follow{FollowerID: a.ID, FolloweeID: b.ID, CreatedAt: now}); !errors.Is(err, userErr.ErrAlreadyFollowed) {
		t.Fatalf("expected ErrAlreadyFollowed, got %v", err)
	}

	exists, err := follows.Exists(ctx, a.ID, b.ID)
	if err != nil || !exists {
		t.Fatalf("exists: ok=%v err=%v", exists, err)
	}
	bFollowers, err := follows.CountFollowers(ctx, b.ID)
	if err != nil {
		t.Fatalf("count followers: %v", err)
	}
	if bFollowers != 1 {
		t.Errorf("expected 1 follower, got %d", bFollowers)
	}

	rows, err := follows.ListFollowers(ctx, b.ID, 10, 0)
	if err != nil {
		t.Fatalf("list followers: %v", err)
	}
	if len(rows) != 1 || rows[0].FollowerID != a.ID {
		t.Fatalf("unexpected followers list: %+v", rows)
	}

	if err := follows.Delete(ctx, a.ID, b.ID); err != nil {
		t.Fatalf("unfollow: %v", err)
	}
	if err := follows.Delete(ctx, a.ID, b.ID); !errors.Is(err, userErr.ErrFollowNotFound) {
		t.Fatalf("expected ErrFollowNotFound, got %v", err)
	}
}

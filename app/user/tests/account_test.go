package tests

import (
	"context"
	"errors"
	"testing"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

// TestChangePassword walks the canonical happy path + the two
// rejection branches we care about: wrong current password and
// new password identical to the current one.
func TestChangePassword(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	id := registerHelper(t, fx, "alice", "alice@example.com")

	if err := fx.svc.ChangePassword(ctx, appUser.ChangePasswordRequest{
		UserID:          id,
		CurrentPassword: "wrong-password",
		NewPassword:     "newpassword1",
	}); !errors.Is(err, userErr.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if err := fx.svc.ChangePassword(ctx, appUser.ChangePasswordRequest{
		UserID:          id,
		CurrentPassword: "hunter2hunter",
		NewPassword:     "hunter2hunter",
	}); !errors.Is(err, userErr.ErrSamePassword) {
		t.Fatalf("expected ErrSamePassword, got %v", err)
	}

	if err := fx.svc.ChangePassword(ctx, appUser.ChangePasswordRequest{
		UserID:          id,
		CurrentPassword: "hunter2hunter",
		NewPassword:     "newpassword1",
	}); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	// Old password no longer works.
	if _, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email: "alice@example.com", Password: "hunter2hunter",
	}); !errors.Is(err, userErr.ErrInvalidCredentials) {
		t.Fatalf("expected old password to fail, got %v", err)
	}
	// New password works.
	if _, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email: "alice@example.com", Password: "newpassword1",
	}); err != nil {
		t.Fatalf("Login with new password: %v", err)
	}
}

// TestRefreshTokenHappyPath confirms a freshly issued refresh token
// can rotate into a new pair, and an empty / unknown token surfaces
// ErrTokenInvalid.
func TestRefreshTokenHappyPath(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	res, err := fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice", Email: "alice@example.com", Password: "hunter2hunter",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	rotated, err := fx.svc.RefreshToken(ctx, appUser.RefreshTokenRequest{RefreshToken: res.RefreshToken})
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if rotated.AccessToken == "" || rotated.RefreshToken == "" {
		t.Fatalf("expected new pair, got %+v", rotated)
	}
	if rotated.User.ID != res.User.ID {
		t.Fatalf("expected same user id, got %d vs %d", rotated.User.ID, res.User.ID)
	}

	if _, err := fx.svc.RefreshToken(ctx, appUser.RefreshTokenRequest{RefreshToken: "  "}); !errors.Is(err, userErr.ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

// TestUpdateAccount applies the moderator-style patch and confirms
// the row reflects the new tier + status.
func TestUpdateAccount(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	id := registerHelper(t, fx, "alice", "alice@example.com")

	tier := valueobject.MembershipTierPro
	status := valueobject.AccountStatusSuspended
	updated, err := fx.svc.UpdateAccount(ctx, appUser.UpdateAccountRequest{
		UserID:         id,
		MembershipTier: &tier,
		Status:         &status,
	})
	if err != nil {
		t.Fatalf("UpdateAccount: %v", err)
	}
	if updated.MembershipTier != valueobject.MembershipTierPro {
		t.Fatalf("expected PRO tier, got %s", updated.MembershipTier)
	}
	if updated.Status != valueobject.AccountStatusSuspended {
		t.Fatalf("expected SUSPENDED, got %d", updated.Status)
	}

	// Login refuses suspended accounts.
	if _, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email: "alice@example.com", Password: "hunter2hunter",
	}); !errors.Is(err, userErr.ErrAccountSuspended) {
		t.Fatalf("expected ErrAccountSuspended, got %v", err)
	}
}

// TestStrategyAndBacktestCounters drives the IncrementStrategyCount /
// IncrementBacktestCount hooks so the profile counters reflect cross-
// service activity. The repository clamps at zero so a stray
// "deleted" event cannot push the value negative.
func TestStrategyAndBacktestCounters(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	id := registerHelper(t, fx, "alice", "alice@example.com")

	if err := fx.svc.IncrementStrategyCount(ctx, id, 1); err != nil {
		t.Fatalf("strategy +1: %v", err)
	}
	if err := fx.svc.IncrementStrategyCount(ctx, id, 1); err != nil {
		t.Fatalf("strategy +1: %v", err)
	}
	if err := fx.svc.IncrementBacktestCount(ctx, id, 1); err != nil {
		t.Fatalf("backtest +1: %v", err)
	}

	prof, err := fx.svc.GetProfile(ctx, id)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if prof.StrategyCount != 2 {
		t.Fatalf("expected StrategyCount=2, got %d", prof.StrategyCount)
	}
	if prof.BacktestCount != 1 {
		t.Fatalf("expected BacktestCount=1, got %d", prof.BacktestCount)
	}

	// Decrement past zero clamps at floor.
	for i := 0; i < 5; i++ {
		if err := fx.svc.IncrementStrategyCount(ctx, id, -1); err != nil {
			t.Fatalf("strategy -1 iter=%d: %v", i, err)
		}
	}
	prof, err = fx.svc.GetProfile(ctx, id)
	if err != nil {
		t.Fatalf("GetProfile after decrement: %v", err)
	}
	if prof.StrategyCount != 0 {
		t.Fatalf("expected clamp to 0, got %d", prof.StrategyCount)
	}
}

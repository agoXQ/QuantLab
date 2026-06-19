package tests

import (
	"context"
	"errors"
	"testing"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
)

// TestRegisterLoginUpdate walks the canonical sign-up path and the
// follow-up profile edit. The flow asserts that:
//   - Register persists the row, issues tokens, and emits Registered.
//   - Login finds the row by email and re-issues tokens.
//   - UpdateProfile applies the patch and emits Updated.
func TestRegisterLoginUpdate(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()

	res, err := fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice",
		Email:    "Alice@example.com",
		Password: "hunter2hunter",
		Nickname: "Alice",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.User.ID == 0 {
		t.Fatalf("expected user id assigned")
	}
	if res.User.Email != "alice@example.com" {
		t.Fatalf("expected email lowercased, got %q", res.User.Email)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatalf("expected token pair, got empty")
	}

	loginRes, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email:    "alice@example.com",
		Password: "hunter2hunter",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginRes.User.ID != res.User.ID {
		t.Fatalf("expected same user id, got %d vs %d", loginRes.User.ID, res.User.ID)
	}

	bio := "long term holder"
	avatar := "https://cdn.example.com/a.png"
	updated, err := fx.svc.UpdateProfile(ctx, appUser.UpdateProfileRequest{
		UserID: res.User.ID,
		Bio:    &bio,
		Avatar: &avatar,
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.Bio != bio {
		t.Fatalf("expected bio %q, got %q", bio, updated.Bio)
	}

	types := collectTypes(fx.publisher.Snapshot())
	for _, want := range []domevent.EventType{
		domevent.EventUserRegistered,
		domevent.EventUserUpdated,
	} {
		if _, ok := types[want]; !ok {
			t.Errorf("missing event %s, saw %v", want, types)
		}
	}
}

// TestRegisterRejectsDuplicates guards the email / username uniqueness
// check.
func TestRegisterRejectsDuplicates(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()

	if _, err := fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "hunter2hunter",
	}); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	_, err := fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice2",
		Email:    "alice@example.com",
		Password: "hunter2hunter",
	})
	if !errors.Is(err, userErr.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}

	_, err = fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice",
		Email:    "alice2@example.com",
		Password: "hunter2hunter",
	})
	if !errors.Is(err, userErr.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

// TestLoginInvalidPassword confirms wrong passwords surface as the
// generic invalid-credentials error.
func TestLoginInvalidPassword(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	if _, err := fx.svc.Register(ctx, appUser.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "hunter2hunter",
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrong-password",
	}); !errors.Is(err, userErr.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if _, err := fx.svc.Login(ctx, appUser.LoginRequest{
		Email:    "ghost@example.com",
		Password: "anything",
	}); !errors.Is(err, userErr.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for missing user, got %v", err)
	}
}

// TestFollowUnfollow walks the canonical social-graph path: follow,
// list, count, unfollow, double-follow rejection, and self-follow
// rejection.
func TestFollowUnfollow(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	a := registerHelper(t, fx, "alice", "alice@example.com")
	b := registerHelper(t, fx, "bob", "bob@example.com")

	if err := fx.svc.Follow(ctx, appUser.FollowRequest{FollowerID: a, FolloweeID: b}); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	if err := fx.svc.Follow(ctx, appUser.FollowRequest{FollowerID: a, FolloweeID: b}); !errors.Is(err, userErr.ErrAlreadyFollowed) {
		t.Fatalf("expected ErrAlreadyFollowed, got %v", err)
	}
	if err := fx.svc.Follow(ctx, appUser.FollowRequest{FollowerID: a, FolloweeID: a}); !errors.Is(err, userErr.ErrSelfFollow) {
		t.Fatalf("expected ErrSelfFollow, got %v", err)
	}

	bProfile, err := fx.svc.GetProfile(ctx, b)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if bProfile.FollowerCount != 1 {
		t.Fatalf("expected b to have 1 follower, got %d", bProfile.FollowerCount)
	}
	if bProfile.FollowingCount != 0 {
		t.Fatalf("expected b to follow 0, got %d", bProfile.FollowingCount)
	}

	followers, err := fx.svc.ListFollowers(ctx, b, 10, 0)
	if err != nil {
		t.Fatalf("ListFollowers: %v", err)
	}
	if len(followers.Users) != 1 || followers.Users[0].ID != a {
		t.Fatalf("expected follower a, got %v", followers.Users)
	}

	if err := fx.svc.Unfollow(ctx, appUser.FollowRequest{FollowerID: a, FolloweeID: b}); err != nil {
		t.Fatalf("Unfollow: %v", err)
	}
	if err := fx.svc.Unfollow(ctx, appUser.FollowRequest{FollowerID: a, FolloweeID: b}); !errors.Is(err, userErr.ErrFollowNotFound) {
		t.Fatalf("expected ErrFollowNotFound, got %v", err)
	}

	types := collectTypes(fx.publisher.Snapshot())
	for _, want := range []domevent.EventType{
		domevent.EventUserFollowed,
		domevent.EventUserUnfollowed,
	} {
		if _, ok := types[want]; !ok {
			t.Errorf("missing event %s, saw %v", want, types)
		}
	}
}

// registerHelper is a small helper used by the follow tests; it
// registers a user with a deterministic password and returns the id.
func registerHelper(t *testing.T, fx *fixture, username, email string) int64 {
	t.Helper()
	res, err := fx.svc.Register(context.Background(), appUser.RegisterRequest{
		Username: username,
		Email:    email,
		Password: "hunter2hunter",
	})
	if err != nil {
		t.Fatalf("register %s: %v", username, err)
	}
	return res.User.ID
}

// collectTypes walks an event slice and collapses it into a set on
// EventType for easy lookup in test assertions.
func collectTypes(in []domevent.Event) map[domevent.EventType]struct{} {
	out := make(map[domevent.EventType]struct{}, len(in))
	for _, e := range in {
		out[e.EventType] = struct{}{}
	}
	return out
}

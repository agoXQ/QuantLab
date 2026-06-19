package user

import (
	"testing"
	"time"

	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{"happy", "alice_01", true},
		{"with-dash", "ali-bob", true},
		{"too-short", "ab", false},
		{"with-space", "ali bob", false},
		{"with-symbol", "alice!", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUsername(tc.in)
			if (err == nil) != tc.ok {
				t.Fatalf("ValidateUsername(%q): want ok=%v, got err=%v", tc.in, tc.ok, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	if err := ValidateEmail("alice@example.com"); err != nil {
		t.Fatalf("good email rejected: %v", err)
	}
	if err := ValidateEmail("not-an-email"); err == nil {
		t.Fatalf("bad email accepted")
	}
	if err := ValidateEmail(""); err == nil {
		t.Fatalf("empty email accepted")
	}
}

func TestUserValidate(t *testing.T) {
	now := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
	u := &User{
		Username:       "alice",
		Email:          "alice@example.com",
		Status:         valueobject.AccountStatusActive,
		MembershipTier: valueobject.MembershipTierFree,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := u.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	bad := *u
	bad.Status = valueobject.AccountStatus(99)
	if err := bad.Validate(); err == nil {
		t.Fatalf("expected invalid status to fail validation")
	}
}

func TestApplyProfilePatch(t *testing.T) {
	now := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
	u := &User{Bio: "old", Nickname: "Alice"}
	bio := "  new bio  "
	loc := "Tokyo"
	u.ApplyProfilePatch(ProfilePatch{Bio: &bio, Location: &loc}, now)
	if u.Bio != "new bio" {
		t.Fatalf("expected trimmed bio, got %q", u.Bio)
	}
	if u.Location != "Tokyo" {
		t.Fatalf("expected location set, got %q", u.Location)
	}
	if u.Nickname != "Alice" {
		t.Fatalf("expected nickname unchanged, got %q", u.Nickname)
	}
	if !u.UpdatedAt.Equal(now) {
		t.Fatalf("expected UpdatedAt stamped")
	}
}

func TestEnsureLoginable(t *testing.T) {
	cases := []struct {
		name    string
		status  valueobject.AccountStatus
		wantErr bool
	}{
		{"active", valueobject.AccountStatusActive, false},
		{"suspended", valueobject.AccountStatusSuspended, true},
		{"banned", valueobject.AccountStatusBanned, true},
		{"deleted", valueobject.AccountStatusDeleted, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Status: tc.status}
			err := u.EnsureLoginable()
			if (err != nil) != tc.wantErr {
				t.Fatalf("EnsureLoginable(%v): wantErr=%v, got err=%v", tc.status, tc.wantErr, err)
			}
		})
	}
}

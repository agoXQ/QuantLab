package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
)

// TestNewJWTIssuer_RejectsShortSecret confirms the platform floor
// kicks in: HS256 keys shorter than 32 bytes fail the boot.
func TestNewJWTIssuer_RejectsShortSecret(t *testing.T) {
	if _, err := token.NewJWTIssuer(token.Config{Secret: "too-short"}); !errors.Is(err, token.ErrSecretTooShort) {
		t.Fatalf("expected ErrSecretTooShort, got %v", err)
	}
}

// TestKeyRotation_VerifyHonoursPreviousKey ensures a rotation is
// non-disruptive: tokens issued under the previous key keep verifying
// while the active key flips to a new id.
func TestKeyRotation_VerifyHonoursPreviousKey(t *testing.T) {
	first := strings.Repeat("a", 32)
	second := strings.Repeat("b", 32)
	keysOld, err := token.NewKeySet("k1", []token.SigningKey{{ID: "k1", Secret: first}})
	if err != nil {
		t.Fatalf("NewKeySet old: %v", err)
	}
	oldIssuer := token.MustNewJWTIssuer(token.Config{Keys: keysOld, AccessTTL: 5 * time.Minute})
	pair, err := oldIssuer.Issue(42)
	if err != nil {
		t.Fatalf("issue old: %v", err)
	}

	// Rotate: append the new key, flip the active id.
	keysNew, err := token.NewKeySet("k2", []token.SigningKey{
		{ID: "k1", Secret: first},
		{ID: "k2", Secret: second},
	})
	if err != nil {
		t.Fatalf("NewKeySet new: %v", err)
	}
	newIssuer := token.MustNewJWTIssuer(token.Config{Keys: keysNew, AccessTTL: 5 * time.Minute})

	// Old token still verifies because k1 is still in the set.
	if id, err := newIssuer.Verify(pair.AccessToken, token.KindAccess); err != nil || id != 42 {
		t.Fatalf("verify old token after rotation: id=%d err=%v", id, err)
	}

	// New tokens are signed with k2 and verify cleanly.
	pair2, err := newIssuer.Issue(7)
	if err != nil {
		t.Fatalf("issue new: %v", err)
	}
	if id, err := newIssuer.Verify(pair2.AccessToken, token.KindAccess); err != nil || id != 7 {
		t.Fatalf("verify new token: id=%d err=%v", id, err)
	}

	// After dropping k1, the old token must no longer verify.
	keysAfter, _ := token.NewKeySet("k2", []token.SigningKey{{ID: "k2", Secret: second}})
	afterIssuer := token.MustNewJWTIssuer(token.Config{Keys: keysAfter, AccessTTL: 5 * time.Minute})
	if _, err := afterIssuer.Verify(pair.AccessToken, token.KindAccess); !errors.Is(err, userErr.ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid after dropping k1, got %v", err)
	}
}

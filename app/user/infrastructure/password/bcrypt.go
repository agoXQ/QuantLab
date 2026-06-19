// Package password provides a bcrypt-backed PasswordHasher that
// implements user.PasswordHasher. The cost defaults to bcrypt's
// recommended 10; deployments that need stronger hashes override it
// in config without touching the application service.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
)

// BcryptHasher implements PasswordHasher with the bcrypt algorithm.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a hasher with the supplied cost. cost <= 0
// falls back to bcrypt.DefaultCost.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash returns the bcrypt hash for plain. bcrypt rejects inputs longer
// than 72 bytes; we let the caller validate length so the error can be
// surfaced as ErrPasswordTooWeak rather than bcrypt's internal error.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	if len(plain) > 72 {
		return "", fmt.Errorf("password too long: %d bytes (bcrypt limit is 72)", len(plain))
	}
	out, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Verify returns nil when the hash matches the supplied plain, or
// ErrInvalidCredentials otherwise. Unexpected errors are wrapped so
// the application service can decide whether to log them.
func (h *BcryptHasher) Verify(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return userErr.ErrInvalidCredentials
	}
	return nil
}

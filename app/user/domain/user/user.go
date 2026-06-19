// Package user defines the User aggregate.
//
// The aggregate owns the account credentials, profile metadata, and
// the lifecycle / verification flags. Heavy state — followers, posts,
// strategies — lives behind sibling aggregates / services. The
// User aggregate is intentionally narrow so authentication can short-
// circuit reads to a single row.
package user

import (
	"context"
	"net/mail"
	"strings"
	"time"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

// MaxUsernameLength caps the username at the schema limit so a renderer
// never has to truncate.
const MaxUsernameLength = 64

// MinUsernameLength enforces a sensible floor; usernames below this
// reduce the surface area for unique identifiers.
const MinUsernameLength = 3

// User is the aggregate root.
type User struct {
	ID             int64                       `json:"id"`
	Username       string                      `json:"username"`
	Email          string                      `json:"email"`
	PasswordHash   string                      `json:"-"`
	Avatar         string                      `json:"avatar,omitempty"`
	Bio            string                      `json:"bio,omitempty"`
	Nickname       string                      `json:"nickname,omitempty"`
	Location       string                      `json:"location,omitempty"`
	Status         valueobject.AccountStatus   `json:"status"`
	CreatorStatus  valueobject.CreatorStatus   `json:"creator_status"`
	VerifiedStatus valueobject.VerifiedStatus  `json:"verified_status"`
	MembershipTier valueobject.MembershipTier  `json:"membership_tier"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
	LastLoginAt    *time.Time                  `json:"last_login_at,omitempty"`
	StrategyCount  int64                       `json:"strategy_count"`
	BacktestCount  int64                       `json:"backtest_count"`
}

// Validate runs structural checks. The aggregate refuses to persist
// rows the rest of the system cannot trust.
func (u *User) Validate() error {
	if u == nil {
		return userErr.ErrInvalidUser
	}
	if err := ValidateUsername(u.Username); err != nil {
		return err
	}
	if err := ValidateEmail(u.Email); err != nil {
		return err
	}
	if !u.Status.IsValid() {
		return userErr.ErrInvalidStatus
	}
	if !u.MembershipTier.IsValid() {
		return userErr.ErrInvalidTier
	}
	return nil
}

// EnsureLoginable checks the account state guards we apply during
// login. A suspended / banned / deleted account refuses tokens even
// when the password matches.
func (u *User) EnsureLoginable() error {
	switch u.Status {
	case valueobject.AccountStatusSuspended:
		return userErr.ErrAccountSuspended
	case valueobject.AccountStatusBanned:
		return userErr.ErrAccountBanned
	case valueobject.AccountStatusDeleted:
		return userErr.ErrAccountDeleted
	}
	return nil
}

// ApplyProfilePatch updates the editable profile fields. Empty pointer
// fields stay untouched; a non-nil pointer to "" clears the field.
func (u *User) ApplyProfilePatch(p ProfilePatch, now time.Time) {
	if p.Avatar != nil {
		u.Avatar = strings.TrimSpace(*p.Avatar)
	}
	if p.Bio != nil {
		u.Bio = strings.TrimSpace(*p.Bio)
	}
	if p.Nickname != nil {
		u.Nickname = strings.TrimSpace(*p.Nickname)
	}
	if p.Location != nil {
		u.Location = strings.TrimSpace(*p.Location)
	}
	u.UpdatedAt = now
}

// MarkLoggedIn stamps the last login timestamp.
func (u *User) MarkLoggedIn(now time.Time) {
	t := now
	u.LastLoginAt = &t
	u.UpdatedAt = now
}

// ValidateUsername enforces the platform-wide rules on usernames so the
// repository layer never accepts a row the API would reject.
func ValidateUsername(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || len(s) < MinUsernameLength || len(s) > MaxUsernameLength {
		return userErr.ErrUsernameInvalid
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '.' || r == '-':
		default:
			return userErr.ErrUsernameInvalid
		}
	}
	return nil
}

// ValidateEmail enforces RFC-5322 parsing through the standard library.
// We deliberately do not check the MX record here; the platform sends
// a verification email separately and surfaces failures to the user.
func ValidateEmail(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return userErr.ErrEmailRequired
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return userErr.ErrEmailInvalid
	}
	return nil
}

// ProfilePatch carries an optional metadata patch. Pointer fields keep
// "leave as-is" semantically distinct from "set to empty string".
type ProfilePatch struct {
	Avatar   *string
	Bio      *string
	Nickname *string
	Location *string
}

// Repository persists User aggregates.
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Get(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)

	// IncrementStrategyCount / IncrementBacktestCount mutate the
	// activity counters atomically. delta may be negative; the
	// repository must clamp the resulting value at zero so a stray
	// "deleted" event cannot push the counter below the floor. A
	// missing user surfaces ErrUserNotFound.
	IncrementStrategyCount(ctx context.Context, userID int64, delta int64) error
	IncrementBacktestCount(ctx context.Context, userID int64, delta int64) error
}

// Package user is the application layer for the User Service. It
// composes the User / Follow repositories, the password hasher, the
// token issuer, and the event publisher into a flat set of use cases
// that the HTTP / gRPC handlers call.
package user

import (
	"context"
	"time"

	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
)

// Service is the application-level interface for the User Service.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResult, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResult, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	UpdateAccount(ctx context.Context, req UpdateAccountRequest) (*domuser.User, error)
	Get(ctx context.Context, userID int64) (*domuser.User, error)
	GetProfile(ctx context.Context, userID int64) (*ProfileSnapshot, error)
	UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*domuser.User, error)
	Follow(ctx context.Context, req FollowRequest) error
	Unfollow(ctx context.Context, req FollowRequest) error
	ListFollowers(ctx context.Context, userID int64, limit, offset int) (*FollowList, error)
	ListFollowing(ctx context.Context, userID int64, limit, offset int) (*FollowList, error)

	// IncrementStrategyCount / IncrementBacktestCount are wired by the
	// cross-service event subscribers so the profile counters reflect
	// real activity. They are tolerant: an unknown user id is a no-op.
	IncrementStrategyCount(ctx context.Context, userID int64, delta int64) error
	IncrementBacktestCount(ctx context.Context, userID int64, delta int64) error
}

// PasswordHasher abstracts the hashing function so tests can use a
// deterministic / fast hasher and production wires bcrypt.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash string, plain string) error
}

// TokenIssuer mints access / refresh tokens for an authenticated user.
type TokenIssuer interface {
	Issue(userID int64) (TokenPair, error)
}

// RefreshTokenVerifier validates a refresh token and returns the user
// id encoded in the subject claim. The refresh use case keeps the
// dependency narrow so a future revocation backend lands here without
// pulling in the full TokenIssuer surface.
type RefreshTokenVerifier interface {
	VerifyRefresh(token string) (int64, error)
}

// TokenPair carries the freshly minted token strings + expiry hint
// (seconds until access token expiration).
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Dependencies bundles the ports the service needs.
type Dependencies struct {
	Users           domuser.Repository
	Follows         domfollow.Repository
	Hasher          PasswordHasher
	Tokens          TokenIssuer
	RefreshVerifier RefreshTokenVerifier
	Publisher       domevent.Publisher
	Clock           func() time.Time
}

type service struct {
	deps Dependencies
}

// NewService builds the default application service. Publisher is
// optional (nil = no events); Clock defaults to time.Now.
func NewService(deps Dependencies) Service {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	return &service{deps: deps}
}

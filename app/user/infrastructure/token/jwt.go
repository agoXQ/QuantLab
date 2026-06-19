// Package token provides a JWT-backed TokenIssuer for the User Service.
// The MVP issues HS256 tokens via a KeySet so secret rotation drops in
// without rewiring callers; production swaps in a JWKS-backed loader
// behind the same TokenIssuer port.
package token

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
)

// audienceAccess / audienceRefresh tag the two token kinds the issuer
// mints. Verify rejects a refresh token when the caller asks for an
// access token (and vice versa) so an attacker that captures a long-
// lived refresh token cannot use it to call protected APIs directly.
const (
	audienceAccess  = "access"
	audienceRefresh = "refresh"
)

// TokenKind names the high-level intent of a token; exported so other
// packages (middleware, refresh use case) can keep their assumptions
// explicit.
type TokenKind string

const (
	KindAccess  TokenKind = audienceAccess
	KindRefresh TokenKind = audienceRefresh
)

// Config configures the JWT issuer.
type Config struct {
	// Keys is the rotation-aware key set. New deployments should
	// supply at least one key; legacy callers may still set Secret.
	Keys *KeySet
	// Secret is the legacy single-key shim. Provided for backwards
	// compatibility with callers that have not yet migrated to Keys;
	// when set without Keys the issuer wraps it in a single-key
	// KeySet so the rest of the path stays uniform.
	Secret string
	// Issuer is the platform-wide iss claim.
	Issuer string
	// AccessTTL is the access-token validity window.
	AccessTTL time.Duration
	// RefreshTTL is the refresh-token validity window.
	RefreshTTL time.Duration
	// Clock is overridable for tests.
	Clock func() time.Time
}

// JWTIssuer implements TokenIssuer with HS256 JWTs.
type JWTIssuer struct {
	cfg  Config
	keys *KeySet
}

// NewJWTIssuer wires the issuer; defaults are filled in for missing
// fields so callers only have to supply Keys + Issuer.
func NewJWTIssuer(cfg Config) (*JWTIssuer, error) {
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 30 * time.Minute
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 14 * 24 * time.Hour
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "quantlab.user"
	}
	keys := cfg.Keys
	if keys == nil {
		built, err := SingleKeySet("default", cfg.Secret)
		if err != nil {
			return nil, err
		}
		keys = built
	}
	return &JWTIssuer{cfg: cfg, keys: keys}, nil
}

// MustNewJWTIssuer is the panicking counterpart used by tests and
// other paths that already validate the configuration upstream.
func MustNewJWTIssuer(cfg Config) *JWTIssuer {
	issuer, err := NewJWTIssuer(cfg)
	if err != nil {
		panic(err)
	}
	return issuer
}

// Issue mints a fresh access + refresh token pair.
func (i *JWTIssuer) Issue(userID int64) (appUser.TokenPair, error) {
	if i == nil || i.keys == nil {
		return appUser.TokenPair{}, fmt.Errorf("token: issuer not initialised")
	}
	now := i.cfg.Clock()
	access, err := i.signClaims(now, userID, audienceAccess, i.cfg.AccessTTL)
	if err != nil {
		return appUser.TokenPair{}, err
	}
	refresh, err := i.signClaims(now, userID, audienceRefresh, i.cfg.RefreshTTL)
	if err != nil {
		return appUser.TokenPair{}, err
	}
	return appUser.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(i.cfg.AccessTTL.Seconds()),
	}, nil
}

// Verify validates the supplied token string and returns the user id
// encoded in the subject claim. Exposed so middleware (HTTP / gRPC
// interceptor) and the refresh use case can reuse the same parser.
// kind selects which audience the token must carry; pass an empty
// string to skip the audience check (useful in tests).
func (i *JWTIssuer) Verify(token string, kind TokenKind) (int64, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, userErr.ErrTokenInvalid
		}
		kid, _ := t.Header["kid"].(string)
		secret, _, ok := i.keys.Lookup(kid)
		if !ok {
			return nil, userErr.ErrTokenInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		if isExpired(err) {
			return 0, userErr.ErrTokenExpired
		}
		return 0, userErr.ErrTokenInvalid
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsed.Valid {
		return 0, userErr.ErrTokenInvalid
	}
	if kind != "" && !audienceMatches(claims.Audience, string(kind)) {
		return 0, userErr.ErrTokenInvalid
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, userErr.ErrTokenInvalid
	}
	return id, nil
}

func (i *JWTIssuer) signClaims(now time.Time, userID int64, kind string, ttl time.Duration) (string, error) {
	active := i.keys.Active()
	claims := jwt.RegisteredClaims{
		Issuer:    i.cfg.Issuer,
		Subject:   strconv.FormatInt(userID, 10),
		Audience:  jwt.ClaimStrings{kind},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		ID:        uuid.NewString(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok.Header["kid"] = active.ID
	return tok.SignedString([]byte(active.Secret))
}

func isExpired(err error) bool {
	var verr *jwt.ValidationError
	if asErr, ok := err.(*jwt.ValidationError); ok {
		verr = asErr
	}
	if verr == nil {
		return false
	}
	return verr.Errors&jwt.ValidationErrorExpired != 0
}

func audienceMatches(have jwt.ClaimStrings, want string) bool {
	for _, a := range have {
		if a == want {
			return true
		}
	}
	return false
}

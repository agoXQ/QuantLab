// Package token provides a JWT-backed TokenIssuer for the User Service.
// The MVP issues HS256 tokens; production swaps in a key-rotation aware
// issuer behind the same TokenIssuer port.
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

// Config configures the JWT issuer.
type Config struct {
	// Secret is the HS256 signing key. Must be at least 32 bytes in
	// production; tests may supply a shorter constant.
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
	cfg Config
}

// NewJWTIssuer wires the issuer; defaults are filled in for missing
// fields so callers only have to supply Secret + Issuer.
func NewJWTIssuer(cfg Config) *JWTIssuer {
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
	return &JWTIssuer{cfg: cfg}
}

// Issue mints a fresh access + refresh token pair.
func (i *JWTIssuer) Issue(userID int64) (appUser.TokenPair, error) {
	if i.cfg.Secret == "" {
		return appUser.TokenPair{}, fmt.Errorf("token: signing secret is empty")
	}
	now := i.cfg.Clock()
	access, err := i.signClaims(now, userID, "access", i.cfg.AccessTTL)
	if err != nil {
		return appUser.TokenPair{}, err
	}
	refresh, err := i.signClaims(now, userID, "refresh", i.cfg.RefreshTTL)
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
// encoded in the subject claim. Exposed so future middleware (HTTP /
// gRPC interceptor) can reuse the same parser.
func (i *JWTIssuer) Verify(token string) (int64, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, userErr.ErrTokenInvalid
		}
		return []byte(i.cfg.Secret), nil
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
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, userErr.ErrTokenInvalid
	}
	return id, nil
}

func (i *JWTIssuer) signClaims(now time.Time, userID int64, kind string, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    i.cfg.Issuer,
		Subject:   strconv.FormatInt(userID, 10),
		Audience:  jwt.ClaimStrings{kind},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		ID:        uuid.NewString(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(i.cfg.Secret))
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

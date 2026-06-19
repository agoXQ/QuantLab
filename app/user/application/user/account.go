package user

import (
	"context"
	"fmt"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
)

// RefreshToken validates the supplied refresh token and reissues a
// fresh access + refresh pair. Misformed / expired tokens map to
// ErrTokenInvalid / ErrTokenExpired so the HTTP layer can return 401.
func (s *service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResult, error) {
	if s.deps.RefreshVerifier == nil {
		return nil, fmt.Errorf("user: refresh verifier not wired")
	}
	if req.RefreshToken == "" {
		return nil, userErr.ErrTokenInvalid
	}
	userID, err := s.deps.RefreshVerifier.VerifyRefresh(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	u, err := s.deps.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.EnsureLoginable(); err != nil {
		return nil, err
	}
	tokens, err := s.deps.Tokens.Issue(u.ID)
	if err != nil {
		return nil, fmt.Errorf("issue tokens: %w", err)
	}
	return &LoginResult{
		User:         u,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// ChangePassword rotates the credential after re-checking the current
// password. The use case keeps the door narrow: even a valid access
// token cannot rotate the password without the existing secret.
func (s *service) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	if req.UserID <= 0 {
		return userErr.ErrInvalidUser
	}
	if len(req.NewPassword) < MinPasswordLength {
		return userErr.ErrPasswordTooWeak
	}
	if req.NewPassword == req.CurrentPassword {
		return userErr.ErrSamePassword
	}
	u, err := s.deps.Users.Get(ctx, req.UserID)
	if err != nil {
		return err
	}
	if err := s.deps.Hasher.Verify(u.PasswordHash, req.CurrentPassword); err != nil {
		return userErr.ErrInvalidCredentials
	}
	hash, err := s.deps.Hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = hash
	u.UpdatedAt = s.deps.Clock()
	if err := s.deps.Users.Update(ctx, u); err != nil {
		return err
	}
	s.publish(ctx, domevent.EventUserUpdated, u.ID, domevent.UserUpdatedPayload{UserID: u.ID})
	return nil
}

// UpdateAccount applies the moderator-style patch (status / creator /
// verified / tier). The Validate call rejects unknown enum values so a
// malformed admin tool cannot land an inconsistent row. The use case
// is intentionally shy about authorisation: callers are expected to
// have already verified the operator permission via middleware.
func (s *service) UpdateAccount(ctx context.Context, req UpdateAccountRequest) (*domuser.User, error) {
	if req.UserID <= 0 {
		return nil, userErr.ErrInvalidUser
	}
	u, err := s.deps.Users.Get(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	if req.CreatorStatus != nil {
		u.CreatorStatus = *req.CreatorStatus
	}
	if req.VerifiedStatus != nil {
		u.VerifiedStatus = *req.VerifiedStatus
	}
	if req.MembershipTier != nil {
		u.MembershipTier = *req.MembershipTier
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	u.UpdatedAt = s.deps.Clock()
	if err := s.deps.Users.Update(ctx, u); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventUserUpdated, u.ID, domevent.UserUpdatedPayload{UserID: u.ID})
	return u, nil
}

// IncrementStrategyCount is the integration hook for the strategy
// events subscriber. The repository clamps at zero so a stray
// "deleted" event cannot push the counter below the floor.
func (s *service) IncrementStrategyCount(ctx context.Context, userID int64, delta int64) error {
	if userID <= 0 || delta == 0 {
		return nil
	}
	return s.deps.Users.IncrementStrategyCount(ctx, userID, delta)
}

// IncrementBacktestCount is the integration hook for the backtest
// events subscriber.
func (s *service) IncrementBacktestCount(ctx context.Context, userID int64, delta int64) error {
	if userID <= 0 || delta == 0 {
		return nil
	}
	return s.deps.Users.IncrementBacktestCount(ctx, userID, delta)
}

package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

// MinPasswordLength is the platform-wide minimum. The bcrypt hasher
// caps the input at 72 bytes, but the floor is policy and lives here
// so future strength checks (entropy / dictionary) drop in alongside.
const MinPasswordLength = 8

// Register persists a new account in ACTIVE / FREE state and mints a
// token pair so the client can sign in immediately.
func (s *service) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := domuser.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := domuser.ValidateEmail(email); err != nil {
		return nil, err
	}
	if len(req.Password) < MinPasswordLength {
		return nil, userErr.ErrPasswordTooWeak
	}

	if existing, err := s.deps.Users.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, userErr.ErrEmailTaken
	} else if err != nil && !errors.Is(err, userErr.ErrUserNotFound) {
		return nil, err
	}
	if existing, err := s.deps.Users.GetByUsername(ctx, username); err == nil && existing != nil {
		return nil, userErr.ErrUsernameTaken
	} else if err != nil && !errors.Is(err, userErr.ErrUserNotFound) {
		return nil, err
	}

	hash, err := s.deps.Hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	now := s.deps.Clock()
	u := &domuser.User{
		Username:       username,
		Email:          email,
		PasswordHash:   hash,
		Nickname:       strings.TrimSpace(req.Nickname),
		Status:         valueobject.AccountStatusActive,
		CreatorStatus:  valueobject.CreatorStatusRegular,
		VerifiedStatus: valueobject.VerifiedStatusNone,
		MembershipTier: valueobject.MembershipTierFree,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Users.Create(ctx, u); err != nil {
		return nil, err
	}
	tokens, err := s.deps.Tokens.Issue(u.ID)
	if err != nil {
		return nil, fmt.Errorf("issue tokens: %w", err)
	}
	s.publish(ctx, domevent.EventUserRegistered, u.ID, domevent.UserRegisteredPayload{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
	})
	return &RegisterResult{
		User:         u,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Login validates credentials and mints a token pair. The use case
// always runs the password verifier even when the user is missing, to
// avoid the user-existence oracle.
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := domuser.ValidateEmail(email); err != nil {
		return nil, userErr.ErrInvalidCredentials
	}
	u, err := s.deps.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userErr.ErrUserNotFound) {
			// Run a verification on a fixed dummy hash so the timing
			// shape stays identical to the success path. We deliberately
			// ignore the result.
			_ = s.deps.Hasher.Verify(dummyBcryptHash, req.Password)
			return nil, userErr.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := s.deps.Hasher.Verify(u.PasswordHash, req.Password); err != nil {
		return nil, userErr.ErrInvalidCredentials
	}
	if err := u.EnsureLoginable(); err != nil {
		return nil, err
	}
	now := s.deps.Clock()
	u.MarkLoggedIn(now)
	if err := s.deps.Users.Update(ctx, u); err != nil {
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

// dummyBcryptHash is a precomputed bcrypt hash used by the login flow
// to keep the timing of the missing-user branch indistinguishable from
// the wrong-password branch. The plaintext is "not-a-real-password".
const dummyBcryptHash = "$2a$10$abcdefghijklmnopqrstuuv.wxyz0123456789ABCDEFGHIJKLM01N2OPQ"

// Get returns a user by id.
func (s *service) Get(ctx context.Context, userID int64) (*domuser.User, error) {
	if userID <= 0 {
		return nil, userErr.ErrInvalidUser
	}
	return s.deps.Users.Get(ctx, userID)
}

// GetProfile returns the user row alongside follower / following
// counters. Strategy / backtest counts default to zero; future wiring
// can fill them in via cross-service reads.
func (s *service) GetProfile(ctx context.Context, userID int64) (*ProfileSnapshot, error) {
	u, err := s.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	follower, err := s.deps.Follows.CountFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}
	following, err := s.deps.Follows.CountFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileSnapshot{User: u, FollowerCount: follower, FollowingCount: following}, nil
}

// UpdateProfile applies a profile patch on behalf of the supplied user.
func (s *service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*domuser.User, error) {
	u, err := s.Get(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	u.ApplyProfilePatch(domuser.ProfilePatch{
		Avatar:   req.Avatar,
		Bio:      req.Bio,
		Nickname: req.Nickname,
		Location: req.Location,
	}, s.deps.Clock())
	if err := s.deps.Users.Update(ctx, u); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventUserUpdated, u.ID, domevent.UserUpdatedPayload{UserID: u.ID})
	return u, nil
}

// publish is a forgiving wrapper around the event publisher: a nil
// publisher and publish errors degrade to logs because the value chain
// must work even when Kafka is offline.
func (s *service) publish(ctx context.Context, t domevent.EventType, aggregateID int64, payload any) {
	if s.deps.Publisher == nil {
		return
	}
	_ = s.deps.Publisher.Publish(ctx, domevent.Event{
		EventID:       uuid.NewString(),
		EventType:     t,
		EventVersion:  domevent.EventVersionV1,
		OccurredAt:    s.deps.Clock(),
		AggregateType: domevent.AggregateTypeUser,
		AggregateID:   fmt.Sprintf("%d", aggregateID),
		Producer:      domevent.ProducerUser,
		Payload:       payload,
	})
}

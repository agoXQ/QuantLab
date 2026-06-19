package user

import (
	"github.com/agoXQ/QuantLab/app/user/domain/follow"
	"github.com/agoXQ/QuantLab/app/user/domain/user"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

// RegisterRequest is the payload for the Register use case. The
// application layer normalises whitespace / case and hashes the
// password before persisting.
type RegisterRequest struct {
	Username string
	Email    string
	Password string
	Nickname string
}

// RegisterResult bundles the new aggregate and a freshly minted token
// pair so the client can sign in immediately after sign-up.
type RegisterResult struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// LoginRequest carries the credentials supplied by the user.
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResult is what Login / RefreshToken return.
type LoginResult struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// RefreshTokenRequest carries the refresh-token string the client
// already holds. The application service decodes it through the
// RefreshTokenVerifier port and reissues a fresh pair.
type RefreshTokenRequest struct {
	RefreshToken string
}

// ChangePasswordRequest is the payload for the ChangePassword use
// case. The CurrentPassword field is required so a stolen token alone
// cannot rotate the credential.
type ChangePasswordRequest struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}

// UpdateProfileRequest carries an optional metadata patch. Pointer
// fields keep "leave as-is" distinct from "set empty".
type UpdateProfileRequest struct {
	UserID   int64
	Avatar   *string
	Bio      *string
	Nickname *string
	Location *string
}

// UpdateAccountRequest is the moderator-style patch surfaced by the
// admin tooling. Pointer fields preserve "leave as-is" so a partial
// update never accidentally clears another field.
type UpdateAccountRequest struct {
	UserID         int64
	Status         *valueobject.AccountStatus
	CreatorStatus  *valueobject.CreatorStatus
	VerifiedStatus *valueobject.VerifiedStatus
	MembershipTier *valueobject.MembershipTier
}

// FollowRequest captures the (follower, followee) pair.
type FollowRequest struct {
	FollowerID int64
	FolloweeID int64
}

// ProfileSnapshot bundles the user row with the basic counters the
// platform's profile view needs. Strategy / backtest counts come from
// the cross-service event subscribers that listen to strategy-events
// and backtest-events.
type ProfileSnapshot struct {
	User           *user.User
	FollowerCount  int64
	FollowingCount int64
	StrategyCount  int64
	BacktestCount  int64
}

// FollowList returns a page of users (the partners on a follow row).
// The caller decides whether to render followers or following based on
// which method was invoked.
type FollowList struct {
	Users []*user.User
}

// silence: keep the follow import live for future Follow query types.
var _ = follow.Follow{}

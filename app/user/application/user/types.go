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

// LoginResult is what Login returns.
type LoginResult struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
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

// FollowRequest captures the (follower, followee) pair.
type FollowRequest struct {
	FollowerID int64
	FolloweeID int64
}

// ProfileSnapshot bundles the user row with the basic counters the
// platform's profile view needs. Strategy / backtest counts come from
// sibling services in production; the MVP ships zeros so the call
// stays self-contained.
type ProfileSnapshot struct {
	User           *user.User
	FollowerCount  int64
	FollowingCount int64
	StrategyCount  int64
	BacktestCount  int64
}

// FollowList returns a page of Follow rows alongside the user records
// of the partners. The caller decides whether to render followers or
// following based on which method was invoked.
type FollowList struct {
	Users []*user.User
}

// AccountUpdates is the optional moderator-style patch we expose for
// admin tooling. None of the MVP HTTP routes wire it; the type exists
// so the application service stays open to status changes from a
// future admin surface without growing more methods.
type AccountUpdates struct {
	Status         *valueobject.AccountStatus
	CreatorStatus  *valueobject.CreatorStatus
	VerifiedStatus *valueobject.VerifiedStatus
	MembershipTier *valueobject.MembershipTier
}

// silence: keep the follow import live for future Follow query types.
var _ = follow.Follow{}

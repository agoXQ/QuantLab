package user

import (
	"context"
	"errors"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
)

// Follow records that FollowerID follows FolloweeID. The use case is
// idempotent on retries (already-followed surfaces as an error so the
// API can render 409, but the row is left intact).
func (s *service) Follow(ctx context.Context, req FollowRequest) error {
	if req.FollowerID <= 0 || req.FolloweeID <= 0 {
		return userErr.ErrInvalidUser
	}
	if req.FollowerID == req.FolloweeID {
		return userErr.ErrSelfFollow
	}
	// Both rows must exist; we deliberately surface NotFound rather
	// than allowing dangling rows.
	if _, err := s.deps.Users.Get(ctx, req.FollowerID); err != nil {
		return err
	}
	if _, err := s.deps.Users.Get(ctx, req.FolloweeID); err != nil {
		return err
	}
	exists, err := s.deps.Follows.Exists(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		return err
	}
	if exists {
		return userErr.ErrAlreadyFollowed
	}
	now := s.deps.Clock()
	if err := s.deps.Follows.Create(ctx, &domfollow.Follow{
		FollowerID: req.FollowerID,
		FolloweeID: req.FolloweeID,
		CreatedAt:  now,
	}); err != nil {
		return err
	}
	s.publish(ctx, domevent.EventUserFollowed, req.FolloweeID, domevent.UserFollowedPayload{
		FollowerID: req.FollowerID,
		FolloweeID: req.FolloweeID,
	})
	return nil
}

// Unfollow drops the row. Missing rows surface NotFound so callers can
// distinguish "did nothing" from "noop"; the HTTP layer translates it
// to 404.
func (s *service) Unfollow(ctx context.Context, req FollowRequest) error {
	if req.FollowerID <= 0 || req.FolloweeID <= 0 {
		return userErr.ErrInvalidUser
	}
	if err := s.deps.Follows.Delete(ctx, req.FollowerID, req.FolloweeID); err != nil {
		return err
	}
	s.publish(ctx, domevent.EventUserUnfollowed, req.FolloweeID, domevent.UserUnfollowedPayload{
		FollowerID: req.FollowerID,
		FolloweeID: req.FolloweeID,
	})
	return nil
}

// ListFollowers returns the users following the supplied user. The use
// case translates Follow rows back into User rows via a per-row Get;
// production deployments should add a JOIN-flavoured repository method
// later for performance.
func (s *service) ListFollowers(ctx context.Context, userID int64, limit, offset int) (*FollowList, error) {
	rows, err := s.deps.Follows.ListFollowers(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.materialiseUsers(ctx, rows, true)
}

// ListFollowing returns the users this user follows.
func (s *service) ListFollowing(ctx context.Context, userID int64, limit, offset int) (*FollowList, error) {
	rows, err := s.deps.Follows.ListFollowing(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.materialiseUsers(ctx, rows, false)
}

// materialiseUsers walks Follow rows and loads the partner user. The
// asFollowers flag picks the correct side of the row.
func (s *service) materialiseUsers(ctx context.Context, rows []*domfollow.Follow, asFollowers bool) (*FollowList, error) {
	out := &FollowList{Users: make([]*domuser.User, 0, len(rows))}
	for _, row := range rows {
		var partnerID int64
		if asFollowers {
			partnerID = row.FollowerID
		} else {
			partnerID = row.FolloweeID
		}
		u, err := s.deps.Users.Get(ctx, partnerID)
		if err != nil {
			if errors.Is(err, userErr.ErrUserNotFound) {
				continue
			}
			return nil, err
		}
		out.Users = append(out.Users, u)
	}
	return out, nil
}

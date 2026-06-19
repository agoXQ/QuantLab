// Package follow defines the Follow aggregate. Follows are stored in
// their own table so the lineage graph can be queried without dragging
// the User row into the JOIN, and so the repository can scale to the
// asymmetric counts inherent in social graphs.
package follow

import (
	"context"
	"time"
)

// Follow records that FollowerID follows FolloweeID.
type Follow struct {
	ID         int64     `json:"id"`
	FollowerID int64     `json:"follower_id"`
	FolloweeID int64     `json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Repository persists Follow rows. Implementations must enforce the
// (follower_id, followee_id) uniqueness constraint via the database
// layer; the aggregate offers no other invariant beyond that pair.
type Repository interface {
	Create(ctx context.Context, f *Follow) error
	Delete(ctx context.Context, followerID, followeeID int64) error
	Exists(ctx context.Context, followerID, followeeID int64) (bool, error)
	ListFollowers(ctx context.Context, userID int64, limit, offset int) ([]*Follow, error)
	ListFollowing(ctx context.Context, userID int64, limit, offset int) ([]*Follow, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
}

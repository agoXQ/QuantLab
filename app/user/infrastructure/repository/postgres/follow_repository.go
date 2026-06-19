package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domfollow "github.com/agoXQ/QuantLab/app/user/domain/follow"
)

// FollowRepository persists Follow rows in PostgreSQL.
type FollowRepository struct {
	db *sql.DB
}

// NewFollowRepository wires the repository against an existing *sql.DB.
func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

// Create inserts a new follow row, mapping the unique-violation case to
// ErrAlreadyFollowed.
func (r *FollowRepository) Create(ctx context.Context, f *domfollow.Follow) error {
	if f == nil {
		return userErr.ErrInvalidUser
	}
	const stmt = `
		INSERT INTO user_follow (follower_id, followee_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	if err := r.db.QueryRowContext(ctx, stmt, f.FollowerID, f.FolloweeID, f.CreatedAt).Scan(&f.ID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
			return userErr.ErrAlreadyFollowed
		}
		return fmt.Errorf("insert user_follow: %w", err)
	}
	return nil
}

// Delete drops the row.
func (r *FollowRepository) Delete(ctx context.Context, followerID, followeeID int64) error {
	const stmt = `DELETE FROM user_follow WHERE follower_id = $1 AND followee_id = $2`
	res, err := r.db.ExecContext(ctx, stmt, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("delete user_follow: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return userErr.ErrFollowNotFound
	}
	return nil
}

// Exists reports whether the (follower, followee) pair is present.
func (r *FollowRepository) Exists(ctx context.Context, followerID, followeeID int64) (bool, error) {
	const q = `SELECT 1 FROM user_follow WHERE follower_id = $1 AND followee_id = $2 LIMIT 1`
	var dummy int
	err := r.db.QueryRowContext(ctx, q, followerID, followeeID).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query user_follow: %w", err)
	}
	return true, nil
}

// ListFollowers returns the follow rows where userID is the followee
// (people who follow userID).
func (r *FollowRepository) ListFollowers(ctx context.Context, userID int64, limit, offset int) ([]*domfollow.Follow, error) {
	return r.list(ctx, "followee_id", userID, limit, offset)
}

// ListFollowing returns the follow rows where userID is the follower
// (people userID follows).
func (r *FollowRepository) ListFollowing(ctx context.Context, userID int64, limit, offset int) ([]*domfollow.Follow, error) {
	return r.list(ctx, "follower_id", userID, limit, offset)
}

// CountFollowers counts the followers of userID.
func (r *FollowRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, "followee_id", userID)
}

// CountFollowing counts the users userID follows.
func (r *FollowRepository) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, "follower_id", userID)
}

func (r *FollowRepository) list(ctx context.Context, column string, userID int64, limit, offset int) ([]*domfollow.Follow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	q := fmt.Sprintf(`
		SELECT id, follower_id, followee_id, created_at
		FROM user_follow
		WHERE %s = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, column)
	rows, err := r.db.QueryContext(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query user_follow: %w", err)
	}
	defer rows.Close()
	out := make([]*domfollow.Follow, 0)
	for rows.Next() {
		var f domfollow.Follow
		if err := rows.Scan(&f.ID, &f.FollowerID, &f.FolloweeID, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

func (r *FollowRepository) count(ctx context.Context, column string, userID int64) (int64, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM user_follow WHERE %s = $1`, column)
	var n int64
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count user_follow: %w", err)
	}
	return n, nil
}

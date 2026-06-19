package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
)

// UserRepository persists User aggregates in PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository wires the repository against an existing *sql.DB.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user, assigning the generated id back. Unique
// violation errors map to ErrEmailTaken / ErrUsernameTaken so the
// application service does not need to inspect driver internals.
func (r *UserRepository) Create(ctx context.Context, u *domuser.User) error {
	if u == nil {
		return userErr.ErrInvalidUser
	}
	const stmt = `
		INSERT INTO app_user (
			username, email, password_hash,
			avatar, bio, nickname, location,
			status, creator_status, verified_status, membership_tier,
			created_at, updated_at, last_login_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14
		) RETURNING id`
	err := r.db.QueryRowContext(ctx, stmt,
		u.Username, u.Email, u.PasswordHash,
		nullableString(u.Avatar), nullableString(u.Bio),
		nullableString(u.Nickname), nullableString(u.Location),
		int32(u.Status), int32(u.CreatorStatus), int32(u.VerifiedStatus),
		string(u.MembershipTier),
		u.CreatedAt, u.UpdatedAt, u.LastLoginAt,
	).Scan(&u.ID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	return nil
}

// Update writes back the mutable columns. CreatedAt is pinned at insert
// time and never updated.
func (r *UserRepository) Update(ctx context.Context, u *domuser.User) error {
	if u == nil || u.ID == 0 {
		return userErr.ErrInvalidUser
	}
	const stmt = `
		UPDATE app_user SET
			username        = $2,
			email           = $3,
			password_hash   = $4,
			avatar          = $5,
			bio             = $6,
			nickname        = $7,
			location        = $8,
			status          = $9,
			creator_status  = $10,
			verified_status = $11,
			membership_tier = $12,
			updated_at      = $13,
			last_login_at   = $14
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, stmt,
		u.ID,
		u.Username, u.Email, u.PasswordHash,
		nullableString(u.Avatar), nullableString(u.Bio),
		nullableString(u.Nickname), nullableString(u.Location),
		int32(u.Status), int32(u.CreatorStatus), int32(u.VerifiedStatus),
		string(u.MembershipTier),
		u.UpdatedAt, u.LastLoginAt,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return userErr.ErrUserNotFound
	}
	return nil
}

// Get loads one user.
func (r *UserRepository) Get(ctx context.Context, id int64) (*domuser.User, error) {
	const q = `
		SELECT id, username, email, password_hash,
		       COALESCE(avatar, ''), COALESCE(bio, ''),
		       COALESCE(nickname, ''), COALESCE(location, ''),
		       status, creator_status, verified_status, membership_tier,
		       created_at, updated_at, last_login_at
		FROM app_user WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, id)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, userErr.ErrUserNotFound
	}
	return u, err
}

// GetByEmail returns the user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domuser.User, error) {
	const q = `
		SELECT id, username, email, password_hash,
		       COALESCE(avatar, ''), COALESCE(bio, ''),
		       COALESCE(nickname, ''), COALESCE(location, ''),
		       status, creator_status, verified_status, membership_tier,
		       created_at, updated_at, last_login_at
		FROM app_user WHERE LOWER(email) = LOWER($1)`
	row := r.db.QueryRowContext(ctx, q, email)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, userErr.ErrUserNotFound
	}
	return u, err
}

// GetByUsername returns the user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domuser.User, error) {
	const q = `
		SELECT id, username, email, password_hash,
		       COALESCE(avatar, ''), COALESCE(bio, ''),
		       COALESCE(nickname, ''), COALESCE(location, ''),
		       status, creator_status, verified_status, membership_tier,
		       created_at, updated_at, last_login_at
		FROM app_user WHERE LOWER(username) = LOWER($1)`
	row := r.db.QueryRowContext(ctx, q, username)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, userErr.ErrUserNotFound
	}
	return u, err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (*domuser.User, error) {
	var (
		u             domuser.User
		statusInt     int32
		creatorInt    int32
		verifiedInt   int32
		tierStr       string
		lastLoginAt   sql.NullTime
	)
	if err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Avatar, &u.Bio, &u.Nickname, &u.Location,
		&statusInt, &creatorInt, &verifiedInt, &tierStr,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	); err != nil {
		return nil, err
	}
	u.Status = valueobject.AccountStatus(statusInt)
	u.CreatorStatus = valueobject.CreatorStatus(creatorInt)
	u.VerifiedStatus = valueobject.VerifiedStatus(verifiedInt)
	u.MembershipTier = valueobject.MembershipTier(tierStr)
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		u.LastLoginAt = &t
	}
	return &u, nil
}

// mapUniqueViolation translates a Postgres unique violation into the
// matching domain error so the application layer can react without
// peeking at driver-level types.
func mapUniqueViolation(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		switch {
		case strings.Contains(msg, "uniq_app_user_email"):
			return userErr.ErrEmailTaken
		case strings.Contains(msg, "uniq_app_user_username"):
			return userErr.ErrUsernameTaken
		}
	}
	return fmt.Errorf("user repository: %w", err)
}

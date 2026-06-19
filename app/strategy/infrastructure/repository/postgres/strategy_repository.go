package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
)

// StrategyRepository persists Strategy aggregates in PostgreSQL.
type StrategyRepository struct {
	db *sql.DB
}

// NewStrategyRepository wires the repository against an existing *sql.DB.
func NewStrategyRepository(db *sql.DB) *StrategyRepository {
	return &StrategyRepository{db: db}
}

// Create inserts a new strategy and assigns the generated ID back.
func (r *StrategyRepository) Create(ctx context.Context, s *domstrategy.Strategy) error {
	if s == nil {
		return stratErr.ErrInvalidStrategy
	}
	const stmt = `
		INSERT INTO strategy (
			author_id, title, description, category, tags, current_version_id,
			status, visibility, source_strategy_id,
			view_count, favorite_count, fork_count,
			created_at, updated_at, published_at, archived_at
		) VALUES (
			$1, $2, $3, $4, $5, NULLIF($6, 0),
			$7, $8, NULLIF($9, 0),
			$10, $11, $12,
			$13, $14, $15, $16
		) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, stmt,
		nullableInt64(s.AuthorID),
		s.Title,
		nullableString(s.Description),
		nullableString(s.Category),
		tagsParam(s.Tags),
		s.CurrentVersionID,
		string(s.Status),
		string(s.Visibility),
		s.SourceStrategyID,
		s.ViewCount, s.FavoriteCount, s.ForkCount,
		s.CreatedAt, s.UpdatedAt, s.PublishedAt, s.ArchivedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert strategy: %w", err)
	}
	s.ID = id
	return nil
}

// Update writes back the mutable columns of a strategy. CreatedAt is
// pinned at insert time and never updated.
func (r *StrategyRepository) Update(ctx context.Context, s *domstrategy.Strategy) error {
	if s == nil || s.ID == 0 {
		return stratErr.ErrInvalidStrategy
	}
	const stmt = `
		UPDATE strategy SET
			title              = $2,
			description        = $3,
			category           = $4,
			tags               = $5,
			current_version_id = NULLIF($6, 0),
			status             = $7,
			visibility         = $8,
			source_strategy_id = NULLIF($9, 0),
			view_count         = $10,
			favorite_count     = $11,
			fork_count         = $12,
			updated_at         = $13,
			published_at       = $14,
			archived_at        = $15
		WHERE id = $1`
	res, err := r.db.ExecContext(ctx, stmt,
		s.ID,
		s.Title,
		nullableString(s.Description),
		nullableString(s.Category),
		tagsParam(s.Tags),
		s.CurrentVersionID,
		string(s.Status),
		string(s.Visibility),
		s.SourceStrategyID,
		s.ViewCount, s.FavoriteCount, s.ForkCount,
		s.UpdatedAt, s.PublishedAt, s.ArchivedAt,
	)
	if err != nil {
		return fmt.Errorf("update strategy: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return stratErr.ErrStrategyNotFound
	}
	return nil
}

// Get loads one strategy.
func (r *StrategyRepository) Get(ctx context.Context, id int64) (*domstrategy.Strategy, error) {
	const q = `
		SELECT id, COALESCE(author_id, 0), title, COALESCE(description, ''),
		       COALESCE(category, ''), tags, COALESCE(current_version_id, 0),
		       status, visibility, COALESCE(source_strategy_id, 0),
		       view_count, favorite_count, fork_count,
		       created_at, updated_at, published_at, archived_at
		FROM strategy WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, id)
	s, err := scanStrategy(row)
	if err == sql.ErrNoRows {
		return nil, stratErr.ErrStrategyNotFound
	}
	return s, err
}

// List returns strategies filtered by the supplied query.
func (r *StrategyRepository) List(ctx context.Context, q domstrategy.ListQuery) ([]*domstrategy.Strategy, error) {
	conditions := []string{"1 = 1"}
	args := []any{}
	if q.AuthorID != 0 {
		args = append(args, q.AuthorID)
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", len(args)))
	}
	if q.Status != "" {
		args = append(args, string(q.Status))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if q.Visibility != "" {
		args = append(args, string(q.Visibility))
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", len(args)))
	}
	if q.Category != "" {
		args = append(args, q.Category)
		conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)))
	}
	if q.Tag != "" {
		args = append(args, strings.ToLower(strings.TrimSpace(q.Tag)))
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", len(args)))
	}
	if q.Keyword != "" {
		args = append(args, "%"+strings.ToLower(q.Keyword)+"%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(COALESCE(description, '')) LIKE $%d)", len(args), len(args)))
	}
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args = append(args, limit)
	limitArg := len(args)
	args = append(args, q.Offset)
	offsetArg := len(args)

	orderBy := "created_at DESC"
	switch q.Sort {
	case "favorite_count_desc":
		orderBy = "favorite_count DESC, id DESC"
	case "fork_count_desc":
		orderBy = "fork_count DESC, id DESC"
	case "view_count_desc":
		orderBy = "view_count DESC, id DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, COALESCE(author_id, 0), title, COALESCE(description, ''),
		       COALESCE(category, ''), tags, COALESCE(current_version_id, 0),
		       status, visibility, COALESCE(source_strategy_id, 0),
		       view_count, favorite_count, fork_count,
		       created_at, updated_at, published_at, archived_at
		FROM strategy
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, strings.Join(conditions, " AND "), orderBy, limitArg, offsetArg)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query strategy: %w", err)
	}
	defer rows.Close()
	out := make([]*domstrategy.Strategy, 0)
	for rows.Next() {
		s, err := scanStrategy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// IncrementForkCount bumps the source strategy's fork counter.
func (r *StrategyRepository) IncrementForkCount(ctx context.Context, id int64) error {
	const stmt = `UPDATE strategy SET fork_count = fork_count + 1, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("increment fork count: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return stratErr.ErrStrategyNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanStrategy(row scannable) (*domstrategy.Strategy, error) {
	var (
		s            domstrategy.Strategy
		tags         pq.StringArray
		statusStr    string
		visibility   string
		publishedAt  sql.NullTime
		archivedAt   sql.NullTime
	)
	err := row.Scan(
		&s.ID, &s.AuthorID, &s.Title, &s.Description,
		&s.Category, &tags, &s.CurrentVersionID,
		&statusStr, &visibility, &s.SourceStrategyID,
		&s.ViewCount, &s.FavoriteCount, &s.ForkCount,
		&s.CreatedAt, &s.UpdatedAt, &publishedAt, &archivedAt,
	)
	if err != nil {
		return nil, err
	}
	s.Tags = []string(tags)
	s.Status = valueobject.LifecycleStatus(statusStr)
	s.Visibility = valueobject.Visibility(visibility)
	if publishedAt.Valid {
		t := publishedAt.Time
		s.PublishedAt = &t
	}
	if archivedAt.Valid {
		t := archivedAt.Time
		s.ArchivedAt = &t
	}
	return &s, nil
}

func nullableInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// tagsParam ensures pq sends an empty TEXT[] (not NULL) when the slice is
// nil; the strategy.tags column is NOT NULL with a default of '{}', so a
// zero-value slice from a freshly built aggregate must materialise as
// the empty array rather than NULL.
func tagsParam(in []string) interface{} {
	if in == nil {
		in = []string{}
	}
	return pq.Array(in)
}

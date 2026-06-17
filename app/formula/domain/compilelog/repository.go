package compilelog

import (
	"context"
	"time"
)

// CompileLogRecord represents a single formula compilation log entry.
type CompileLogRecord struct {
	ID            int64     `json:"id"`
	FormulaHash   string    `json:"formula_hash"`
	Formula       string    `json:"formula"`
	Success       bool      `json:"success"`
	ErrorCode     int       `json:"error_code,omitempty"`
	CompileTimeMs int       `json:"compile_time_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

// Repository defines the interface for persisting compile log records.
// Implementations must be safe for concurrent use.
type Repository interface {
	// Save persists a compile log record.
	Save(ctx context.Context, record *CompileLogRecord) error

	// ListByHash retrieves compile log records for a given formula hash, ordered by created_at DESC.
	ListByHash(ctx context.Context, formulaHash string, limit, offset int) ([]CompileLogRecord, error)

	// EnsureTable ensures the underlying storage schema exists.
	EnsureTable(ctx context.Context) error
}

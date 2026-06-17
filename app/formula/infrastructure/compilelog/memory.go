package compilelog

import (
	"context"
	"sync"
	"time"

	domainLog "github.com/agoXQ/QuantLab/app/formula/domain/compilelog"
)

type memoryRepository struct {
	mu      sync.RWMutex
	records []domainLog.CompileLogRecord
	nextID  int64
}

// NewMemoryRepository creates a new in-memory compile log repository (for testing).
func NewMemoryRepository() domainLog.Repository {
	return &memoryRepository{
		records: make([]domainLog.CompileLogRecord, 0),
		nextID:  1,
	}
}

func (r *memoryRepository) EnsureTable(_ context.Context) error {
	return nil // no-op for in-memory
}

func (r *memoryRepository) Save(_ context.Context, record *domainLog.CompileLogRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record.ID = r.nextID
	r.nextID++
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}

	r.records = append(r.records, *record)
	return nil
}

func (r *memoryRepository) ListByHash(_ context.Context, formulaHash string, limit, offset int) ([]domainLog.CompileLogRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var filtered []domainLog.CompileLogRecord
	for _, rec := range r.records {
		if rec.FormulaHash == formulaHash {
			filtered = append(filtered, rec)
		}
	}

	// Reverse to get DESC order
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	start := offset
	if start >= len(filtered) {
		return []domainLog.CompileLogRecord{}, nil
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

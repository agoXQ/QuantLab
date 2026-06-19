// Package fork defines the StrategyFork aggregate, which records the
// "copied from" relationship between two strategies. Forks are stored
// separately from the Strategy table so the lineage tree can be queried
// without dragging the strategy body into the JOIN.
package fork

import (
	"context"
	"time"
)

// StrategyFork records that TargetStrategyID was forked from
// SourceStrategyID by CreatorID.
type StrategyFork struct {
	ID               int64     `json:"id"`
	SourceStrategyID int64     `json:"source_strategy_id"`
	TargetStrategyID int64     `json:"target_strategy_id"`
	CreatorID        int64     `json:"creator_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Repository persists StrategyFork rows.
type Repository interface {
	Create(ctx context.Context, f *StrategyFork) error
	ListBySource(ctx context.Context, sourceID int64, limit int) ([]*StrategyFork, error)
}

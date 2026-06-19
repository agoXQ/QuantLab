// Package version defines the StrategyVersion aggregate.
//
// A StrategyVersion is an immutable snapshot of a strategy's body
// (formula, buy/sell/risk/position/rebalance rules) at a point in time.
// Edits never mutate an existing version; every save lands a new row,
// which gives users diff / rollback semantics for free and lets the
// Backtest Engine reference a stable artefact.
package version

import (
	"context"
	"strings"
	"time"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
)

// StrategyVersion is the aggregate root.
type StrategyVersion struct {
	ID            int64     `json:"id"`
	StrategyID    int64     `json:"strategy_id"`
	VersionNo     string    `json:"version_no"`
	FormulaText   string    `json:"formula_text"`
	BuyRule       string    `json:"buy_rule,omitempty"`
	SellRule      string    `json:"sell_rule,omitempty"`
	RiskRule      string    `json:"risk_rule,omitempty"`
	PositionRule  string    `json:"position_rule,omitempty"`
	RebalanceRule string    `json:"rebalance_rule,omitempty"`
	ChangeLog     string    `json:"change_log,omitempty"`
	CreatedBy     int64     `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Validate ensures the snapshot has the minimum payload to be runnable.
// We require a non-empty FormulaText only; the rule fields are optional
// because the MVP UI lets users iterate body-by-body.
func (v *StrategyVersion) Validate() error {
	if v == nil {
		return stratErr.ErrInvalidVersion
	}
	if strings.TrimSpace(v.FormulaText) == "" {
		return stratErr.ErrFormulaRequired
	}
	if v.StrategyID <= 0 {
		return stratErr.ErrInvalidVersion
	}
	return nil
}

// Repository persists StrategyVersion aggregates.
//
// The repository is append-only on the write side: Create writes a new
// row, Get / ListByStrategy read. The interface deliberately omits an
// Update so a stray caller cannot rewrite history; rolling back to an
// older version is implemented at the application layer by cloning the
// chosen snapshot into a fresh row.
type Repository interface {
	Create(ctx context.Context, v *StrategyVersion) error
	Get(ctx context.Context, id int64) (*StrategyVersion, error)
	ListByStrategy(ctx context.Context, strategyID int64, limit int) ([]*StrategyVersion, error)
	LatestNumber(ctx context.Context, strategyID int64) (int, error)
}

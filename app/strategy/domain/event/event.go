// Package event defines domain events emitted by the Strategy Service.
//
// The envelope mirrors the Backtest / Market Data shape so downstream
// consumers (Ranking, Notification, Community, AI) can decode every
// platform event with one schema.
package event

import (
	"context"
	"time"
)

// EventType lists the canonical Strategy events.
type EventType string

const (
	EventStrategyCreated        EventType = "StrategyCreated"
	EventStrategyUpdated        EventType = "StrategyUpdated"
	EventStrategyVersionCreated EventType = "StrategyVersionCreated"
	EventStrategyPublished      EventType = "StrategyPublished"
	EventStrategyArchived       EventType = "StrategyArchived"
	EventStrategyForked         EventType = "StrategyForked"
)

const (
	AggregateTypeStrategy = "STRATEGY"
	ProducerStrategy      = "strategy-service"
	TopicStrategyEvents   = "strategy-events"
	EventVersionV1        = "1.0"
)

// Event is the canonical envelope used by every Strategy event.
type Event struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Producer      string    `json:"producer"`
	Payload       any       `json:"payload"`
}

// StrategyCreatedPayload is emitted right after a strategy row lands.
type StrategyCreatedPayload struct {
	StrategyID int64  `json:"strategy_id"`
	AuthorID   int64  `json:"author_id,omitempty"`
	Title      string `json:"title"`
}

// StrategyUpdatedPayload covers metadata edits (title / description /
// category / tags). Body changes go through StrategyVersionCreated.
type StrategyUpdatedPayload struct {
	StrategyID int64    `json:"strategy_id"`
	Title      string   `json:"title,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// StrategyVersionCreatedPayload is emitted when a fresh version snapshot
// lands. Carrying the version_no avoids forcing every consumer to
// reload the row to know which revision was published.
type StrategyVersionCreatedPayload struct {
	StrategyID int64  `json:"strategy_id"`
	VersionID  int64  `json:"version_id"`
	VersionNo  string `json:"version_no"`
}

// StrategyPublishedPayload notifies the rest of the platform that a
// version is now public. Ranking subscribes to it so the strategy
// becomes eligible for the leaderboard.
type StrategyPublishedPayload struct {
	StrategyID int64 `json:"strategy_id"`
	VersionID  int64 `json:"version_id"`
	AuthorID   int64 `json:"author_id,omitempty"`
}

// StrategyArchivedPayload notifies consumers a strategy is gone from
// the public surface. Ranking removes the row from any active board.
type StrategyArchivedPayload struct {
	StrategyID int64 `json:"strategy_id"`
}

// StrategyForkedPayload records a fork relationship. Notification can
// alert the source author; Ranking attributes lineage.
type StrategyForkedPayload struct {
	SourceStrategyID int64 `json:"source_strategy_id"`
	TargetStrategyID int64 `json:"target_strategy_id"`
	CreatorID        int64 `json:"creator_id,omitempty"`
}

// Publisher publishes Strategy events.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

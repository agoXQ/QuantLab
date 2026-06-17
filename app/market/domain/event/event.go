// Package event defines domain events emitted by the Market Data service.
package event

import (
	"context"
	"time"
)

// EventType represents the type of a domain event.
type EventType string

const (
	EventMarketDataUpdated  EventType = "MarketDataUpdated"
	EventFactorUpdated      EventType = "FactorUpdated"
	EventDataVersionCreated EventType = "DataVersionCreated"
)

const (
	AggregateTypeSecurity = "SECURITY"
	ProducerMarketService = "market-data-service"
	TopicMarketEvents     = "market-events"
	EventVersionV1        = "1.0"
)

// Event is the canonical envelope for market-data domain events.
//
// Field naming matches the platform-wide Event Specification used by the
// Formula Engine and downstream consumers.
type Event struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	EventVersion  string    `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id,omitempty"`
	Producer      string    `json:"producer"`
	Payload       any       `json:"payload"`
}

// MarketDataUpdatedPayload is the payload for MarketDataUpdated events.
type MarketDataUpdatedPayload struct {
	TradeDate   string `json:"trade_date"`
	Period      string `json:"period"`
	Count       int    `json:"count"`
	DataVersion string `json:"data_version"`
}

// FactorUpdatedPayload is the payload for FactorUpdated events.
type FactorUpdatedPayload struct {
	TradeDate   string `json:"trade_date"`
	Count       int    `json:"count"`
	DataVersion string `json:"data_version"`
}

// DataVersionCreatedPayload is the payload for DataVersionCreated events.
type DataVersionCreatedPayload struct {
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Publisher publishes domain events to the message bus.
//
// Implementations must be safe for concurrent use.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

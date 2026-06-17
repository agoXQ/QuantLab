package event

import "time"

// EventType represents the type of a domain event.
type EventType string

const (
	EventFormulaValidated    EventType = "FormulaValidated"
	EventFormulaCompileFailed EventType = "FormulaCompileFailed"
)

const (
	AggregateTypeStrategy = "STRATEGY"
	ProducerFormulaEngine = "formula-engine"
	TopicStrategyEvents   = "strategy-events"
)

// Event is the base structure for all domain events published by Formula Engine.
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

// FormulaValidatedPayload is the payload for FormulaValidated events.
type FormulaValidatedPayload struct {
	FormulaHash string `json:"formula_hash"`
	Valid       bool   `json:"valid"`
}

// FormulaCompileFailedPayload is the payload for FormulaCompileFailed events.
type FormulaCompileFailedPayload struct {
	FormulaHash string `json:"formula_hash"`
	Error       string `json:"error"`
	ErrorCode   int    `json:"error_code"`
}

// Publisher defines the interface for publishing domain events.
// Implementations must be safe for concurrent use.
type Publisher interface {
	// Publish publishes a domain event to the message bus.
	Publish(event Event) error
}

// Package event provides a Kafka-backed implementation of the Strategy
// domain event publisher and a Noop fallback. Mirrors the shape used
// by Backtest / Market Data so consumers can rely on a single envelope
// structure across services.
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher returns a domain.Publisher backed by Kafka.
func NewKafkaPublisher(brokers []string) domevent.Publisher {
	return &kafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        domevent.TopicStrategyEvents,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
			Async:        true,
		},
	}
}

// Publish implements domevent.Publisher.
func (p *kafkaPublisher) Publish(ctx context.Context, e domevent.Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(e.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(string(e.EventType))},
			{Key: "producer", Value: []byte(e.Producer)},
		},
	}
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.writer.WriteMessages(timeout, msg); err != nil {
		return fmt.Errorf("publish event %s: %w", e.EventType, err)
	}
	return nil
}

// Close releases the Kafka writer.
func (p *kafkaPublisher) Close() error { return p.writer.Close() }

// Noop is a publisher that drops events on the floor; useful in tests
// and when no broker is configured.
type Noop struct{}

// Publish implements domevent.Publisher and never returns an error.
func (Noop) Publish(_ context.Context, _ domevent.Event) error { return nil }

// Package event provides a Kafka-backed implementation of the domain
// event publisher.
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	domainEvent "github.com/agoXQ/QuantLab/app/market/domain/event"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher returns a domain.Publisher backed by Kafka.
func NewKafkaPublisher(brokers []string) domainEvent.Publisher {
	return &kafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        domainEvent.TopicMarketEvents,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
			Async:        true,
		},
	}
}

func (p *kafkaPublisher) Publish(ctx context.Context, e domainEvent.Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(e.EventID),
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
func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}

// Noop is a domain.Publisher that drops events on the floor. It is useful in
// tests and when no broker is configured.
type Noop struct{}

// Publish implements domain.Publisher and never returns an error.
func (Noop) Publish(_ context.Context, _ domainEvent.Event) error { return nil }

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	domainEvent "github.com/agoXQ/QuantLab/app/formula/domain/event"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a new Kafka-backed event publisher.
func NewKafkaPublisher(brokers []string) domainEvent.Publisher {
	return &kafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        domainEvent.TopicStrategyEvents,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
			Async:        true,
		},
	}
}

func (p *kafkaPublisher) Publish(event domainEvent.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.EventID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(string(event.EventType))},
			{Key: "producer", Value: []byte(event.Producer)},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("publish event %s: %w", event.EventType, err)
	}

	return nil
}

// Close closes the underlying Kafka writer.
func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}

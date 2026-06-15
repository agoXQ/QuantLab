// Package kafka provides shared Kafka utilities for all QuantLab services.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// WriterConfig holds Kafka writer configuration.
type WriterConfig struct {
	Brokers string
	Topic   string
}

// Writer wraps a Kafka writer for publishing domain events.
type Writer struct {
	writer *kafka.Writer
}

// NewWriter creates a new Kafka Writer.
func NewWriter(cfg WriterConfig) *Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}
	log.Printf("[kafka] writer created for topic=%s", cfg.Topic)
	return &Writer{writer: w}
}

// Publish sends a domain event to Kafka.
func (w *Writer) Publish(ctx context.Context, key string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return w.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
}

// Close shuts down the Kafka writer.
func (w *Writer) Close() error {
	return w.writer.Close()
}

// ReaderConfig holds Kafka reader configuration.
type ReaderConfig struct {
	Brokers string
	Topic   string
	GroupID string
}

// NewReader creates a new Kafka Reader for consuming events.
func NewReader(cfg ReaderConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Brokers},
		Topic:       cfg.Topic,
		GroupID:     cfg.GroupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
}

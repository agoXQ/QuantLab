// Package backtestsync provides the Kafka-backed adapter that consumes
// Backtest events and dispatches them to a domain Handler.
package backtestsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	domsync "github.com/agoXQ/QuantLab/app/notification/domain/backtestsync"
)

// DefaultTopic mirrors backtest/domain/event.TopicBacktestEvents.
const DefaultTopic = "backtest-events"

// DefaultGroup is the consumer group id Notification uses.
const DefaultGroup = "notification-backtest-sync"

// Config configures the Kafka consumer.
type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Consumer drains a Kafka topic, decodes Backtest events, and
// forwards them to the supplied Handler.
type Consumer struct {
	reader  *kafka.Reader
	handler domsync.Handler
}

// NewConsumer wires the consumer; brokers + handler must be non-nil.
func NewConsumer(cfg Config, handler domsync.Handler) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("backtestsync: brokers required")
	}
	if handler == nil {
		return nil, fmt.Errorf("backtestsync: handler required")
	}
	topic := cfg.Topic
	if topic == "" {
		topic = DefaultTopic
	}
	group := cfg.GroupID
	if group == "" {
		group = DefaultGroup
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          topic,
		GroupID:        group,
		MinBytes:       1,
		MaxBytes:       10 << 20,
		CommitInterval: time.Second,
	})
	return &Consumer{reader: reader, handler: handler}, nil
}

// Run drains the reader until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.reader == nil {
		return fmt.Errorf("backtestsync: consumer not initialised")
	}
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("backtestsync: read message: %w", err)
		}
		if err := dispatch(ctx, c.handler, msg.Value); err != nil {
			log.Printf("[backtestsync] dispatch event: %v", err)
		}
	}
}

// Close releases the underlying kafka reader.
func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

// Dispatch is the exported counterpart of dispatch, useful for tests.
func Dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	return dispatch(ctx, handler, raw)
}

func dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	var env domsync.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	switch env.EventType {
	case domsync.EventBacktestFinished:
		var p domsync.FinishedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode finished payload: %w", err)
		}
		return handler.OnFinished(ctx, env, p)
	case domsync.EventBacktestFailed:
		var p domsync.FailedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode failed payload: %w", err)
		}
		return handler.OnFailed(ctx, env, p)
	default:
		return nil
	}
}

func remarshal(in any, out any) error {
	buf, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}

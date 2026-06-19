// Package strategysync provides the Kafka-backed adapter that consumes
// Strategy events and dispatches them to a domain Handler. The shape
// mirrors strategy/infrastructure/backtestsync verbatim so the cross-
// service integration paths stay consistent.
package strategysync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	domsync "github.com/agoXQ/QuantLab/app/user/domain/strategysync"
)

// DefaultTopic mirrors strategy/domain/event.TopicStrategyEvents.
const DefaultTopic = "strategy-events"

// DefaultGroup is the consumer group id User uses; one group per
// service so each service maintains its own offsets.
const DefaultGroup = "user-strategy-sync"

// Config configures the Kafka consumer.
type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Consumer drains a Kafka topic, decodes Strategy events, and
// forwards them to the supplied Handler.
type Consumer struct {
	reader  *kafka.Reader
	handler domsync.Handler
}

// NewConsumer wires the consumer; brokers + handler must be non-nil.
func NewConsumer(cfg Config, handler domsync.Handler) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("strategysync: brokers required")
	}
	if handler == nil {
		return nil, fmt.Errorf("strategysync: handler required")
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

// Run drains the reader until ctx is cancelled. Decode failures and
// handler failures are logged and the message is committed anyway: the
// MVP prefers progress over retrying noisy events forever.
func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.reader == nil {
		return fmt.Errorf("strategysync: consumer not initialised")
	}
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("strategysync: read message: %w", err)
		}
		if err := dispatch(ctx, c.handler, msg.Value); err != nil {
			log.Printf("[strategysync] dispatch event: %v", err)
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

// Dispatch is the exported counterpart of dispatch, useful for tests
// and any future adapter that wants to drive routing without going
// through Kafka.
func Dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	return dispatch(ctx, handler, raw)
}

func dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	var env domsync.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	switch env.EventType {
	case domsync.EventStrategyCreated:
		var p domsync.CreatedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode created payload: %w", err)
		}
		return handler.OnCreated(ctx, env, p)
	case domsync.EventStrategyArchived:
		var p domsync.ArchivedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode archived payload: %w", err)
		}
		return handler.OnArchived(ctx, env, p)
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

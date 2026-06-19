// Package strategysync provides the Kafka-backed adapter that consumes
// Strategy lifecycle events and dispatches them to a Handler. The
// adapter only depends on the domain Envelope / Handler types so the
// same code drives unit tests (with an in-memory queue) and production
// (with kafka-go reader).
package strategysync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
)

// DefaultTopic is the canonical Strategy events topic. Mirrors
// strategy-service domain.event.TopicStrategyEvents but re-declared to
// avoid a build-time dependency on the sibling service's package.
const DefaultTopic = "strategy-events"

// DefaultGroup is the consumer group id Backtest uses; one group per
// service so each service maintains its own offsets.
const DefaultGroup = "backtest-strategy-sync"

// Config configures the Kafka consumer. Brokers + Topic are required;
// the rest carry sensible defaults for the MVP.
type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Consumer drains a Kafka topic, decodes Strategy events, and forwards
// them to the supplied Handler. Lifecycle: NewConsumer wires it,
// Run(ctx) blocks until the context is done or the reader errors out;
// Close drains the underlying reader.
type Consumer struct {
	reader  *kafka.Reader
	handler domsync.Handler
}

// NewConsumer wires the consumer; brokers must be non-empty.
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
// MVP prefers progress over retrying noisy events forever, since the
// Strategy aggregate itself is the source of truth and any consumer
// can rebuild state by reading the strategy row directly.
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

// dispatch decodes the envelope and routes the message to the right
// handler method. Exported as a free function so tests can drive the
// dispatch path without spinning up Kafka.
func dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	var env domsync.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	switch env.EventType {
	case domsync.EventStrategyPublished:
		var p domsync.PublishedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode published payload: %w", err)
		}
		return handler.OnPublished(ctx, env, p)
	case domsync.EventStrategyVersionCreated:
		var p domsync.VersionCreatedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode version-created payload: %w", err)
		}
		return handler.OnVersionCreated(ctx, env, p)
	default:
		// Other Strategy events (Created / Updated / Archived / Forked)
		// are intentionally ignored here. Future handlers can extend
		// the switch without changing the consumer surface.
		return nil
	}
}

// Dispatch is the exported counterpart of dispatch, useful for tests
// and for any future adapter (e.g. an in-process bus) that wants to
// drive the same routing without going through Kafka.
func Dispatch(ctx context.Context, handler domsync.Handler, raw []byte) error {
	return dispatch(ctx, handler, raw)
}

// remarshal converts the generic payload map back into the typed
// payload shape. JSON is faster than reflect for this volume and keeps
// the consumer free of struct tags duplicated from the strategy
// service.
func remarshal(in any, out any) error {
	buf, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}

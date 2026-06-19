// Package usersync provides the Kafka-backed adapter that consumes
// User events and dispatches them to a domain Handler. The shape
// mirrors user/infrastructure/strategysync verbatim so the cross-
// service integration paths stay consistent.
package usersync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	domsync "github.com/agoXQ/QuantLab/app/notification/domain/usersync"
)

// DefaultTopic mirrors user/domain/event.TopicUserEvents.
const DefaultTopic = "user-events"

// DefaultGroup is the consumer group id Notification uses; one group
// per service so each service maintains its own offsets.
const DefaultGroup = "notification-user-sync"

// Config configures the Kafka consumer.
type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Consumer drains a Kafka topic, decodes User events, and forwards
// them to the supplied Handler.
type Consumer struct {
	reader  *kafka.Reader
	handler domsync.Handler
}

// NewConsumer wires the consumer; brokers + handler must be non-nil.
func NewConsumer(cfg Config, handler domsync.Handler) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("usersync: brokers required")
	}
	if handler == nil {
		return nil, fmt.Errorf("usersync: handler required")
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
		return fmt.Errorf("usersync: consumer not initialised")
	}
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("usersync: read message: %w", err)
		}
		if err := dispatch(ctx, c.handler, msg.Value); err != nil {
			log.Printf("[usersync] dispatch event: %v", err)
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
	case domsync.EventUserRegistered:
		var p domsync.RegisteredPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode registered payload: %w", err)
		}
		return handler.OnRegistered(ctx, env, p)
	case domsync.EventUserFollowed:
		var p domsync.FollowedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode followed payload: %w", err)
		}
		return handler.OnFollowed(ctx, env, p)
	case domsync.EventUserUnfollowed:
		var p domsync.UnfollowedPayload
		if err := remarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode unfollowed payload: %w", err)
		}
		return handler.OnUnfollowed(ctx, env, p)
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

package svc

import (
	"context"
	"log"
	"time"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	appBacktestSync "github.com/agoXQ/QuantLab/app/notification/application/backtestsync"
	appStrategySync "github.com/agoXQ/QuantLab/app/notification/application/strategysync"
	appUserSync "github.com/agoXQ/QuantLab/app/notification/application/usersync"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	infraBacktestSync "github.com/agoXQ/QuantLab/app/notification/infrastructure/backtestsync"
	infraStrategySync "github.com/agoXQ/QuantLab/app/notification/infrastructure/strategysync"
	infraUserSync "github.com/agoXQ/QuantLab/app/notification/infrastructure/usersync"
	"github.com/agoXQ/QuantLab/app/notification/internal/config"
)

// syncRunner is the small lifecycle owner returned by the buildXxx
// helpers. It bundles the start function, the context cancel, the
// done channel, and the consumer Closer so the ServiceContext can
// shut everything down deterministically.
type syncRunner struct {
	start  func()
	cancel context.CancelFunc
	done   chan struct{}
	closer interface{ Close() error }
}

func buildUserSync(c config.Config, svc appNotif.Service, subs domSub.Repository) *syncRunner {
	if !c.UserSync.Enabled {
		return nil
	}
	if len(c.UserSync.Brokers) == 0 {
		log.Printf("[notification] user sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appUserSync.NewHandler(svc, subs, time.Now)
	consumer, err := infraUserSync.NewConsumer(infraUserSync.Config{
		Brokers: c.UserSync.Brokers,
		Topic:   c.UserSync.Topic,
		GroupID: c.UserSync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[notification] user sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[notification] user sync started topic=%s group=%s",
					orDefault(c.UserSync.Topic, infraUserSync.DefaultTopic),
					orDefault(c.UserSync.GroupID, infraUserSync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[notification] user sync stopped: %v", err)
				}
			}()
		},
		cancel: cancel,
		done:   doneCh,
		closer: consumer,
	}
}

func buildStrategySync(c config.Config, svc appNotif.Service, subs domSub.Repository) *syncRunner {
	if !c.StrategySync.Enabled {
		return nil
	}
	if len(c.StrategySync.Brokers) == 0 {
		log.Printf("[notification] strategy sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appStrategySync.NewHandler(svc, subs)
	consumer, err := infraStrategySync.NewConsumer(infraStrategySync.Config{
		Brokers: c.StrategySync.Brokers,
		Topic:   c.StrategySync.Topic,
		GroupID: c.StrategySync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[notification] strategy sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[notification] strategy sync started topic=%s group=%s",
					orDefault(c.StrategySync.Topic, infraStrategySync.DefaultTopic),
					orDefault(c.StrategySync.GroupID, infraStrategySync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[notification] strategy sync stopped: %v", err)
				}
			}()
		},
		cancel: cancel,
		done:   doneCh,
		closer: consumer,
	}
}

func buildBacktestSync(c config.Config, svc appNotif.Service) *syncRunner {
	if !c.BacktestSync.Enabled {
		return nil
	}
	if len(c.BacktestSync.Brokers) == 0 {
		log.Printf("[notification] backtest sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appBacktestSync.NewHandler(svc)
	consumer, err := infraBacktestSync.NewConsumer(infraBacktestSync.Config{
		Brokers: c.BacktestSync.Brokers,
		Topic:   c.BacktestSync.Topic,
		GroupID: c.BacktestSync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[notification] backtest sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[notification] backtest sync started topic=%s group=%s",
					orDefault(c.BacktestSync.Topic, infraBacktestSync.DefaultTopic),
					orDefault(c.BacktestSync.GroupID, infraBacktestSync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[notification] backtest sync stopped: %v", err)
				}
			}()
		},
		cancel: cancel,
		done:   doneCh,
		closer: consumer,
	}
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

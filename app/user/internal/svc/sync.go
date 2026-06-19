package svc

import (
	"context"
	"log"

	appBacktestSync "github.com/agoXQ/QuantLab/app/user/application/backtestsync"
	appStrategySync "github.com/agoXQ/QuantLab/app/user/application/strategysync"
	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	infraBacktestSync "github.com/agoXQ/QuantLab/app/user/infrastructure/backtestsync"
	infraStrategySync "github.com/agoXQ/QuantLab/app/user/infrastructure/strategysync"
	"github.com/agoXQ/QuantLab/app/user/internal/config"
)

// syncRunner is the small lifecycle owner returned by buildStrategySync /
// buildBacktestSync; it bundles the start function, the context cancel,
// the done channel and the consumer Closer so the ServiceContext can
// shut everything down deterministically.
type syncRunner struct {
	start  func()
	cancel context.CancelFunc
	done   chan struct{}
	closer interface{ Close() error }
}

func buildStrategySync(c config.Config, svc appUser.Service) *syncRunner {
	if !c.StrategySync.Enabled {
		return nil
	}
	if len(c.StrategySync.Brokers) == 0 {
		log.Printf("[user] strategy sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appStrategySync.NewCounterHandler(svc)
	consumer, err := infraStrategySync.NewConsumer(infraStrategySync.Config{
		Brokers: c.StrategySync.Brokers,
		Topic:   c.StrategySync.Topic,
		GroupID: c.StrategySync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[user] strategy sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[user] strategy sync started topic=%s group=%s",
					orDefault(c.StrategySync.Topic, infraStrategySync.DefaultTopic),
					orDefault(c.StrategySync.GroupID, infraStrategySync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[user] strategy sync stopped: %v", err)
				}
			}()
		},
		cancel: cancel,
		done:   doneCh,
		closer: consumer,
	}
}

func buildBacktestSync(c config.Config, svc appUser.Service) *syncRunner {
	if !c.BacktestSync.Enabled {
		return nil
	}
	if len(c.BacktestSync.Brokers) == 0 {
		log.Printf("[user] backtest sync enabled but no Kafka brokers; skipping")
		return nil
	}
	handler := appBacktestSync.NewCounterHandler(svc)
	consumer, err := infraBacktestSync.NewConsumer(infraBacktestSync.Config{
		Brokers: c.BacktestSync.Brokers,
		Topic:   c.BacktestSync.Topic,
		GroupID: c.BacktestSync.GroupID,
	}, handler)
	if err != nil {
		log.Printf("[user] backtest sync: build consumer: %v", err)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	return &syncRunner{
		start: func() {
			go func() {
				defer close(doneCh)
				log.Printf("[user] backtest sync started topic=%s group=%s",
					orDefault(c.BacktestSync.Topic, infraBacktestSync.DefaultTopic),
					orDefault(c.BacktestSync.GroupID, infraBacktestSync.DefaultGroup))
				if err := consumer.Run(ctx); err != nil {
					log.Printf("[user] backtest sync stopped: %v", err)
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

// Package queue defines the domain-level port for asynchronous backtest
// execution. The aggregate root (BacktestJob) is persisted by the job
// repository; this port carries job IDs only, decoupling the application
// service from the chosen transport (in-memory channel today, Kafka /
// Redis Streams later).
//
// Keeping the contract domain-side means the application service depends
// on a single abstraction regardless of transport, and the infrastructure
// layer can grow new adapters without rippling through use cases.
package queue

import (
	"context"
	"errors"
)

// ErrQueueClosed is returned by Enqueue/Dequeue once the queue is shut
// down. Adapters should map their transport-specific shutdown signals to
// this sentinel so callers can react uniformly.
var ErrQueueClosed = errors.New("backtest queue: closed")

// Job is the lightweight envelope carried through the queue. We do not
// serialise the whole BacktestJob aggregate; the worker re-loads it from
// the repository so it always sees the latest status (e.g. a cancellation
// that landed between submit and pickup).
type Job struct {
	ID int64
}

// Queue is the asynchronous transport for queued backtest jobs.
//
// Implementations must guarantee:
//
//   - Enqueue returns nil only after the job has been accepted for
//     delivery to a worker; transport errors propagate to the caller so
//     the application service can fall back (e.g. mark the job FAILED).
//   - Dequeue blocks until either a job is available, the supplied
//     context is cancelled, or the queue is closed (returns
//     ErrQueueClosed).
//   - Close is idempotent and safe to call concurrently with active
//     consumers; it must unblock any in-flight Dequeue.
//
// At-least-once delivery is the MVP target: in-memory adapter is exactly
// once because it has a single in-process consumer set, but durable
// adapters (Kafka) will deliver duplicates after a crash. The application
// service must remain idempotent on Run (it is: a terminal job is
// short-circuited by collectRunResult).
type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	Dequeue(ctx context.Context) (Job, error)
	Close() error
}

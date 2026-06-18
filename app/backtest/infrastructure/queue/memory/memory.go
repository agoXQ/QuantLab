// Package memory provides an in-process channel-backed implementation of
// the backtest queue port. It is intended for the MVP single-binary
// deployment and for tests; durable adapters (Kafka / Redis Streams)
// will live alongside it once the producer / worker contract has stabilised.
//
// The implementation is intentionally small:
//
//   - one buffered channel of Job, sized at construction;
//   - Enqueue blocks (respecting the caller's ctx) until there is room or
//     the queue is closed;
//   - Dequeue blocks on the channel and the consumer's ctx.
//
// On shutdown we close the channel: outstanding consumers receive the
// zero value and a closed signal that we surface as ErrQueueClosed.
// In-flight items left in the channel are dropped; the design TODO is to
// reconcile QUEUED jobs after restart from the repository, which the
// backtest service can do at boot once the durable adapter exists.
package memory

import (
	"context"
	"sync"

	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"
)

// DefaultBufferSize is used when callers ask for a non-positive buffer.
// It is generous enough to soak short bursts without blocking the HTTP
// handler, but small enough that a stuck worker pool does not hide
// systemic failures behind an unbounded queue.
const DefaultBufferSize = 256

// Queue is the in-memory implementation of domqueue.Queue.
type Queue struct {
	ch     chan domqueue.Job
	mu     sync.Mutex
	closed bool
}

// New returns a queue with the given buffer size. Non-positive values
// fall back to DefaultBufferSize.
func New(buffer int) *Queue {
	if buffer <= 0 {
		buffer = DefaultBufferSize
	}
	return &Queue{ch: make(chan domqueue.Job, buffer)}
}

// Enqueue adds the job to the queue. The call respects ctx cancellation
// and surfaces ErrQueueClosed once Close has run.
func (q *Queue) Enqueue(ctx context.Context, job domqueue.Job) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return domqueue.ErrQueueClosed
	}
	q.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.ch <- job:
		return nil
	}
}

// Dequeue blocks until a job is available, ctx is done, or the queue is
// closed. The closed signal returns ErrQueueClosed so callers do not
// confuse it with a delivered zero job.
func (q *Queue) Dequeue(ctx context.Context) (domqueue.Job, error) {
	select {
	case <-ctx.Done():
		return domqueue.Job{}, ctx.Err()
	case job, ok := <-q.ch:
		if !ok {
			return domqueue.Job{}, domqueue.ErrQueueClosed
		}
		return job, nil
	}
}

// Close shuts the queue. Subsequent Enqueue calls return ErrQueueClosed;
// any already-buffered jobs are dropped on shutdown but a future durable
// adapter can replay them from the repository (status=QUEUED) on restart.
// Close is idempotent.
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true
	close(q.ch)
	return nil
}

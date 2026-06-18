// Package worker implements a small worker pool that consumes queued
// BacktestJob IDs from a domqueue.Queue and runs them through the
// application Service.
//
// The pool is intentionally framework-free: it owns a context, spins up
// N goroutines, and stops them by cancelling the context and waiting for
// the WaitGroup. The application service is responsible for state
// transitions (MarkRunning / MarkCompleted / MarkFailed) and for
// idempotency on a re-delivered job; the pool only cares about plumbing
// jobs from the queue to the service while keeping the process alive
// across panics in user-supplied formulas.
package worker

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"

	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"
)

// Runner is the application-side hook the pool calls per dequeued job.
// We avoid importing the application package here so the dependency
// arrow points the conventional direction (infrastructure -> domain) and
// the application layer drives composition through this small adapter.
type Runner interface {
	RunQueued(ctx context.Context, jobID int64) error
}

// Config controls the pool sizing and timing knobs.
type Config struct {
	// Workers is the number of concurrent goroutines draining the queue.
	// Defaults to 2 when the supplied value is non-positive.
	Workers int
	// JobTimeout caps the runtime of a single job. Zero disables the
	// timeout, which is the safer default for the MVP because backtests
	// can legitimately take several minutes.
	JobTimeout time.Duration
	// Logger is the destination for lifecycle messages. Defaults to the
	// standard logger; tests can swap in a no-op.
	Logger *log.Logger
}

// Pool is a tiny worker pool tied to one queue.
type Pool struct {
	cfg    Config
	queue  domqueue.Queue
	runner Runner

	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

// New returns a Pool ready to be started.
func New(queue domqueue.Queue, runner Runner, cfg Config) *Pool {
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Pool{cfg: cfg, queue: queue, runner: runner}
}

// Start spins up the workers. It is safe to call once; subsequent calls
// are no-ops because the parent ctx for the goroutines is captured the
// first time around.
func (p *Pool) Start(ctx context.Context) {
	p.once.Do(func() {
		ctx, p.cancel = context.WithCancel(ctx)
		for i := 0; i < p.cfg.Workers; i++ {
			p.wg.Add(1)
			go p.run(ctx, i)
		}
	})
}

// Stop cancels the workers and blocks until they have all returned.
// Closing the queue is the caller's responsibility; if the queue is
// already closed, the workers exit on the closed channel.
func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}

func (p *Pool) run(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		job, err := p.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, domqueue.ErrQueueClosed) {
				return
			}
			p.cfg.Logger.Printf("[backtest worker %d] dequeue error: %v", id, err)
			// Brief backoff so a transient error does not spin the CPU.
			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
			}
			continue
		}
		p.handle(ctx, id, job.ID)
	}
}

// handle wraps a single job execution with panic recovery so a bug in
// the formula or executor does not take the whole worker down.
func (p *Pool) handle(parent context.Context, id int, jobID int64) {
	defer func() {
		if rec := recover(); rec != nil {
			p.cfg.Logger.Printf("[backtest worker %d] panic running job %d: %v\n%s",
				id, jobID, rec, debug.Stack())
		}
	}()

	ctx := parent
	if p.cfg.JobTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(parent, p.cfg.JobTimeout)
		defer cancel()
	}

	if err := p.runner.RunQueued(ctx, jobID); err != nil {
		// The runner is expected to write FAILED back to the repository
		// itself; we only log here so operators see worker-level failures
		// in the same place as queue errors.
		p.cfg.Logger.Printf("[backtest worker %d] job %d failed: %v", id, jobID, err)
	}
}

package backtest

import (
	"context"
	"fmt"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// ReconcileResult summarises what Reconcile did at boot.
type ReconcileResult struct {
	Requeued    int
	FailedStuck int
	Inspected   int
}

// Reconcile recovers from an ungraceful shutdown. Any job that the
// previous run left in QUEUED is re-enqueued so a worker can pick it
// up; jobs left in RUNNING are flipped to FAILED with a deterministic
// reason because the in-process state needed to resume them
// (portfolio, calendar pointer) is gone.
//
// The caller is expected to invoke Reconcile once at startup, after the
// queue + worker pool are wired but before the HTTP / RPC servers
// accept new traffic. It is safe to call Reconcile when the queue is
// nil; the function then only flips RUNNING jobs to FAILED and skips
// the requeue branch.
func (s *service) Reconcile(ctx context.Context) (ReconcileResult, error) {
	res := ReconcileResult{}
	stuck, err := s.collectStuck(ctx)
	if err != nil {
		return res, err
	}
	res.Inspected = len(stuck)
	for _, job := range stuck {
		switch job.Status {
		case valueobject.JobStatusQueued:
			if s.deps.Queue == nil {
				continue
			}
			if err := s.deps.Queue.Enqueue(ctx, domqueue.Job{ID: job.ID}); err != nil {
				// Queue refused (closed or full); nothing more we can
				// do here, the row stays in QUEUED for the next boot.
				return res, fmt.Errorf("requeue job %d: %w", job.ID, err)
			}
			res.Requeued++
		case valueobject.JobStatusRunning:
			if err := job.MarkFailed(s.deps.Clock(), "interrupted by service restart"); err != nil {
				return res, fmt.Errorf("mark stuck job %d failed: %w", job.ID, err)
			}
			if err := s.deps.Jobs.Update(ctx, job); err != nil {
				return res, fmt.Errorf("persist stuck job %d: %w", job.ID, err)
			}
			res.FailedStuck++
		}
	}
	return res, nil
}

// collectStuck loads jobs left in non-terminal lifecycle states by the
// previous run. We deliberately scan QUEUED first, then RUNNING, so the
// recovery order matches the natural lifecycle.
func (s *service) collectStuck(ctx context.Context) ([]*backtestjob.BacktestJob, error) {
	const lim = 500
	queued, err := s.deps.Jobs.List(ctx, backtestjob.ListQuery{
		Status: valueobject.JobStatusQueued,
		Limit:  lim,
	})
	if err != nil {
		return nil, err
	}
	running, err := s.deps.Jobs.List(ctx, backtestjob.ListQuery{
		Status: valueobject.JobStatusRunning,
		Limit:  lim,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*backtestjob.BacktestJob, 0, len(queued)+len(running))
	out = append(out, queued...)
	out = append(out, running...)
	return out, nil
}

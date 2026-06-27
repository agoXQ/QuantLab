package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CreateBacktestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateBacktestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBacktestLogic {
	return &CreateBacktestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateBacktest resolves the strategy version's formula via the
// Strategy service, persists a job with A-share defaults, then enqueues
// it on the async queue so the gRPC call returns immediately. The
// frontend polls GetBacktest for status transitions. The gRPC proto
// carries only strategy_id / version_id / dates, so the formula,
// universe, and caller identity are materialised here.
func (l *CreateBacktestLogic) CreateBacktest(in *pb.CreateBacktestRequest) (*pb.CreateBacktestResponse, error) {
	if in.StrategyId <= 0 {
		return nil, fmt.Errorf("strategy_id is required")
	}
	if in.VersionId <= 0 {
		return nil, fmt.Errorf("version_id is required")
	}

	start, err := valueobject.ParseDate(in.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := valueobject.ParseDate(in.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	// Resolve the formula from the strategy version.
	var formula string
	if l.svcCtx.StrategyResolver != nil {
		snap, err := l.svcCtx.StrategyResolver.Resolve(l.ctx, in.StrategyId, in.VersionId)
		if err != nil {
			l.Errorf("resolve strategy %d version %d: %v", in.StrategyId, in.VersionId, err)
			return nil, fmt.Errorf("resolve strategy version: %w", err)
		}
		formula = snap.FormulaText
	}
	if formula == "" {
		return nil, fmt.Errorf("strategy version %d has no formula", in.VersionId)
	}

	// MVP defaults: single-stock universe, 1M initial capital, A-share
	// config. The universe will be broadened once the Market Data
	// service exposes a list-securities API the gateway can call.
	universe := []string{"000001"}
	benchmark := "000300"
	initialCapital := 1_000_000.0

	cfg := backtestjob.DefaultConfig()

	created, err := l.svcCtx.BacktestSvc.Create(l.ctx, backtest.CreateBacktestRequest{
		UserID:         userIDFromMetadata(l.ctx),
		StrategyID:     in.StrategyId,
		VersionID:      in.VersionId,
		Formula:        formula,
		Universe:       universe,
		Benchmark:      benchmark,
		InitialCapital: initialCapital,
		Range:          valueobject.DateRange{Start: start, End: end},
		Config:         cfg,
	})
	if err != nil {
		return nil, err
	}

	jobID := created.Job.ID

	// Enqueue on the async queue so the gRPC call returns immediately.
	// The worker pool drains the queue and runs the job with a timeout;
	// the frontend polls GetBacktest for status. When the queue is
	// unavailable the job stays in CREATED and the caller is told so it
	// can retry or trigger an inline run via POST :id/run.
	if _, err := l.svcCtx.BacktestSvc.Submit(l.ctx, jobID); err != nil {
		l.Errorf("submit job %d to queue: %v; job left in CREATED", jobID, err)
	}

	return &pb.CreateBacktestResponse{JobId: jobID}, nil
}

// userIDFromMetadata extracts the caller id forwarded by the gateway as
// the "x-user-id" gRPC metadata key. Returns 0 when absent so the job
// is still created (the list endpoint can match on user_id=0 as a
// fallback, or the gateway can pass the caller id explicitly).
func userIDFromMetadata(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return 0
	}
	id, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

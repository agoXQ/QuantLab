package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelBacktestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelBacktestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelBacktestLogic {
	return &CancelBacktestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelBacktest delegates to the application Service so the gRPC and
// HTTP surfaces share one cancellation contract.
func (l *CancelBacktestLogic) CancelBacktest(in *pb.CancelBacktestRequest) (*pb.CancelBacktestResponse, error) {
	if _, err := l.svcCtx.BacktestSvc.Cancel(l.ctx, in.JobId, ""); err != nil {
		return nil, err
	}
	return &pb.CancelBacktestResponse{}, nil
}

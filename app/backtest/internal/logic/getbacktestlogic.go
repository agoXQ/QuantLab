package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBacktestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBacktestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBacktestLogic {
	return &GetBacktestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetBacktest returns the job aggregate and its config.
func (l *GetBacktestLogic) GetBacktest(in *pb.GetBacktestRequest) (*pb.GetBacktestResponse, error) {
	job, err := l.svcCtx.BacktestSvc.Get(l.ctx, in.JobId)
	if err != nil {
		return nil, err
	}
	return &pb.GetBacktestResponse{
		Job: jobToProto(job),
		Config: &pb.BacktestConfig{
			JobId:            job.ID,
			CommissionRate:   job.Config.CommissionRate,
			SlippageRate:     job.Config.SlippageRate,
			TaxRate:          job.Config.StampDutyRate,
			RebalancePeriod:  string(job.Config.RebalanceFrequency),
			MaxPositionCount: int32(job.Config.MaxPositionCount),
		},
	}, nil
}

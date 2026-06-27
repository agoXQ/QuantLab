package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReportLogic {
	return &GetReportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetReport returns the performance report for a finished job.
func (l *GetReportLogic) GetReport(in *pb.GetReportRequest) (*pb.GetReportResponse, error) {
	rep, err := l.svcCtx.BacktestSvc.GetReport(l.ctx, in.JobId)
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return &pb.GetReportResponse{}, nil
	}
	return &pb.GetReportResponse{
		Report: &pb.PerformanceReport{
			JobId:         rep.JobID,
			AnnualReturn:  rep.AnnualReturn,
			TotalReturn:   rep.TotalReturn,
			SharpeRatio:   rep.SharpeRatio,
			MaxDrawdown:   rep.MaxDrawdown,
			WinRate:       rep.WinRate,
			Volatility:    rep.Volatility,
		},
	}, nil
}

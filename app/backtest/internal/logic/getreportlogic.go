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

func (l *GetReportLogic) GetReport(in *pb.GetReportRequest) (*pb.GetReportResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetReportResponse{}, nil
}

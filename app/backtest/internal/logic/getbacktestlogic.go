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

func (l *GetBacktestLogic) GetBacktest(in *pb.GetBacktestRequest) (*pb.GetBacktestResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetBacktestResponse{}, nil
}

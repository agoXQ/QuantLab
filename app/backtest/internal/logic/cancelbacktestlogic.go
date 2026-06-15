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

func (l *CancelBacktestLogic) CancelBacktest(in *pb.CancelBacktestRequest) (*pb.CancelBacktestResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CancelBacktestResponse{}, nil
}

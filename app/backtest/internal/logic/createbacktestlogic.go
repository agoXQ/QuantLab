package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *CreateBacktestLogic) CreateBacktest(in *pb.CreateBacktestRequest) (*pb.CreateBacktestResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CreateBacktestResponse{}, nil
}

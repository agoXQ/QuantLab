package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradesLogic {
	return &GetTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTradesLogic) GetTrades(in *pb.GetTradesRequest) (*pb.GetTradesResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetTradesResponse{}, nil
}

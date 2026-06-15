package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionsLogic {
	return &GetPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPositionsLogic) GetPositions(in *pb.GetPositionsRequest) (*pb.GetPositionsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetPositionsResponse{}, nil
}

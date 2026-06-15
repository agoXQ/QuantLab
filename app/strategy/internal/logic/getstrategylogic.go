package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStrategyLogic {
	return &GetStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetStrategyLogic) GetStrategy(in *pb.GetStrategyRequest) (*pb.GetStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetStrategyResponse{}, nil
}

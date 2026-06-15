package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForkStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewForkStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForkStrategyLogic {
	return &ForkStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ForkStrategyLogic) ForkStrategy(in *pb.ForkStrategyRequest) (*pb.ForkStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ForkStrategyResponse{}, nil
}

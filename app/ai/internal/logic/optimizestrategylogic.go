package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OptimizeStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOptimizeStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptimizeStrategyLogic {
	return &OptimizeStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OptimizeStrategyLogic) OptimizeStrategy(in *pb.OptimizeStrategyRequest) (*pb.OptimizeStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.OptimizeStrategyResponse{}, nil
}

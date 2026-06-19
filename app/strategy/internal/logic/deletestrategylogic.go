package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteStrategyLogic {
	return &DeleteStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteStrategyLogic) DeleteStrategy(in *pb.DeleteStrategyRequest) (*pb.DeleteStrategyResponse, error) {
	if err := l.svcCtx.StrategySvc.Delete(l.ctx, in.StrategyId, 0); err != nil {
		return nil, err
	}
	return &pb.DeleteStrategyResponse{}, nil
}

package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
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

// ForkStrategy snapshots the source strategy + its current version into
// a new aggregate via the application service.
func (l *ForkStrategyLogic) ForkStrategy(in *pb.ForkStrategyRequest) (*pb.ForkStrategyResponse, error) {
	res, err := l.svcCtx.StrategySvc.Fork(l.ctx, appStrategy.ForkRequest{
		SourceStrategyID: in.SourceStrategyId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ForkStrategyResponse{NewStrategyId: res.Strategy.ID}, nil
}

package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishStrategyLogic {
	return &PublishStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PublishStrategy hands the request to the application service so the
// HTTP and gRPC surfaces share state-transition rules.
func (l *PublishStrategyLogic) PublishStrategy(in *pb.PublishStrategyRequest) (*pb.PublishStrategyResponse, error) {
	if _, err := l.svcCtx.StrategySvc.Publish(l.ctx, appStrategy.PublishRequest{
		StrategyID: in.StrategyId,
		VersionID:  in.VersionId,
	}); err != nil {
		return nil, err
	}
	return &pb.PublishStrategyResponse{}, nil
}

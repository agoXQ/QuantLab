package logic

import (
	"context"

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

func (l *PublishStrategyLogic) PublishStrategy(in *pb.PublishStrategyRequest) (*pb.PublishStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.PublishStrategyResponse{}, nil
}

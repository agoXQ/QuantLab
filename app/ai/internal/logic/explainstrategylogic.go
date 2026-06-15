package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExplainStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExplainStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExplainStrategyLogic {
	return &ExplainStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExplainStrategyLogic) ExplainStrategy(in *pb.ExplainStrategyRequest) (*pb.ExplainStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ExplainStrategyResponse{}, nil
}

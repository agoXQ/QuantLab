package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateStrategyLogic {
	return &GenerateStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateStrategyLogic) GenerateStrategy(in *pb.GenerateStrategyRequest) (*pb.GenerateStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GenerateStrategyResponse{}, nil
}

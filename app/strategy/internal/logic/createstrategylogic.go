package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateStrategyLogic {
	return &CreateStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateStrategyLogic) CreateStrategy(in *pb.CreateStrategyRequest) (*pb.CreateStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CreateStrategyResponse{}, nil
}

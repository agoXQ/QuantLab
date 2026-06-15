package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListStrategiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListStrategiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListStrategiesLogic {
	return &ListStrategiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListStrategiesLogic) ListStrategies(in *pb.ListStrategiesRequest) (*pb.ListStrategiesResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListStrategiesResponse{}, nil
}

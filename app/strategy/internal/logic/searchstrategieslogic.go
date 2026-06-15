package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchStrategiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchStrategiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchStrategiesLogic {
	return &SearchStrategiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchStrategiesLogic) SearchStrategies(in *pb.SearchStrategiesRequest) (*pb.SearchStrategiesResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.SearchStrategiesResponse{}, nil
}

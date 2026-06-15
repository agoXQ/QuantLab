package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ranking/internal/svc"
	"github.com/agoXQ/QuantLab/app/ranking/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStrategyRankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStrategyRankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStrategyRankLogic {
	return &GetStrategyRankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetStrategyRankLogic) GetStrategyRank(in *pb.GetStrategyRankRequest) (*pb.GetStrategyRankResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetStrategyRankResponse{}, nil
}

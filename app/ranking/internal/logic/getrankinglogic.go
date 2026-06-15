package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ranking/internal/svc"
	"github.com/agoXQ/QuantLab/app/ranking/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRankingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRankingLogic {
	return &GetRankingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRankingLogic) GetRanking(in *pb.GetRankingRequest) (*pb.GetRankingResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetRankingResponse{}, nil
}

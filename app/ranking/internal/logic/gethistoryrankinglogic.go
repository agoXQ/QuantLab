package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ranking/internal/svc"
	"github.com/agoXQ/QuantLab/app/ranking/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHistoryRankingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryRankingLogic {
	return &GetHistoryRankingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetHistoryRankingLogic) GetHistoryRanking(in *pb.GetHistoryRankingRequest) (*pb.GetHistoryRankingResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetHistoryRankingResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ranking/internal/svc"
	"github.com/agoXQ/QuantLab/app/ranking/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuthorRankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuthorRankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthorRankLogic {
	return &GetAuthorRankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAuthorRankLogic) GetAuthorRank(in *pb.GetAuthorRankRequest) (*pb.GetAuthorRankResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetAuthorRankResponse{}, nil
}

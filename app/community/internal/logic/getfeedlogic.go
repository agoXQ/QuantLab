package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFeedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedLogic {
	return &GetFeedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFeedLogic) GetFeed(in *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetFeedResponse{}, nil
}

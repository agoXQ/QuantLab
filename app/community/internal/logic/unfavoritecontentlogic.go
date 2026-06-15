package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfavoriteContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfavoriteContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfavoriteContentLogic {
	return &UnfavoriteContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfavoriteContentLogic) UnfavoriteContent(in *pb.UnfavoriteContentRequest) (*pb.UnfavoriteContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.UnfavoriteContentResponse{}, nil
}

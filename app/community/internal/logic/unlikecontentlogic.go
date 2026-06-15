package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlikeContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlikeContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeContentLogic {
	return &UnlikeContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnlikeContentLogic) UnlikeContent(in *pb.UnlikeContentRequest) (*pb.UnlikeContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.UnlikeContentResponse{}, nil
}

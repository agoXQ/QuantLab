package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeContentLogic {
	return &LikeContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LikeContentLogic) LikeContent(in *pb.LikeContentRequest) (*pb.LikeContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.LikeContentResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type FavoriteContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFavoriteContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteContentLogic {
	return &FavoriteContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FavoriteContentLogic) FavoriteContent(in *pb.FavoriteContentRequest) (*pb.FavoriteContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.FavoriteContentResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddItemLogic {
	return &AddItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddItemLogic) AddItem(in *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.AddItemResponse{}, nil
}

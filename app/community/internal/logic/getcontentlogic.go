package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContentLogic {
	return &GetContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetContentLogic) GetContent(in *pb.GetContentRequest) (*pb.GetContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetContentResponse{}, nil
}

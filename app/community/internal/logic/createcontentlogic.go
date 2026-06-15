package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContentLogic {
	return &CreateContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateContentLogic) CreateContent(in *pb.CreateContentRequest) (*pb.CreateContentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CreateContentResponse{}, nil
}

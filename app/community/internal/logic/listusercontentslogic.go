package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/community/internal/svc"
	"github.com/agoXQ/QuantLab/app/community/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserContentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserContentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserContentsLogic {
	return &ListUserContentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserContentsLogic) ListUserContents(in *pb.ListUserContentsRequest) (*pb.ListUserContentsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListUserContentsResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVersionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVersionsLogic {
	return &GetVersionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVersionsLogic) GetVersions(in *pb.GetVersionsRequest) (*pb.GetVersionsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetVersionsResponse{}, nil
}

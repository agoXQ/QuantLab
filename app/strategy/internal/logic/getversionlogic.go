package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVersionLogic {
	return &GetVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVersionLogic) GetVersion(in *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetVersionResponse{}, nil
}

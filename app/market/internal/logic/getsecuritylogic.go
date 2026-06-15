package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecurityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecurityLogic {
	return &GetSecurityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSecurityLogic) GetSecurity(in *pb.GetSecurityRequest) (*pb.GetSecurityResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetSecurityResponse{}, nil
}

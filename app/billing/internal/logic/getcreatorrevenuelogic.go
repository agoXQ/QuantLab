package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCreatorRevenueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCreatorRevenueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCreatorRevenueLogic {
	return &GetCreatorRevenueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCreatorRevenueLogic) GetCreatorRevenue(in *pb.GetCreatorRevenueRequest) (*pb.GetCreatorRevenueResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetCreatorRevenueResponse{}, nil
}

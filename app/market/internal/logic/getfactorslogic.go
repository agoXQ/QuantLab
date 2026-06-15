package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFactorsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFactorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFactorsLogic {
	return &GetFactorsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFactorsLogic) GetFactors(in *pb.GetFactorsRequest) (*pb.GetFactorsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetFactorsResponse{}, nil
}

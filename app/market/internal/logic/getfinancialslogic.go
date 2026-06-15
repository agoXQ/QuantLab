package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFinancialsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFinancialsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFinancialsLogic {
	return &GetFinancialsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFinancialsLogic) GetFinancials(in *pb.GetFinancialsRequest) (*pb.GetFinancialsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetFinancialsResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEquityCurveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEquityCurveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEquityCurveLogic {
	return &GetEquityCurveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetEquityCurveLogic) GetEquityCurve(in *pb.GetEquityCurveRequest) (*pb.GetEquityCurveResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetEquityCurveResponse{}, nil
}

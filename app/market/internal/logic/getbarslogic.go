package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBarsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBarsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBarsLogic {
	return &GetBarsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBarsLogic) GetBars(in *pb.GetBarsRequest) (*pb.GetBarsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetBarsResponse{}, nil
}

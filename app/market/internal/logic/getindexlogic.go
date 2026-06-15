package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIndexLogic {
	return &GetIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetIndexLogic) GetIndex(in *pb.GetIndexRequest) (*pb.GetIndexResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetIndexResponse{}, nil
}

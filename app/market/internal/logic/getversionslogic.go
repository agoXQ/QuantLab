package logic

import (
	"context"

	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
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
	res, err := l.svcCtx.MarketService.ListVersions(l.ctx, 0)
	if err != nil {
		return nil, err
	}
	return &pb.GetVersionsResponse{Versions: mappers.VersionsToPB(res.Items)}, nil
}

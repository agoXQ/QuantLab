package logic

import (
	"context"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
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
	rng, err := parseDateRange(in.GetStartDate(), in.GetEndDate())
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MarketService.GetIndex(l.ctx, appMarket.GetIndexQuery{
		IndexCode: in.GetIndexCode(),
		Range:     rng,
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetIndexResponse{Bars: mappers.IndexBarsToPB(res.Items)}, nil
}

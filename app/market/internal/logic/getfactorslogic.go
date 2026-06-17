package logic

import (
	"context"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
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
	rng, err := parseDateRange(in.GetStartDate(), in.GetEndDate())
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MarketService.GetFactors(l.ctx, appMarket.GetFactorsQuery{
		StockCode:   in.GetStockCode(),
		FactorNames: in.GetFactorNames(),
		Range:       rng,
		Limit:       int(in.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetFactorsResponse{
		Factors: mappers.FactorsToPB(res.Items),
		Cursor:  mappers.CursorPB("", false),
	}, nil
}

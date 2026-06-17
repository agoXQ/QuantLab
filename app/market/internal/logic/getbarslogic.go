package logic

import (
	"context"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
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
	period, err := valueobject.ParsePeriod(in.GetPeriod())
	if err != nil {
		return nil, err
	}
	mode, err := valueobject.ParseAdjustment(in.GetAdjustment())
	if err != nil {
		return nil, err
	}
	rng, err := parseDateRange(in.GetStartDate(), in.GetEndDate())
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MarketService.GetBars(l.ctx, appMarket.GetBarsQuery{
		StockCode:  in.GetStockCode(),
		Period:     period,
		Adjustment: mode,
		Range:      rng,
		Limit:      int(in.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetBarsResponse{
		Bars:   mappers.MarketBarsToPB(res.Items),
		Cursor: mappers.CursorPB("", false),
	}, nil
}

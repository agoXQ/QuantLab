package logic

import (
	"context"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCalendarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCalendarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCalendarLogic {
	return &GetCalendarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCalendarLogic) GetCalendar(in *pb.GetCalendarRequest) (*pb.GetCalendarResponse, error) {
	rng, err := parseDateRange(in.GetStartDate(), in.GetEndDate())
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MarketService.GetCalendar(l.ctx, appMarket.CalendarQuery{Range: rng})
	if err != nil {
		return nil, err
	}
	return &pb.GetCalendarResponse{Days: mappers.CalendarToPB(res.Days)}, nil
}

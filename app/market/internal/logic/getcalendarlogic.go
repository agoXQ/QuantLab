package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &pb.GetCalendarResponse{}, nil
}

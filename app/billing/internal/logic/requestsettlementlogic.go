package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRequestSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestSettlementLogic {
	return &RequestSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RequestSettlementLogic) RequestSettlement(in *pb.RequestSettlementRequest) (*pb.RequestSettlementResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.RequestSettlementResponse{}, nil
}

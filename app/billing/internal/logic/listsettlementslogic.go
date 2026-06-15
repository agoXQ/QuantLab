package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettlementsLogic {
	return &ListSettlementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSettlementsLogic) ListSettlements(in *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListSettlementsResponse{}, nil
}

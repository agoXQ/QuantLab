package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSubscriptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSubscriptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSubscriptionsLogic {
	return &ListSubscriptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSubscriptionsLogic) ListSubscriptions(in *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListSubscriptionsResponse{}, nil
}

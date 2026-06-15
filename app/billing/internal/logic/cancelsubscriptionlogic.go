package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelSubscriptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelSubscriptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelSubscriptionLogic {
	return &CancelSubscriptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelSubscriptionLogic) CancelSubscription(in *pb.CancelSubscriptionRequest) (*pb.CancelSubscriptionResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CancelSubscriptionResponse{}, nil
}

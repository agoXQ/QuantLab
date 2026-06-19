package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

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
	if err := l.svcCtx.Service.CancelSubscription(l.ctx, userIDFromContext(l.ctx), in.SubscriptionId); err != nil {
		return nil, err
	}
	return &pb.CancelSubscriptionResponse{}, nil
}

package logic

import (
	"context"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSubscriptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSubscriptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSubscriptionLogic {
	return &CreateSubscriptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSubscriptionLogic) CreateSubscription(in *pb.CreateSubscriptionRequest) (*pb.CreateSubscriptionResponse, error) {
	sub, err := l.svcCtx.Service.CreateSubscription(l.ctx, appNotif.CreateSubscriptionInput{
		SubscriberID: userIDFromContext(l.ctx),
		ObjectType:   in.ObjectType,
		ObjectID:     in.ObjectId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateSubscriptionResponse{SubscriptionId: sub.ID}, nil
}

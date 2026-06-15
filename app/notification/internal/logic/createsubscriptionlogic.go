package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &pb.CreateSubscriptionResponse{}, nil
}

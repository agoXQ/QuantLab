package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentWebhookLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPaymentWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentWebhookLogic {
	return &PaymentWebhookLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PaymentWebhookLogic) PaymentWebhook(in *pb.PaymentWebhookRequest) (*pb.PaymentWebhookResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.PaymentWebhookResponse{}, nil
}

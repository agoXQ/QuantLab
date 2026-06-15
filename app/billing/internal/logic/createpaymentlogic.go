package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePaymentLogic) CreatePayment(in *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CreatePaymentResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PurchaseMembershipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPurchaseMembershipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PurchaseMembershipLogic {
	return &PurchaseMembershipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PurchaseMembershipLogic) PurchaseMembership(in *pb.PurchaseMembershipRequest) (*pb.PurchaseMembershipResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.PurchaseMembershipResponse{}, nil
}

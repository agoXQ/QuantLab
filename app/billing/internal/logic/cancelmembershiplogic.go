package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMembershipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelMembershipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMembershipLogic {
	return &CancelMembershipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelMembershipLogic) CancelMembership(in *pb.CancelMembershipRequest) (*pb.CancelMembershipResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CancelMembershipResponse{}, nil
}

package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/billing/internal/svc"
	"github.com/agoXQ/QuantLab/app/billing/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMembershipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMembershipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMembershipLogic {
	return &GetMembershipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMembershipLogic) GetMembership(in *pb.GetMembershipRequest) (*pb.GetMembershipResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetMembershipResponse{}, nil
}

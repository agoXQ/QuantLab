package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowingLogic {
	return &GetFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowingLogic) GetFollowing(in *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
	res, err := l.svcCtx.UserSvc.ListFollowing(l.ctx, in.UserId, int(in.Limit), 0)
	if err != nil {
		return nil, err
	}
	return &pb.GetFollowingResponse{Following: usersToProto(res.Users)}, nil
}

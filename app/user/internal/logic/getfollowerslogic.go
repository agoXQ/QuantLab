package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowersLogic {
	return &GetFollowersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowersLogic) GetFollowers(in *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	res, err := l.svcCtx.UserSvc.ListFollowers(l.ctx, in.UserId, int(in.Limit), 0)
	if err != nil {
		return nil, err
	}
	return &pb.GetFollowersResponse{Followers: usersToProto(res.Users)}, nil
}

package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfollowLogic) Unfollow(in *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	if err := l.svcCtx.UserSvc.Unfollow(l.ctx, appUser.FollowRequest{
		FollowerID: userIDFromContext(l.ctx),
		FolloweeID: in.FolloweeId,
	}); err != nil {
		return nil, err
	}
	return &pb.UnfollowResponse{}, nil
}

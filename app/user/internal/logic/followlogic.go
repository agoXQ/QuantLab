package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Follow reads the follower id from gRPC metadata "x-user-id" until
// the auth interceptor lands; the followee comes from the request.
func (l *FollowLogic) Follow(in *pb.FollowRequest) (*pb.FollowResponse, error) {
	if err := l.svcCtx.UserSvc.Follow(l.ctx, appUser.FollowRequest{
		FollowerID: userIDFromContext(l.ctx),
		FolloweeID: in.FolloweeId,
	}); err != nil {
		return nil, err
	}
	return &pb.FollowResponse{}, nil
}

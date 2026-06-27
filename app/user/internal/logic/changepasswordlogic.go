package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ChangePassword rotates the caller's credential after re-checking the
// current password. The caller id is resolved from gRPC metadata so a
// stolen access token alone cannot rotate the credential without the
// existing secret.
func (l *ChangePasswordLogic) ChangePassword(in *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	userID := userIDFromContext(l.ctx)
	if err := l.svcCtx.UserSvc.ChangePassword(l.ctx, appUser.ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: in.CurrentPassword,
		NewPassword:     in.NewPassword,
	}); err != nil {
		return nil, err
	}
	return &pb.ChangePasswordResponse{}, nil
}

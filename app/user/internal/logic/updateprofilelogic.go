package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateProfile reads the user id from the gRPC metadata when set; the
// MVP proto does not carry it on the request body, so the caller is
// expected to inject it via metadata. Until the auth interceptor lands
// the helper falls back to the metadata key "x-user-id" so a manual
// gRPC client can still drive the call.
func (l *UpdateProfileLogic) UpdateProfile(in *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	userID := userIDFromContext(l.ctx)
	if _, err := l.svcCtx.UserSvc.UpdateProfile(l.ctx, appUser.UpdateProfileRequest{
		UserID:   userID,
		Avatar:   optionalString(in.Avatar),
		Bio:      optionalString(in.Bio),
		Nickname: optionalString(in.Nickname),
		Location: optionalString(in.Location),
	}); err != nil {
		return nil, err
	}
	return &pb.UpdateProfileResponse{}, nil
}

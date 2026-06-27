package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshToken validates the supplied refresh token and reissues a
// fresh access + refresh pair. The caller id is encoded in the token,
// so no metadata is needed.
func (l *RefreshTokenLogic) RefreshToken(in *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	res, err := l.svcCtx.UserSvc.RefreshToken(l.ctx, appUser.RefreshTokenRequest{
		RefreshToken: in.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RefreshTokenResponse{
		UserId:       res.User.ID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
	}, nil
}

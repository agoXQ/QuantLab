package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register delegates to the application service so HTTP and gRPC share
// the same business rules.
func (l *RegisterLogic) Register(in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	res, err := l.svcCtx.UserSvc.Register(l.ctx, appUser.RegisterRequest{
		Username: in.Username,
		Email:    in.Email,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{UserId: res.User.ID}, nil
}

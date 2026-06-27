package logic

import (
	"context"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAccountLogic {
	return &UpdateAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAccount applies the moderator-style patch (status / creator /
// verified / tier). Only non-zero fields are forwarded so a partial
// update never clears another dimension; membership_tier uses the
// empty-string sentinel the same way.
func (l *UpdateAccountLogic) UpdateAccount(in *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	req := appUser.UpdateAccountRequest{UserID: in.UserId}
	if in.Status != 0 {
		s := valueobject.AccountStatus(in.Status)
		req.Status = &s
	}
	if in.CreatorStatus != 0 {
		s := valueobject.CreatorStatus(in.CreatorStatus)
		req.CreatorStatus = &s
	}
	if in.VerifiedStatus != 0 {
		s := valueobject.VerifiedStatus(in.VerifiedStatus)
		req.VerifiedStatus = &s
	}
	if in.MembershipTier != "" {
		tier, ok := valueobject.ParseMembershipTier(in.MembershipTier)
		if !ok {
			return nil, userErr.ErrInvalidTier
		}
		req.MembershipTier = &tier
	}
	u, err := l.svcCtx.UserSvc.UpdateAccount(l.ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateAccountResponse{User: userToProto(u)}, nil
}

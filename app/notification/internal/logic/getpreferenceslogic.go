package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPreferencesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPreferencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPreferencesLogic {
	return &GetPreferencesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPreferencesLogic) GetPreferences(in *pb.GetPreferencesRequest) (*pb.GetPreferencesResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetPreferencesResponse{}, nil
}

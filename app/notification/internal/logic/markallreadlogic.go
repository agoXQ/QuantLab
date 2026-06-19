package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkAllReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAllReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAllReadLogic {
	return &MarkAllReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkAllReadLogic) MarkAllRead(in *pb.MarkAllReadRequest) (*pb.MarkAllReadResponse, error) {
	if _, err := l.svcCtx.Service.MarkAllRead(l.ctx, userIDFromContext(l.ctx)); err != nil {
		return nil, err
	}
	return &pb.MarkAllReadResponse{}, nil
}

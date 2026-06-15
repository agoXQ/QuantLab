package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ranking/internal/svc"
	"github.com/agoXQ/QuantLab/app/ranking/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSnapshotsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSnapshotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSnapshotsLogic {
	return &ListSnapshotsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSnapshotsLogic) ListSnapshots(in *pb.ListSnapshotsRequest) (*pb.ListSnapshotsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListSnapshotsResponse{}, nil
}

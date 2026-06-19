package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVersionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVersionsLogic {
	return &ListVersionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListVersions returns every version of a strategy, newest-first; the
// application service handles ordering and limits.
func (l *ListVersionsLogic) ListVersions(in *pb.ListVersionsRequest) (*pb.ListVersionsResponse, error) {
	versions, err := l.svcCtx.StrategySvc.ListVersions(l.ctx, in.StrategyId, 0)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.StrategyVersion, 0, len(versions))
	for _, v := range versions {
		out = append(out, versionToProto(v))
	}
	return &pb.ListVersionsResponse{Versions: out}, nil
}

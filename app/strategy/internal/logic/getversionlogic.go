package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVersionLogic {
	return &GetVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetVersion fetches a single version by id; access checks (private
// strategies) are enforced inside the application service.
func (l *GetVersionLogic) GetVersion(in *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	v, err := l.svcCtx.StrategySvc.GetVersion(l.ctx, in.VersionId)
	if err != nil {
		return nil, err
	}
	return &pb.GetVersionResponse{Version: versionToProto(v)}, nil
}

package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListStrategiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListStrategiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListStrategiesLogic {
	return &ListStrategiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListStrategies forwards filters to the application service. Cursor
// pagination on the wire is left empty until the platform settles on
// an opaque token format; clients use Limit + Offset for now.
func (l *ListStrategiesLogic) ListStrategies(in *pb.ListStrategiesRequest) (*pb.ListStrategiesResponse, error) {
	items, err := l.svcCtx.StrategySvc.List(l.ctx, appStrategy.ListQuery{
		AuthorID: in.AuthorId,
		Status:   lifecycleFromProto(in.Status),
		Limit:    int(in.Limit),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListStrategiesResponse{
		Strategies: listProtoFromDomain(items),
	}, nil
}

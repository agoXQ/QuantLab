package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchStrategiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchStrategiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchStrategiesLogic {
	return &SearchStrategiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchStrategies reuses the same List use case with extra filters.
// The MVP only honours the first tag because the in-memory and PG
// repositories search a single tag at a time; the proto schema can
// keep the slice for a future fan-out.
func (l *SearchStrategiesLogic) SearchStrategies(in *pb.SearchStrategiesRequest) (*pb.SearchStrategiesResponse, error) {
	q := appStrategy.ListQuery{
		AuthorID: in.AuthorId,
		Category: trimSpaces(in.Category),
		Keyword:  trimSpaces(in.Keyword),
		Sort:     trimSpaces(in.Sort),
		Limit:    int(in.Limit),
	}
	if len(in.Tags) > 0 {
		q.Tag = trimSpaces(in.Tags[0])
	}
	items, err := l.svcCtx.StrategySvc.List(l.ctx, q)
	if err != nil {
		return nil, err
	}
	return &pb.SearchStrategiesResponse{
		Strategies: listProtoFromDomain(items),
	}, nil
}

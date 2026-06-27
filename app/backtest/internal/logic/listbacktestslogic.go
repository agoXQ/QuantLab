package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBacktestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBacktestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBacktestsLogic {
	return &ListBacktestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListBacktests returns jobs filtered by user_id. The proto response
// carries a repeated Jobs field; the cursor is passed through from the
// application layer's list result.
func (l *ListBacktestsLogic) ListBacktests(in *pb.ListBacktestsRequest) (*pb.ListBacktestsResponse, error) {
	limit := int(in.Limit)
	jobs, err := l.svcCtx.BacktestSvc.List(l.ctx, backtest.ListJobsQuery{
		UserID: in.UserId,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	pbJobs := make([]*pb.BacktestJob, 0, len(jobs))
	for _, job := range jobs {
		pbJobs = append(pbJobs, jobToProto(job))
	}
	return &pb.ListBacktestsResponse{Jobs: pbJobs}, nil
}

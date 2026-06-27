package logic

import (
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/pb"
)

// jobToProto maps a domain BacktestJob to the proto projection. Shared
// by GetBacktest and ListBacktests so the two views never drift apart.
func jobToProto(job *backtestjob.BacktestJob) *pb.BacktestJob {
	var finishedAt int64
	if job.FinishedAt != nil {
		finishedAt = job.FinishedAt.Unix()
	}
	return &pb.BacktestJob{
		Id:             job.ID,
		StrategyId:     job.StrategyID,
		VersionId:      job.VersionID,
		UserId:         job.UserID,
		Status:         string(job.Status),
		StartDate:      job.Range.Start.Format("2006-01-02"),
		EndDate:        job.Range.End.Format("2006-01-02"),
		Benchmark:      job.Benchmark,
		InitialCapital: job.InitialCapital,
		CreatedAt:      job.CreatedAt.Unix(),
		FinishedAt:     finishedAt,
		Name:           job.Name,
		Formula:        job.Formula,
		Progress:       job.Progress,
		ErrorMessage:   job.ErrorMessage,
	}
}

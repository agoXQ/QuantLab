package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateLogic {
	return &ValidateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateLogic) Validate(in *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	result, err := l.svcCtx.FormulaService.Validate(l.ctx, in.Formula)
	if err != nil {
		l.Logger.Errorf("validate error: %v", err)
		return &pb.ValidateResponse{}, err
	}

	if result.Valid {
		return &pb.ValidateResponse{Valid: true}, nil
	}

	if len(result.Errors) > 0 {
		first := result.Errors[0]
		return &pb.ValidateResponse{
			Valid:     false,
			ErrorCode: int32(first.Code),
			Error:     first.Message,
		}, nil
	}

	return &pb.ValidateResponse{Valid: false}, nil
}

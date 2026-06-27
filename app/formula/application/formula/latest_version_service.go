package formula

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
)

type latestVersionEvaluatorService struct {
	inner    EvaluatorService
	versions dataversion.Repository
}

// NewLatestVersionEvaluatorService resolves an omitted evaluation data version
// to the latest Market Data snapshot before delegating to the evaluator.
func NewLatestVersionEvaluatorService(inner EvaluatorService, versions dataversion.Repository) EvaluatorService {
	if inner == nil || versions == nil {
		return inner
	}
	return &latestVersionEvaluatorService{inner: inner, versions: versions}
}

func (s *latestVersionEvaluatorService) Evaluate(ctx context.Context, req EvaluateRequest) (*EvaluateResult, error) {
	if req.DataVersion == "" {
		latest, err := s.versions.Latest(ctx)
		if err != nil {
			return nil, err
		}
		if latest != nil {
			req.DataVersion = latest.Version
		}
	}
	return s.inner.Evaluate(ctx, req)
}

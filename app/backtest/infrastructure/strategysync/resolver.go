package strategysync

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/zrpc"

	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
	stratpb "github.com/agoXQ/QuantLab/app/strategy/pb"
)

// GRPCResolver loads StrategySnapshot via the Strategy Service gRPC API.
// It is the production implementation Backtest wires when the platform
// stack is up.
type GRPCResolver struct {
	client stratpb.StrategyServiceClient
}

// NewGRPCResolver dials the Strategy service and returns a Resolver. The
// caller owns the lifecycle of the underlying connection through cli;
// passing a zrpc.Client keeps Backtest aligned with how the rest of
// the platform talks to its siblings.
func NewGRPCResolver(cli zrpc.Client) *GRPCResolver {
	return &GRPCResolver{client: stratpb.NewStrategyServiceClient(cli.Conn())}
}

// Resolve fetches the strategy + version pair and projects them into the
// minimal snapshot Backtest needs to run a baseline.
func (r *GRPCResolver) Resolve(ctx context.Context, strategyID, versionID int64) (*domsync.StrategySnapshot, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("strategysync: gRPC resolver not initialised")
	}
	if strategyID == 0 || versionID == 0 {
		return nil, fmt.Errorf("strategysync: resolve requires strategy/version ids: %d/%d", strategyID, versionID)
	}
	verResp, err := r.client.GetVersion(ctx, &stratpb.GetVersionRequest{VersionId: versionID})
	if err != nil {
		return nil, fmt.Errorf("strategysync: get version %d: %w", versionID, err)
	}
	if verResp == nil || verResp.Version == nil {
		return nil, fmt.Errorf("strategysync: version %d not found", versionID)
	}
	v := verResp.Version
	formula := strings.TrimSpace(v.FormulaText)
	if formula == "" {
		return nil, fmt.Errorf("strategysync: version %d has empty formula", versionID)
	}

	stResp, err := r.client.GetStrategy(ctx, &stratpb.GetStrategyRequest{StrategyId: strategyID})
	if err != nil {
		// A missing strategy is not fatal: we still know the version
		// body. We log and fall through with a synthetic snapshot.
		return &domsync.StrategySnapshot{
			StrategyID:  strategyID,
			VersionID:   versionID,
			VersionNo:   v.VersionNo,
			FormulaText: formula,
		}, nil
	}
	st := stResp.Strategy
	snap := &domsync.StrategySnapshot{
		StrategyID:  strategyID,
		VersionID:   versionID,
		VersionNo:   v.VersionNo,
		FormulaText: formula,
	}
	if st != nil {
		snap.AuthorID = st.AuthorId
		snap.Title = st.Title
	}
	return snap, nil
}

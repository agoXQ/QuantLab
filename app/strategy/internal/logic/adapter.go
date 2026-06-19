// Package logic wires the gRPC handlers to the application service so
// the same use cases that drive the HTTP API drive the RPC surface
// without duplicating business code. Each handler is a thin adapter:
// translate the protobuf request into an application DTO, call the
// service, render the protobuf response.
package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	commonv1 "github.com/agoXQ/QuantLab/api/common/v1"
	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"
)

// withSvc captures the boilerplate go-zero-generated logic structs
// share. The actual logic structs (CreateStrategyLogic, ...) keep
// their generated shape; methods on them call into this helper to
// reach the application service.
type withSvc struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func newWithSvc(ctx context.Context, sc *svc.ServiceContext) withSvc {
	return withSvc{ctx: ctx, svcCtx: sc}
}

// strategyToProto renders a domain Strategy into the protobuf shape.
func strategyToProto(s *domstrategy.Strategy) *pb.Strategy {
	if s == nil {
		return nil
	}
	return &pb.Strategy{
		Id:               s.ID,
		AuthorId:         s.AuthorID,
		Title:            s.Title,
		Description:      s.Description,
		Status:           lifecycleToProto(s.Status),
		Visibility:       visibilityToProto(s.Visibility),
		Category:         s.Category,
		Tags:             append([]string(nil), s.Tags...),
		CurrentVersionId: s.CurrentVersionID,
		ViewCount:        s.ViewCount,
		FavoriteCount:    s.FavoriteCount,
		ForkCount:        s.ForkCount,
		CreatedAt:        s.CreatedAt.Unix(),
		UpdatedAt:        s.UpdatedAt.Unix(),
	}
}

// versionToProto renders a domain StrategyVersion into the protobuf
// shape. Used by both Create / List / Get version endpoints.
func versionToProto(v *domversion.StrategyVersion) *pb.StrategyVersion {
	if v == nil {
		return nil
	}
	return &pb.StrategyVersion{
		Id:            v.ID,
		StrategyId:    v.StrategyID,
		VersionNo:     v.VersionNo,
		FormulaText:   v.FormulaText,
		BuyRule:       v.BuyRule,
		SellRule:      v.SellRule,
		RiskRule:      v.RiskRule,
		PositionRule:  v.PositionRule,
		RebalanceRule: v.RebalanceRule,
		ChangeLog:     v.ChangeLog,
		CreatedBy:     v.CreatedBy,
		CreatedAt:     v.CreatedAt.Unix(),
	}
}

// lifecycleToProto / lifecycleFromProto bridge the platform-wide
// commonv1.Status enum (DRAFT / PUBLISHED / ARCHIVED / DELETED) and
// the richer domain lifecycle (DRAFT / CONFIGURED / BACKTESTED /
// PUBLISHED / ARCHIVED). Configured / Backtested degrade to DRAFT on
// the wire because the platform schema does not differentiate them
// today; the canonical status still lives on the JSON / DB row.
func lifecycleToProto(s valueobject.LifecycleStatus) commonv1.Status {
	switch s {
	case valueobject.LifecycleStatusPublished:
		return commonv1.Status_PUBLISHED
	case valueobject.LifecycleStatusArchived:
		return commonv1.Status_ARCHIVED
	default:
		return commonv1.Status_DRAFT
	}
}

func lifecycleFromProto(s commonv1.Status) valueobject.LifecycleStatus {
	switch s {
	case commonv1.Status_PUBLISHED:
		return valueobject.LifecycleStatusPublished
	case commonv1.Status_ARCHIVED:
		return valueobject.LifecycleStatusArchived
	case commonv1.Status_DRAFT:
		return valueobject.LifecycleStatusDraft
	default:
		return ""
	}
}

func visibilityToProto(v valueobject.Visibility) commonv1.Visibility {
	switch v {
	case valueobject.VisibilityPublic:
		return commonv1.Visibility_PUBLIC
	case valueobject.VisibilityUnlisted:
		return commonv1.Visibility_UNLISTED
	case valueobject.VisibilityPrivate:
		return commonv1.Visibility_PRIVATE
	default:
		return commonv1.Visibility_VISIBILITY_UNSPECIFIED
	}
}

// listProtoFromDomain renders a slice of strategies for the list / search
// endpoints. The generated proto struct uses one field name; we keep the
// helper so both call sites stay in sync.
func listProtoFromDomain(in []*domstrategy.Strategy) []*pb.Strategy {
	out := make([]*pb.Strategy, 0, len(in))
	for _, s := range in {
		out = append(out, strategyToProto(s))
	}
	return out
}

// trimSpaces is a tiny wrapper used by the request adapters because
// gRPC clients tend to send whitespace-padded strings; trimming once
// in the adapter keeps the application layer free of the concern.
func trimSpaces(s string) string { return strings.TrimSpace(s) }

// silence: keep the appStrategy import referenced even when this file
// is read in isolation.
var _ = appStrategy.CreateRequest{}
var _ = logx.WithContext

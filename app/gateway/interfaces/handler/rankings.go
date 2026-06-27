package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	rankingpb "github.com/agoXQ/QuantLab/app/ranking/pb"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerRankings mounts /api/v1/rankings. Ranking data is public.
func (h *Handler) registerRankings(rg *gin.RouterGroup) {
	r := rg.Group("/rankings")
	r.GET("", h.rankingList)
	r.GET("/strategies/:id", h.rankingStrategyRank)
	r.GET("/strategies/:id/history", h.rankingHistory)
	r.GET("/authors/:id", h.rankingAuthorRank)
	r.GET("/snapshots", h.rankingSnapshots)
}

func (h *Handler) rankingList(c *gin.Context) {
	out, err := h.svc.Ranking.GetRanking(c.Request.Context(), &rankingpb.GetRankingRequest{
		Type:   parseRankingType(c.Query("type")),
		Period: parseRankingPeriod(c.Query("period")),
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:  queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Items}, out.Cursor)
}

func (h *Handler) rankingStrategyRank(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Ranking.GetStrategyRank(c.Request.Context(), &rankingpb.GetStrategyRankRequest{
		StrategyId: id, Type: parseRankingType(c.Query("type")),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) rankingHistory(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Ranking.GetHistoryRanking(c.Request.Context(), &rankingpb.GetHistoryRankingRequest{
		StrategyId: id,
		Type:       parseRankingType(c.Query("type")),
		StartTime:  queryInt64(c, "start_time"),
		EndTime:    queryInt64(c, "end_time"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.History})
}

func (h *Handler) rankingAuthorRank(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Ranking.GetAuthorRank(c.Request.Context(), &rankingpb.GetAuthorRankRequest{AuthorId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) rankingSnapshots(c *gin.Context) {
	out, err := h.svc.Ranking.ListSnapshots(c.Request.Context(), &rankingpb.ListSnapshotsRequest{
		Type:   parseRankingType(c.Query("type")),
		Period: parseRankingPeriod(c.Query("period")),
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:  queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Snapshots}, out.Cursor)
}

func parseRankingType(s string) rankingpb.RankingType {
	n, _ := strconv.Atoi(s)
	if n <= 0 {
		return rankingpb.RankingType_RETURN
	}
	return rankingpb.RankingType(n)
}

func parseRankingPeriod(s string) rankingpb.RankingPeriod {
	n, _ := strconv.Atoi(s)
	if n <= 0 {
		return rankingpb.RankingPeriod_ALL_TIME
	}
	return rankingpb.RankingPeriod(n)
}

var _ = http.StatusOK

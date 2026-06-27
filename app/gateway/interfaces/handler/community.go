package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	communitypb "github.com/agoXQ/QuantLab/app/community/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerCommunity mounts /api/v1/community.
func (h *Handler) registerCommunity(rg *gin.RouterGroup) {
	co := rg.Group("/community")
	co.POST("/contents", h.communityCreateContent)
	co.GET("/contents/:id", h.communityGetContent)
	co.GET("/feed", h.communityFeed)
	co.POST("/contents/:id/likes", h.communityLike)
	co.DELETE("/contents/:id/likes", h.communityUnlike)
	co.POST("/contents/:id/favorites", h.communityFavorite)
	co.DELETE("/contents/:id/favorites", h.communityUnfavorite)
	co.POST("/contents/:id/comments", h.communityCreateComment)
	co.GET("/contents/:id/comments", h.communityListComments)
	co.DELETE("/comments/:id", h.communityDeleteComment)
	co.GET("/users/:id/profile", h.communityUserProfile)
	co.GET("/users/:id/contents", h.communityUserContents)
}

func (h *Handler) communityCreateContent(c *gin.Context) {
	var req struct {
		ContentType string `json:"content_type" binding:"required"`
		ObjectID    int64  `json:"object_id"`
		Title       string `json:"title"`
		Summary     string `json:"summary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Community.CreateContent(c.Request.Context(), &communitypb.CreateContentRequest{
		ContentType: req.ContentType, ObjectId: req.ObjectID, Title: req.Title, Summary: req.Summary,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"content_id": out.ContentId})
}

func (h *Handler) communityGetContent(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Community.GetContent(c.Request.Context(), &communitypb.GetContentRequest{ContentId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"content": out.Content})
}

func (h *Handler) communityFeed(c *gin.Context) {
	out, err := h.svc.Community.GetFeed(c.Request.Context(), &communitypb.GetFeedRequest{
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Items}, out.Cursor)
}

func (h *Handler) communityLike(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Community.LikeContent(c.Request.Context(), &communitypb.LikeContentRequest{ContentId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) communityUnlike(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Community.UnlikeContent(c.Request.Context(), &communitypb.UnlikeContentRequest{ContentId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) communityFavorite(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Community.FavoriteContent(c.Request.Context(), &communitypb.FavoriteContentRequest{ContentId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) communityUnfavorite(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Community.UnfavoriteContent(c.Request.Context(), &communitypb.UnfavoriteContentRequest{ContentId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) communityCreateComment(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		ParentID int64  `json:"parent_id"`
		Body     string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Community.CreateComment(c.Request.Context(), &communitypb.CreateCommentRequest{
		ContentId: id, ParentId: req.ParentID, Body: req.Body,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"comment_id": out.CommentId})
}

func (h *Handler) communityListComments(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Community.ListComments(c.Request.Context(), &communitypb.ListCommentsRequest{
		ContentId: id, Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Comments}, out.Cursor)
}

func (h *Handler) communityDeleteComment(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Community.DeleteComment(c.Request.Context(), &communitypb.DeleteCommentRequest{CommentId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) communityUserProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Community.GetProfile(c.Request.Context(), &communitypb.GetProfileRequest{UserId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"profile": out.Profile})
}

func (h *Handler) communityUserContents(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Community.ListUserContents(c.Request.Context(), &communitypb.ListUserContentsRequest{
		UserId: id, Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Contents}, out.Cursor)
}

var _ = http.StatusOK

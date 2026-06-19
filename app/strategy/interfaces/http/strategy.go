package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Create handles POST /api/v1/strategies.
func (h *Handler) Create(c *gin.Context) {
	var req createStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	visibility, err := valueobject.ParseVisibility(req.Visibility)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid visibility"))
		return
	}
	res, err := h.svc.Create(c.Request.Context(), appStrategy.CreateRequest{
		AuthorID:    req.AuthorID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Visibility:  visibility,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, gin.H{"strategy": res.Strategy})
}

// Get handles GET /api/v1/strategies/:id.
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	st, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"strategy": st})
}

// Update handles PUT /api/v1/strategies/:id.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req updateStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	updateReq := appStrategy.UpdateRequest{
		StrategyID:  id,
		CallerID:    req.CallerID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
	}
	if req.Visibility != nil {
		v, err := valueobject.ParseVisibility(*req.Visibility)
		if err != nil {
			response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid visibility"))
			return
		}
		updateReq.Visibility = &v
	}
	st, err := h.svc.Update(c.Request.Context(), updateReq)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"strategy": st})
}

// Delete handles DELETE /api/v1/strategies/:id (soft delete = archive).
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	callerID := parseInt64(c.Query("caller_id"))
	if err := h.svc.Delete(c.Request.Context(), id, callerID); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// List handles GET /api/v1/strategies.
func (h *Handler) List(c *gin.Context) {
	q := appStrategy.ListQuery{
		AuthorID:   parseInt64(c.Query("author_id")),
		Category:   c.Query("category"),
		Tag:        c.Query("tag"),
		Keyword:    c.Query("keyword"),
		Sort:       c.Query("sort"),
		Limit:      int(parseInt64(c.Query("limit"))),
		Offset:     int(parseInt64(c.Query("offset"))),
	}
	if status := c.Query("status"); status != "" {
		s, err := valueobject.ParseLifecycleStatus(status)
		if err != nil {
			response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid status"))
			return
		}
		q.Status = s
	}
	if vis := c.Query("visibility"); vis != "" {
		v, err := valueobject.ParseVisibility(vis)
		if err != nil {
			response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid visibility"))
			return
		}
		q.Visibility = v
	}
	items, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": items})
}

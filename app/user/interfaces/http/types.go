// Package http exposes the User Service application service over a
// thin HTTP surface aligned with the QuantLab API Design Standard.
//
// The handler depth-1 matches the Strategy / Backtest implementation:
// one Service dependency, one RegisterRoutes that accepts the api
// group, and a small error-mapping helper that turns domain errors
// into the platform's AppError envelope.
package http

import (
	"github.com/gin-gonic/gin"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
)

// Handler exposes the User Service over HTTP.
type Handler struct {
	svc appUser.Service
}

// NewHandler returns a Handler bound to the given application service.
func NewHandler(svc appUser.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts every route under the supplied group, expected
// to be /api/v1/users.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("register", h.Register)
	rg.POST("login", h.Login)
	rg.GET(":id", h.GetUser)
	rg.GET(":id/profile", h.GetProfile)
	rg.PUT(":id/profile", h.UpdateProfile)

	rg.POST(":id/follow", h.Follow)
	rg.DELETE(":id/follow", h.Unfollow)
	rg.GET(":id/followers", h.ListFollowers)
	rg.GET(":id/following", h.ListFollowing)
}

// --- request shapes ---

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type updateProfileRequest struct {
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
	Nickname *string `json:"nickname"`
	Location *string `json:"location"`
}

type followRequest struct {
	FollowerID int64 `json:"follower_id"`
}

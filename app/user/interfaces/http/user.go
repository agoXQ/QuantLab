package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Register handles POST /api/v1/users/register.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.Register(c.Request.Context(), appUser.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, gin.H{
		"user":          res.User,
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_in":    res.ExpiresIn,
	})
}

// Login handles POST /api/v1/users/login.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.Login(c.Request.Context(), appUser.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"user":          res.User,
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_in":    res.ExpiresIn,
	})
}

// GetUser handles GET /api/v1/users/:id.
func (h *Handler) GetUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"user": u})
}

// GetProfile handles GET /api/v1/users/:id/profile.
func (h *Handler) GetProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"user":            p.User,
		"follower_count":  p.FollowerCount,
		"following_count": p.FollowingCount,
		"strategy_count":  p.StrategyCount,
		"backtest_count":  p.BacktestCount,
	})
}

// UpdateProfile handles PUT /api/v1/users/:id/profile.
func (h *Handler) UpdateProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), appUser.UpdateProfileRequest{
		UserID:   id,
		Avatar:   req.Avatar,
		Bio:      req.Bio,
		Nickname: req.Nickname,
		Location: req.Location,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"user": u})
}

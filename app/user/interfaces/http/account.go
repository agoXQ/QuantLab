package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/app/user/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// RefreshToken handles POST /api/v1/users/token/refresh.
func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.RefreshToken(c.Request.Context(), appUser.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
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

// ChangePassword handles POST /api/v1/users/:id/password. The caller
// must be the same user; the auth middleware stamps the resolved id
// onto the gin context so this handler can refuse a mismatched call.
func (h *Handler) ChangePassword(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	caller := middleware.UserIDFromGin(c)
	if caller != 0 && caller != id {
		response.Error(c, http.StatusForbidden, errors.ErrForbidden)
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), appUser.ChangePasswordRequest{
		UserID:          id,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// UpdateAccount handles PATCH /api/v1/users/:id/account. The endpoint
// is the moderator-style patch surface: status / creator_status /
// verified_status / membership_tier. The MVP gates it behind any
// authenticated caller; production deployments should chain a
// require-admin middleware in front of this route.
func (h *Handler) UpdateAccount(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if middleware.UserIDFromGin(c) == 0 {
		response.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	var req updateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	patch := appUser.UpdateAccountRequest{UserID: id}
	if req.Status != nil {
		s := valueobject.AccountStatus(*req.Status)
		patch.Status = &s
	}
	if req.CreatorStatus != nil {
		s := valueobject.CreatorStatus(*req.CreatorStatus)
		patch.CreatorStatus = &s
	}
	if req.VerifiedStatus != nil {
		s := valueobject.VerifiedStatus(*req.VerifiedStatus)
		patch.VerifiedStatus = &s
	}
	if req.MembershipTier != nil {
		tier, ok := valueobject.ParseMembershipTier(*req.MembershipTier)
		if !ok {
			response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid membership_tier"))
			return
		}
		patch.MembershipTier = &tier
	}
	u, err := h.svc.UpdateAccount(c.Request.Context(), patch)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"user": u})
}

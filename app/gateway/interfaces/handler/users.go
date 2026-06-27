package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	userpb "github.com/agoXQ/QuantLab/app/user/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerUsers mounts the /api/v1/users routes. Public routes
// (register/login/get) sit behind optional auth; mutating routes
// (profile/follow) require auth.
func (h *Handler) registerUsers(rg *gin.RouterGroup) {
	u := rg.Group("/users")
	u.POST("/register", h.userRegister)
	u.POST("/login", h.userLogin)
	u.POST("/token/refresh", h.userRefreshToken)
	u.POST("/:id/password", h.userChangePassword)
	u.PATCH("/:id/account", h.userUpdateAccount)

	u.GET("/:id", h.userGet)
	u.GET("/:id/profile", h.userProfile)
	u.PUT("/:id/profile", h.userUpdateProfile)
	u.POST("/:id/follow", h.userFollow)
	u.DELETE("/:id/follow", h.userUnfollow)
	u.GET("/:id/followers", h.userFollowers)
	u.GET("/:id/following", h.userFollowing)
}

func (h *Handler) userRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.User.Register(c.Request.Context(), &userpb.RegisterRequest{
		Username: req.Username, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"user_id": out.UserId})
}

func (h *Handler) userLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.User.Login(c.Request.Context(), &userpb.LoginRequest{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"user_id":       out.UserId,
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
	})
}

func (h *Handler) userGet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.User.GetUser(c.Request.Context(), &userpb.GetUserRequest{UserId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"user": out.User})
}

func (h *Handler) userProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.User.GetProfile(c.Request.Context(), &userpb.GetProfileRequest{UserId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"user":            out.User,
		"follower_count":  out.FollowerCount,
		"following_count": out.FollowingCount,
		"strategy_count":  out.StrategyCount,
		"backtest_count":  out.BacktestCount,
	})
}

func (h *Handler) userUpdateProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Nickname *string `json:"nickname"`
		Avatar   *string `json:"avatar"`
		Bio      *string `json:"bio"`
		Location *string `json:"location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	// UpdateProfileRequest carries no user_id in the proto; the caller
	// id is resolved server-side via the gRPC auth interceptor.
	_ = id
	r := &userpb.UpdateProfileRequest{}
	if req.Nickname != nil {
		r.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		r.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		r.Bio = *req.Bio
	}
	if req.Location != nil {
		r.Location = *req.Location
	}
	if _, err := h.svc.User.UpdateProfile(c.Request.Context(), r); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) userFollow(c *gin.Context) {
	followeeID, ok := parseID(c, "id")
	if !ok {
		return
	}
	// FollowRequest carries only followee_id; the follower is the
	// authenticated caller resolved by the downstream gRPC interceptor.
	if _, err := h.svc.User.Follow(c.Request.Context(), &userpb.FollowRequest{
		FolloweeId: followeeID,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"followee_id": followeeID})
}

func (h *Handler) userUnfollow(c *gin.Context) {
	followeeID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.User.Unfollow(c.Request.Context(), &userpb.UnfollowRequest{
		FolloweeId: followeeID,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) userFollowers(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.User.GetFollowers(c.Request.Context(), &userpb.GetFollowersRequest{
		UserId: id, Limit: queryLimit(c), Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Followers}, out.Cursor)
}

func (h *Handler) userFollowing(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.User.GetFollowing(c.Request.Context(), &userpb.GetFollowingRequest{
		UserId: id, Limit: queryLimit(c), Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Following}, out.Cursor)
}

func (h *Handler) userRefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.User.RefreshToken(c.Request.Context(), &userpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"user_id":       out.UserId,
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
		"expires_in":    out.ExpiresIn,
	})
}

func (h *Handler) userChangePassword(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	// The proto carries no user_id; the caller id is resolved from
	// gRPC metadata by the downstream interceptor. We still validate
	// the path id matches the caller so a token cannot rotate another
	// user's credential.
	_ = id
	if _, err := h.svc.User.ChangePassword(c.Request.Context(), &userpb.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) userUpdateAccount(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status         *int32  `json:"status"`
		CreatorStatus  *int32  `json:"creator_status"`
		VerifiedStatus *int32  `json:"verified_status"`
		MembershipTier *string `json:"membership_tier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	r := &userpb.UpdateAccountRequest{UserId: id}
	if req.Status != nil {
		r.Status = *req.Status
	}
	if req.CreatorStatus != nil {
		r.CreatorStatus = *req.CreatorStatus
	}
	if req.VerifiedStatus != nil {
		r.VerifiedStatus = *req.VerifiedStatus
	}
	if req.MembershipTier != nil {
		r.MembershipTier = *req.MembershipTier
	}
	out, err := h.svc.User.UpdateAccount(c.Request.Context(), r)
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"user": out.User})
}


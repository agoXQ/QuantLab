package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Follow handles POST /api/v1/users/:id/follow. The path param :id is
// the followee; the body / query supplies the follower id (no auth
// middleware in the MVP).
func (h *Handler) Follow(c *gin.Context) {
	followeeID, ok := parseID(c, "id")
	if !ok {
		return
	}
	followerID := readFollowerID(c)
	if followerID <= 0 {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, "invalid follower_id"))
		return
	}
	if err := h.svc.Follow(c.Request.Context(), appUser.FollowRequest{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, gin.H{"follower_id": followerID, "followee_id": followeeID})
}

// Unfollow handles DELETE /api/v1/users/:id/follow.
func (h *Handler) Unfollow(c *gin.Context) {
	followeeID, ok := parseID(c, "id")
	if !ok {
		return
	}
	followerID := readFollowerID(c)
	if followerID <= 0 {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, "invalid follower_id"))
		return
	}
	if err := h.svc.Unfollow(c.Request.Context(), appUser.FollowRequest{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// ListFollowers handles GET /api/v1/users/:id/followers.
func (h *Handler) ListFollowers(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	limit, offset := readPaging(c)
	res, err := h.svc.ListFollowers(c.Request.Context(), id, limit, offset)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": res.Users})
}

// ListFollowing handles GET /api/v1/users/:id/following.
func (h *Handler) ListFollowing(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	limit, offset := readPaging(c)
	res, err := h.svc.ListFollowing(c.Request.Context(), id, limit, offset)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": res.Users})
}

// readFollowerID accepts either a query param or a JSON body so the
// MVP route works from curl and from a browser session without needing
// a fully wired auth middleware. Production deployments should pull
// the id from the verified JWT claims via middleware.
func readFollowerID(c *gin.Context) int64 {
	if v := parseInt64(c.Query("follower_id")); v > 0 {
		return v
	}
	var req followRequest
	_ = c.ShouldBindJSON(&req)
	return req.FollowerID
}

func readPaging(c *gin.Context) (int, int) {
	limit := int(parseInt64(c.Query("limit")))
	offset := int(parseInt64(c.Query("offset")))
	return limit, offset
}

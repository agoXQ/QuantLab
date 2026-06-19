// Package logic wires the gRPC handlers to the application service so
// the same use cases that drive the HTTP API drive the RPC surface
// without duplicating business code. Each handler is a thin adapter:
// translate the protobuf request into an application DTO, call the
// service, render the protobuf response.
package logic

import (
	"context"
	"strings"

	"github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
	domuser "github.com/agoXQ/QuantLab/app/user/domain/user"
	"github.com/agoXQ/QuantLab/app/user/pb"
)

// userToProto renders a domain User into the protobuf shape.
func userToProto(u *domuser.User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Avatar:         u.Avatar,
		Bio:            u.Bio,
		Status:         int32(u.Status),
		CreatorStatus:  int32(u.CreatorStatus),
		VerifiedStatus: int32(u.VerifiedStatus),
		MembershipTier: string(u.MembershipTier),
		CreatedAt:      u.CreatedAt.Unix(),
	}
}

// usersToProto renders a slice of users for the follow listing
// endpoints; both sides use the same conversion so any future field
// addition lands once.
func usersToProto(in []*domuser.User) []*pb.User {
	out := make([]*pb.User, 0, len(in))
	for _, u := range in {
		out = append(out, userToProto(u))
	}
	return out
}

// optionalString returns nil when s is empty so the application layer
// can distinguish "unset" from "set to empty"; the gRPC API does not
// expose explicit nullability today, so an empty string from the wire
// is treated as "leave as-is".
func optionalString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// userIDFromContext returns the caller id stamped onto ctx by the
// auth interceptor. The interceptor accepts either an Authorization:
// Bearer JWT or a metadata "x-user-id" fallback so the gRPC adapters
// stay free of token plumbing.
func userIDFromContext(ctx context.Context) int64 {
	return middleware.UserIDFromContext(ctx)
}

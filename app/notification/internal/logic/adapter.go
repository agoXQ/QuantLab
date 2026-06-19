// Package logic wires the gRPC handlers to the application service so
// the same use cases that drive the HTTP API drive the RPC surface
// without duplicating business code. Each handler is a thin adapter:
// translate the protobuf request into an application DTO, call the
// service, render the protobuf response.
package logic

import (
	"context"

	commonv1 "github.com/agoXQ/QuantLab/api/common/v1"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	"github.com/agoXQ/QuantLab/app/notification/pb"
	"github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
)

// notificationToProto renders a domain Notification into the protobuf
// shape. Status is encoded as the canonical upper-case name; gRPC
// clients that want the int form can re-derive it from the string.
func notificationToProto(n *domNotif.Notification) *pb.Notification {
	if n == nil {
		return nil
	}
	view := &pb.Notification{
		Id:        n.ID,
		UserId:    n.UserID,
		Type:      pb.NotificationType(n.Type),
		Title:     n.Title,
		Content:   n.Content,
		Status:    n.Status.String(),
		CreatedAt: n.CreatedAt.Unix(),
	}
	if n.ReadAt != nil {
		view.ReadAt = n.ReadAt.Unix()
	}
	return view
}

func notificationsToProto(in []*domNotif.Notification) []*pb.Notification {
	out := make([]*pb.Notification, 0, len(in))
	for _, n := range in {
		out = append(out, notificationToProto(n))
	}
	return out
}

func subscriptionToProto(s *domSub.Subscription) *pb.NotificationSubscription {
	if s == nil {
		return nil
	}
	return &pb.NotificationSubscription{
		Id:           s.ID,
		SubscriberId: s.SubscriberID,
		ObjectType:   s.ObjectType,
		ObjectId:     s.ObjectID,
		CreatedAt:    s.CreatedAt.Unix(),
	}
}

func subscriptionsToProto(in []*domSub.Subscription) []*pb.NotificationSubscription {
	out := make([]*pb.NotificationSubscription, 0, len(in))
	for _, s := range in {
		out = append(out, subscriptionToProto(s))
	}
	return out
}

// cursorString reads a cursor from the protobuf wrapper without
// panicking on nil; an empty string asks the service for the first page.
func cursorString(c *commonv1.Cursor) string {
	if c == nil {
		return ""
	}
	return c.NextCursor
}

// cursorProto wraps next into the platform Cursor envelope. has_more
// flips when the service produced a non-empty cursor.
func cursorProto(next string) *commonv1.Cursor {
	return &commonv1.Cursor{NextCursor: next, HasMore: next != ""}
}

// userIDFromContext returns the caller id stamped onto ctx by the
// shared auth interceptor. Notification reuses the User-side
// middleware so a single secret + key set drives every service.
func userIDFromContext(ctx context.Context) int64 {
	return middleware.UserIDFromContext(ctx)
}

package notification

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	notifErr "github.com/agoXQ/QuantLab/app/notification/domain/errors"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
)

// DefaultPageLimit is applied when the caller passes Limit <= 0.
const DefaultPageLimit = 20

// MaxPageLimit caps a single page so a misbehaving client cannot drain
// the whole table in one request.
const MaxPageLimit = 100

// Service is the orchestration surface used by every adapter.
type Service interface {
	CreateNotification(ctx context.Context, in CreateNotificationInput) (*domNotif.Notification, error)
	ListNotifications(ctx context.Context, in ListNotificationsInput) (ListNotificationsOutput, error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID, id int64) error
	MarkAllRead(ctx context.Context, userID int64) (int64, error)
	DeleteNotification(ctx context.Context, userID, id int64) error

	GetPreferences(ctx context.Context, userID int64) (*domPref.Preference, error)
	UpdatePreferences(ctx context.Context, in UpdatePreferencesInput) (*domPref.Preference, error)

	CreateSubscription(ctx context.Context, in CreateSubscriptionInput) (*domSub.Subscription, error)
	CancelSubscription(ctx context.Context, subscriberID, id int64) error
	ListSubscriptions(ctx context.Context, in ListSubscriptionsInput) (ListSubscriptionsOutput, error)
}

type service struct {
	deps Dependencies
}

// NewService wires the application service. Repositories are required;
// the clock defaults to time.Now when unset.
func NewService(deps Dependencies) Service {
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	return &service{deps: deps}
}

func (s *service) now() time.Time { return s.deps.Clock() }

func clampLimit(limit int) int {
	if limit <= 0 {
		return DefaultPageLimit
	}
	if limit > MaxPageLimit {
		return MaxPageLimit
	}
	return limit
}

func decodeOffsetCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(cursor))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func encodeOffsetCursor(offset int) string {
	if offset <= 0 {
		return ""
	}
	return strconv.Itoa(offset)
}

func (s *service) CreateNotification(ctx context.Context, in CreateNotificationInput) (*domNotif.Notification, error) {
	if in.UserID <= 0 {
		return nil, notifErr.ErrInvalidUserID
	}
	if !in.Type.IsValid() {
		return nil, notifErr.ErrInvalidType
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, notifErr.ErrInvalidNotification
	}
	notif := &domNotif.Notification{
		UserID:    in.UserID,
		Type:      in.Type,
		Title:     title,
		Content:   in.Content,
		Status:    valueobject.NotificationStatusUnread,
		CreatedAt: s.now(),
	}
	if err := notif.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Notifications.Create(ctx, notif); err != nil {
		return nil, err
	}
	return notif, nil
}

func (s *service) ListNotifications(ctx context.Context, in ListNotificationsInput) (ListNotificationsOutput, error) {
	if in.UserID <= 0 {
		return ListNotificationsOutput{}, notifErr.ErrInvalidUserID
	}
	limit := clampLimit(in.Limit)
	offset := decodeOffsetCursor(in.Cursor)
	statuses := in.Statuses
	if len(statuses) == 0 {
		statuses = []valueobject.NotificationStatus{
			valueobject.NotificationStatusUnread,
			valueobject.NotificationStatusRead,
		}
	}
	rows, err := s.deps.Notifications.List(ctx, domNotif.ListFilter{
		UserID:   in.UserID,
		Statuses: statuses,
		Types:    in.Types,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return ListNotificationsOutput{}, err
	}
	next := ""
	if len(rows) == limit {
		next = encodeOffsetCursor(offset + len(rows))
	}
	return ListNotificationsOutput{Items: rows, NextCursor: next}, nil
}

func (s *service) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, notifErr.ErrInvalidUserID
	}
	return s.deps.Notifications.CountUnread(ctx, userID)
}

func (s *service) MarkRead(ctx context.Context, userID, id int64) error {
	if userID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	if id <= 0 {
		return notifErr.ErrNotificationNotFound
	}
	return s.deps.Notifications.MarkRead(ctx, userID, id, s.now())
}

func (s *service) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, notifErr.ErrInvalidUserID
	}
	return s.deps.Notifications.MarkAllRead(ctx, userID, s.now())
}

func (s *service) DeleteNotification(ctx context.Context, userID, id int64) error {
	if userID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	if id <= 0 {
		return notifErr.ErrNotificationNotFound
	}
	return s.deps.Notifications.Delete(ctx, userID, id)
}

func (s *service) GetPreferences(ctx context.Context, userID int64) (*domPref.Preference, error) {
	if userID <= 0 {
		return nil, notifErr.ErrInvalidUserID
	}
	pref, err := s.deps.Preferences.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, notifErr.ErrPreferenceNotFound) {
			defaults := domPref.Defaults(userID, s.now())
			return &defaults, nil
		}
		return nil, err
	}
	return pref, nil
}

func (s *service) UpdatePreferences(ctx context.Context, in UpdatePreferencesInput) (*domPref.Preference, error) {
	if in.UserID <= 0 {
		return nil, notifErr.ErrInvalidUserID
	}
	pref := &domPref.Preference{
		UserID:         in.UserID,
		InAppEnabled:   in.InAppEnabled,
		EmailEnabled:   in.EmailEnabled,
		WebhookEnabled: in.WebhookEnabled,
		PushEnabled:    in.PushEnabled,
		UpdatedAt:      s.now(),
	}
	if err := pref.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Preferences.Upsert(ctx, pref); err != nil {
		return nil, err
	}
	return pref, nil
}

func (s *service) CreateSubscription(ctx context.Context, in CreateSubscriptionInput) (*domSub.Subscription, error) {
	if in.SubscriberID <= 0 {
		return nil, notifErr.ErrInvalidUserID
	}
	objType := domSub.NormaliseObjectType(in.ObjectType)
	if objType == "" {
		return nil, notifErr.ErrInvalidObjectType
	}
	if in.ObjectID <= 0 {
		return nil, notifErr.ErrInvalidObjectID
	}
	exists, err := s.deps.Subscriptions.ExistsByObject(ctx, in.SubscriberID, objType, in.ObjectID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, notifErr.ErrSubscriptionConflict
	}
	sub := &domSub.Subscription{
		SubscriberID: in.SubscriberID,
		ObjectType:   objType,
		ObjectID:     in.ObjectID,
		CreatedAt:    s.now(),
	}
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *service) CancelSubscription(ctx context.Context, subscriberID, id int64) error {
	if subscriberID <= 0 {
		return notifErr.ErrInvalidUserID
	}
	if id <= 0 {
		return notifErr.ErrSubscriptionNotFound
	}
	return s.deps.Subscriptions.Delete(ctx, subscriberID, id)
}

func (s *service) ListSubscriptions(ctx context.Context, in ListSubscriptionsInput) (ListSubscriptionsOutput, error) {
	if in.SubscriberID <= 0 {
		return ListSubscriptionsOutput{}, notifErr.ErrInvalidUserID
	}
	limit := clampLimit(in.Limit)
	offset := decodeOffsetCursor(in.Cursor)
	rows, err := s.deps.Subscriptions.List(ctx, domSub.ListFilter{
		SubscriberID: in.SubscriberID,
		ObjectType:   domSub.NormaliseObjectType(in.ObjectType),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return ListSubscriptionsOutput{}, err
	}
	next := ""
	if len(rows) == limit {
		next = encodeOffsetCursor(offset + len(rows))
	}
	return ListSubscriptionsOutput{Items: rows, NextCursor: next}, nil
}

// Package errors defines domain-level errors for the Notification
// Service.
//
// Codes follow the platform-wide allocation in pkg/errors. The
// Notification service reserves 40060-40069 for resource errors,
// 10060-10069 for validation, 20060-20069 for state-machine refusals,
// 30060-30069 for authentication / authorisation, and 50060-50069 for
// system errors. The shape mirrors User / Strategy / Backtest so the
// HTTP adapter can translate every domain error with one helper.
package errors

import "fmt"

// NotificationError represents a domain-level error with a numeric
// code.
type NotificationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *NotificationError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a NotificationError.
func New(code int, message string) *NotificationError {
	return &NotificationError{Code: code, Message: message}
}

// 4xxxx — Resource errors.
var (
	ErrNotificationNotFound = New(40060, "notification not found")
	ErrPreferenceNotFound   = New(40061, "notification preference not found")
	ErrSubscriptionNotFound = New(40062, "notification subscription not found")
)

// 1xxxx — Validation errors.
var (
	ErrInvalidNotification = New(10060, "invalid notification")
	ErrInvalidUserID       = New(10061, "user id is required")
	ErrInvalidType         = New(10062, "invalid notification type")
	ErrInvalidStatus       = New(10063, "invalid notification status")
	ErrInvalidObjectType   = New(10064, "invalid subscription object type")
	ErrInvalidObjectID     = New(10065, "invalid subscription object id")
)

// 2xxxx — Conflict / state-machine refusals.
var (
	ErrAlreadyRead          = New(20060, "notification already read")
	ErrAlreadyDeleted       = New(20061, "notification already deleted")
	ErrSubscriptionConflict = New(20062, "subscription already exists")
)

// 3xxxx — Authentication / authorisation errors.
var (
	ErrForbidden = New(30060, "forbidden")
)

// 5xxxx — System errors.
var (
	ErrPersistence = New(50060, "notification persistence error")
)

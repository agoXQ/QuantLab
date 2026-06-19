// Package valueobject defines shared value objects for the
// Notification Service domain. The MVP models two enums:
//
//   - NotificationType — the platform-wide categorisation that the
//     notifications proto already publishes.
//   - NotificationStatus — the read / unread / deleted lifecycle the
//     repository uses to filter rows.
//
// Both values are persisted as compact integers so the schema mirrors
// other services (Strategy lifecycle, User account) without dragging
// extra string tables along.
package valueobject

import "strings"

// NotificationType lists the canonical notification categories. The
// numeric values stay in sync with api/notification/v1.NotificationType
// so the gRPC adapter can convert without a translation table.
type NotificationType int32

const (
	NotificationTypeUnspecified NotificationType = 0
	NotificationTypeSystem      NotificationType = 1
	NotificationTypeComment     NotificationType = 2
	NotificationTypeLike        NotificationType = 3
	NotificationTypeFollow      NotificationType = 4
	NotificationTypeMention     NotificationType = 5
	NotificationTypeRanking     NotificationType = 6
	NotificationTypeStrategy    NotificationType = 7
	NotificationTypePortfolio   NotificationType = 8
	NotificationTypeBacktest    NotificationType = 9
	NotificationTypeMembership  NotificationType = 10
)

// IsValid reports whether t is a recognised value (excluding the
// unspecified zero, which the repository refuses to persist).
func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeSystem,
		NotificationTypeComment,
		NotificationTypeLike,
		NotificationTypeFollow,
		NotificationTypeMention,
		NotificationTypeRanking,
		NotificationTypeStrategy,
		NotificationTypePortfolio,
		NotificationTypeBacktest,
		NotificationTypeMembership:
		return true
	default:
		return false
	}
}

// String returns the canonical upper-case name; the HTTP layer uses it
// when echoing notifications back to the caller.
func (t NotificationType) String() string {
	switch t {
	case NotificationTypeSystem:
		return "SYSTEM"
	case NotificationTypeComment:
		return "COMMENT"
	case NotificationTypeLike:
		return "LIKE"
	case NotificationTypeFollow:
		return "FOLLOW"
	case NotificationTypeMention:
		return "MENTION"
	case NotificationTypeRanking:
		return "RANKING"
	case NotificationTypeStrategy:
		return "STRATEGY"
	case NotificationTypePortfolio:
		return "PORTFOLIO"
	case NotificationTypeBacktest:
		return "BACKTEST"
	case NotificationTypeMembership:
		return "MEMBERSHIP"
	default:
		return "UNSPECIFIED"
	}
}

// ParseNotificationType parses a string into the matching enum,
// case-insensitively. Unknown values map to NotificationTypeUnspecified.
func ParseNotificationType(s string) NotificationType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SYSTEM":
		return NotificationTypeSystem
	case "COMMENT":
		return NotificationTypeComment
	case "LIKE":
		return NotificationTypeLike
	case "FOLLOW":
		return NotificationTypeFollow
	case "MENTION":
		return NotificationTypeMention
	case "RANKING":
		return NotificationTypeRanking
	case "STRATEGY":
		return NotificationTypeStrategy
	case "PORTFOLIO":
		return NotificationTypePortfolio
	case "BACKTEST":
		return NotificationTypeBacktest
	case "MEMBERSHIP":
		return NotificationTypeMembership
	default:
		return NotificationTypeUnspecified
	}
}

// NotificationStatus tracks the read lifecycle. Deleted rows stay in
// the table so analytics can still count notifications produced; the
// repository filters them out of the user-facing list.
type NotificationStatus int32

const (
	NotificationStatusUnknown NotificationStatus = 0
	NotificationStatusUnread  NotificationStatus = 1
	NotificationStatusRead    NotificationStatus = 2
	NotificationStatusDeleted NotificationStatus = 3
)

// IsValid reports whether s is a recognised value.
func (s NotificationStatus) IsValid() bool {
	switch s {
	case NotificationStatusUnread, NotificationStatusRead, NotificationStatusDeleted:
		return true
	default:
		return false
	}
}

// String returns the canonical upper-case name.
func (s NotificationStatus) String() string {
	switch s {
	case NotificationStatusUnread:
		return "UNREAD"
	case NotificationStatusRead:
		return "READ"
	case NotificationStatusDeleted:
		return "DELETED"
	default:
		return "UNKNOWN"
	}
}

// ParseNotificationStatus parses a string into the matching enum,
// case-insensitively. Unknown values map to NotificationStatusUnknown.
func ParseNotificationStatus(s string) NotificationStatus {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UNREAD":
		return NotificationStatusUnread
	case "READ":
		return NotificationStatusRead
	case "DELETED":
		return NotificationStatusDeleted
	default:
		return NotificationStatusUnknown
	}
}

// Package valueobject defines shared value objects for the User
// Service domain. The platform standard reserves four user-state
// dimensions: account status (active / suspended / banned), creator
// status (regular / pro / verified), verified status (none / phone /
// id), and membership tier. We model each as a typed enum so the
// repository layer can reject bad values at parse time.
package valueobject

import (
	"strings"
)

// AccountStatus reflects whether the account is usable.
type AccountStatus int32

const (
	AccountStatusUnknown   AccountStatus = 0
	AccountStatusActive    AccountStatus = 1
	AccountStatusSuspended AccountStatus = 2
	AccountStatusBanned    AccountStatus = 3
	AccountStatusDeleted   AccountStatus = 4
)

// IsActive reports whether the user can log in / publish actions.
func (s AccountStatus) IsActive() bool { return s == AccountStatusActive }

// IsValid reports whether the status is a recognised value.
func (s AccountStatus) IsValid() bool {
	switch s {
	case AccountStatusActive, AccountStatusSuspended, AccountStatusBanned, AccountStatusDeleted:
		return true
	default:
		return false
	}
}

// CreatorStatus tracks the user's creator program tier.
type CreatorStatus int32

const (
	CreatorStatusRegular CreatorStatus = 0
	CreatorStatusPro     CreatorStatus = 1
	CreatorStatusVerified CreatorStatus = 2
)

// IsValid reports whether c is a recognised value.
func (c CreatorStatus) IsValid() bool {
	switch c {
	case CreatorStatusRegular, CreatorStatusPro, CreatorStatusVerified:
		return true
	default:
		return false
	}
}

// VerifiedStatus tracks the strongest verification on file.
type VerifiedStatus int32

const (
	VerifiedStatusNone  VerifiedStatus = 0
	VerifiedStatusEmail VerifiedStatus = 1
	VerifiedStatusPhone VerifiedStatus = 2
	VerifiedStatusID    VerifiedStatus = 3
)

// IsValid reports whether v is a recognised value.
func (v VerifiedStatus) IsValid() bool {
	switch v {
	case VerifiedStatusNone, VerifiedStatusEmail, VerifiedStatusPhone, VerifiedStatusID:
		return true
	default:
		return false
	}
}

// MembershipTier names the paid plan a user belongs to. The MVP only
// uses Free / Pro / Enterprise; the value is stored as a short string
// so the catalogue can grow without a schema migration.
type MembershipTier string

const (
	MembershipTierFree       MembershipTier = "FREE"
	MembershipTierPro        MembershipTier = "PRO"
	MembershipTierEnterprise MembershipTier = "ENTERPRISE"
)

// IsValid reports whether t is a recognised tier.
func (t MembershipTier) IsValid() bool {
	switch t {
	case MembershipTierFree, MembershipTierPro, MembershipTierEnterprise:
		return true
	default:
		return false
	}
}

// ParseMembershipTier accepts the canonical labels in any case; empty
// string yields the Free tier so callers can opt out of explicit
// configuration. Unknown labels surface as ("", false) so the caller
// can map the error to a 400.
func ParseMembershipTier(s string) (MembershipTier, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return MembershipTierFree, true
	}
	t := MembershipTier(s)
	if !t.IsValid() {
		return "", false
	}
	return t, true
}

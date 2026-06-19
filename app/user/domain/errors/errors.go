// Package errors defines domain-level errors for the User Service.
//
// Codes follow the platform-wide allocation in pkg/errors. The User
// service reserves 40010-40019 for resource errors, 10010-10019 for
// validation, 20010-20019 for state-machine refusals, 30010-30019 for
// authentication / authorisation, and 50010-50019 for system errors.
package errors

import "fmt"

// UserError represents a domain-level error with a numeric code.
type UserError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *UserError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a UserError.
func New(code int, message string) *UserError {
	return &UserError{Code: code, Message: message}
}

// 4xxxx — Resource errors.
var (
	ErrUserNotFound     = New(40010, "user not found")
	ErrFollowNotFound   = New(40011, "follow relation not found")
)

// 1xxxx — Validation errors.
var (
	ErrInvalidUser     = New(10010, "invalid user")
	ErrEmailRequired   = New(10011, "email is required")
	ErrEmailInvalid    = New(10012, "email format is invalid")
	ErrUsernameInvalid = New(10013, "username is invalid")
	ErrPasswordTooWeak = New(10014, "password is too weak")
	ErrInvalidStatus   = New(10015, "invalid account status")
	ErrInvalidTier     = New(10016, "invalid membership tier")
	ErrSelfFollow      = New(10017, "cannot follow yourself")
)

// 2xxxx — Conflict / state-machine refusals.
var (
	ErrEmailTaken    = New(20010, "email is already taken")
	ErrUsernameTaken = New(20011, "username is already taken")
	ErrAlreadyFollowed = New(20012, "already followed")
	ErrAccountSuspended = New(20013, "account is suspended")
	ErrAccountBanned    = New(20014, "account is banned")
	ErrAccountDeleted   = New(20015, "account is deleted")
)

// 3xxxx — Authentication / authorisation errors.
var (
	ErrInvalidCredentials = New(30010, "invalid credentials")
	ErrTokenInvalid       = New(30011, "token is invalid")
	ErrTokenExpired       = New(30012, "token is expired")
	ErrUnauthorized       = New(30013, "unauthorized")
)

// 5xxxx — System errors.
var (
	ErrPersistence = New(50010, "user persistence error")
)

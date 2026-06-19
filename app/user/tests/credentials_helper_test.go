package tests

import (
	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
)

// userInvalidCredentials returns the platform's ErrInvalidCredentials
// without making the fake hasher import the domain package twice; the
// helper exists so the fakeHasher above stays a tight value type.
func userInvalidCredentials() error { return userErr.ErrInvalidCredentials }

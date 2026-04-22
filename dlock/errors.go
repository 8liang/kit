package dlock

import "errors"

var (
	ErrLockFailed   = errors.New("failed to acquire lock")
	ErrInvalidToken = errors.New("invalid token or lock expired")
)

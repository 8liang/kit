package dlock

import (
	"context"
)

// Locker is the unified interface for distributed locks.
type Locker interface {
	// Lock acquires the lock, blocking until it succeeds or ctx is canceled.
	Lock(ctx context.Context, key string, opts ...Option) (Lock, error)

	// TryLock attempts to acquire the lock without blocking.
	TryLock(ctx context.Context, key string, opts ...Option) (Lock, bool, error)
}

// Lock represents an acquired distributed lock.
type Lock interface {
	// Unlock releases the lock. It must be idempotent.
	Unlock(ctx context.Context) error

	// Valid checks if the lock is still valid (non-blocking).
	Valid() bool

	// Done returns a channel that's closed when the lock is lost or unlocked.
	Done() <-chan struct{}
}

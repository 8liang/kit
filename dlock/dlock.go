package dlock

import "context"

type Locker interface {
	Lock(ctx context.Context, key string, opts ...Option) (Lock, error)
	TryLock(ctx context.Context, key string, opts ...Option) (Lock, bool, error)
}

type Lock interface {
	Key() string
	Token() string
	Unlock(ctx context.Context) error
	Refresh(ctx context.Context) error
}

package dlock

import (
	"context"
)

// DoWithLock acquires the lock, runs the function fn, and safely releases the lock.
// It returns an error if the lock cannot be acquired, or if fn returns an error.
func DoWithLock(ctx context.Context, locker Locker, key string, fn func() error, opts ...Option) error {
	lock, err := locker.Lock(ctx, key, opts...)
	if err != nil {
		return err
	}
	// Use Background context to ensure unlock succeeds even if the original ctx is canceled
	defer lock.Unlock(context.Background())

	return fn()
}

// TryDoWithLock attempts to acquire the lock non-blocking, runs the function fn, and safely releases the lock.
// It returns a boolean indicating if the lock was acquired, and an error if fn returns an error or backend fails.
func TryDoWithLock(ctx context.Context, locker Locker, key string, fn func() error, opts ...Option) (bool, error) {
	lock, ok, err := locker.TryLock(ctx, key, opts...)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	defer lock.Unlock(context.Background())

	return true, fn()
}

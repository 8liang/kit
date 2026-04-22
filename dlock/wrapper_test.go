package dlock

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockLocker and mockLock for testing
type mockLocker struct {
	lockErr error
	tryOk   bool
	tryErr  error
	locked  bool
}

type mockLock struct {
	locker *mockLocker
}

func (l *mockLocker) Lock(ctx context.Context, key string, opts ...Option) (Lock, error) {
	if l.lockErr != nil {
		return nil, l.lockErr
	}
	l.locked = true
	return &mockLock{locker: l}, nil
}

func (l *mockLocker) TryLock(ctx context.Context, key string, opts ...Option) (Lock, bool, error) {
	if l.tryErr != nil {
		return nil, false, l.tryErr
	}
	if !l.tryOk {
		return nil, false, nil
	}
	l.locked = true
	return &mockLock{locker: l}, true, nil
}

func (m *mockLock) Unlock(ctx context.Context) error {
	m.locker.locked = false
	return nil
}

func (m *mockLock) Valid() bool {
	return m.locker.locked
}

func (m *mockLock) Done() <-chan struct{} {
	return make(chan struct{})
}

func TestDoWithLock(t *testing.T) {
	locker := &mockLocker{}

	// Test success
	called := false
	err := DoWithLock(context.Background(), locker, "key", func() error {
		called = true
		assert.True(t, locker.locked)
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, called)
	assert.False(t, locker.locked) // Should be unlocked

	// Test function error
	expectedErr := errors.New("fn error")
	err = DoWithLock(context.Background(), locker, "key", func() error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
	assert.False(t, locker.locked)

	// Test lock error
	locker.lockErr = errors.New("lock error")
	called = false
	err = DoWithLock(context.Background(), locker, "key", func() error {
		called = true
		return nil
	})
	assert.Equal(t, locker.lockErr, err)
	assert.False(t, called)
}

func TestTryDoWithLock(t *testing.T) {
	locker := &mockLocker{}

	// Test success
	locker.tryOk = true
	called := false
	ok, err := TryDoWithLock(context.Background(), locker, "key", func() error {
		called = true
		assert.True(t, locker.locked)
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, called)
	assert.False(t, locker.locked) // Should be unlocked

	// Test not acquired
	locker.tryOk = false
	called = false
	ok, err = TryDoWithLock(context.Background(), locker, "key", func() error {
		called = true
		return nil
	})
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, called)

	// Test backend error
	locker.tryErr = errors.New("try error")
	called = false
	ok, err = TryDoWithLock(context.Background(), locker, "key", func() error {
		called = true
		return nil
	})
	assert.Equal(t, locker.tryErr, err)
	assert.False(t, ok)
	assert.False(t, called)

	// Test function error
	locker.tryErr = nil
	locker.tryOk = true
	expectedErr := errors.New("fn error")
	called = false
	ok, err = TryDoWithLock(context.Background(), locker, "key", func() error {
		called = true
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
	assert.True(t, ok)
	assert.True(t, called)
	assert.False(t, locker.locked)
}

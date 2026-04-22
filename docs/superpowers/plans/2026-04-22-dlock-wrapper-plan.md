# dlock Wrapper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create generic functional wrappers (`DoWithLock` and `TryDoWithLock`) to reduce boilerplate when acquiring and releasing distributed locks.

**Architecture:** A new file `wrapper.go` will be added to the `dlock` package. It will provide high-level functions that accept a `Locker`, context, key, options, and a closure `fn func() error`. The wrapper will handle lock acquisition, ensure deferred unlocking, and execute the closure.

**Tech Stack:** Go

---

### Task 1: Implement Wrapper Functions

**Files:**
- Create: `dlock/wrapper.go`
- Create: `dlock/wrapper_test.go`

- [ ] **Step 1: Write the wrapper code**

Create `dlock/wrapper.go`:
```go
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
```

- [ ] **Step 2: Write tests for wrappers**

Create `dlock/wrapper_test.go`:
```go
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
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./dlock -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add dlock/wrapper.go dlock/wrapper_test.go
git commit -m "feat(dlock): add wrapper functions DoWithLock and TryDoWithLock"
```

# etcdlock Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the `etcdlock` implementation to use the official `go.etcd.io/etcd/client/v3/concurrency` package, solving auto-renewal, watch-based queueing, and reducing lease creation overhead without modifying the core `dlock.Locker` and `dlock.Lock` interfaces.

**Architecture:** 
- Instead of manually tracking `client.Grant` and `client.Txn`, use `concurrency.NewSession` and `concurrency.NewMutex`.
- `TryLock` creates a session and calls `mutex.TryLock`.
- `Lock` creates a session and calls `mutex.Lock` (which automatically uses etcd watch for efficient queueing).
- `Unlock` simply calls `mutex.Unlock` and then `session.Close()` to clean up.
- `Refresh` acts as a health check by checking `session.Done()`.

**Tech Stack:** Go, `go.etcd.io/etcd/client/v3/concurrency`

---

### Task 1: Refactor `etcdlock.go`

**Files:**
- Modify: `dlock/etcdlock/etcdlock.go`

- [ ] **Step 1: Write the minimal implementation**

Overwrite `dlock/etcdlock/etcdlock.go` with the following code:

```go
package etcdlock

import (
	"context"
	"fmt"
	"time"

	"github.com/8liang/kit/dlock"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const minLeaseTTL = 1

type etcdLocker struct {
	client *clientv3.Client
}

type etcdLock struct {
	key     string
	token   string
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func New(client *clientv3.Client) dlock.Locker {
	return &etcdLocker{client: client}
}

func (l *etcdLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	options := dlock.NewOptions(opts...)

	ttlSeconds := int(options.TTL.Seconds())
	if options.TTL.Seconds() > float64(ttlSeconds) {
		ttlSeconds++
	}
	if ttlSeconds <= 0 {
		ttlSeconds = minLeaseTTL
	}

	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(ttlSeconds), concurrency.WithContext(ctx))
	if err != nil {
		return nil, false, err
	}

	mutex := concurrency.NewMutex(session, key)

	err = mutex.TryLock(ctx)
	if err != nil {
		session.Close()
		if err == concurrency.ErrLocked {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &etcdLock{
		key:     key,
		token:   options.Token,
		session: session,
		mutex:   mutex,
	}, true, nil
}

func (l *etcdLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	options := dlock.NewOptions(opts...)

	ttlSeconds := int(options.TTL.Seconds())
	if options.TTL.Seconds() > float64(ttlSeconds) {
		ttlSeconds++
	}
	if ttlSeconds <= 0 {
		ttlSeconds = minLeaseTTL
	}

	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(ttlSeconds), concurrency.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	mutex := concurrency.NewMutex(session, key)

	err = mutex.Lock(ctx)
	if err != nil {
		session.Close()
		return nil, err
	}

	return &etcdLock{
		key:     key,
		token:   options.Token,
		session: session,
		mutex:   mutex,
	}, nil
}

func (l *etcdLock) Key() string {
	return l.key
}

func (l *etcdLock) Token() string {
	// Note: etcd concurrency.Mutex doesn't expose a way to set custom values/tokens natively on the lock node easily,
	// but we fulfill the interface requirement by returning the token from options.
	return l.token
}

func (l *etcdLock) Unlock(ctx context.Context) error {
	defer func() {
		// Closing the session revokes the lease and deletes the lock keys automatically
		_ = l.session.Close()
	}()

	err := l.mutex.Unlock(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *etcdLock) Refresh(ctx context.Context) error {
	// The session handles keep-alive automatically in the background.
	// We just check if the session is still alive (i.e. we haven't lost the lock due to network partition/expiration).
	select {
	case <-l.session.Done():
		return fmt.Errorf("%w: etcd session expired or closed", dlock.ErrInvalidToken)
	default:
		return nil
	}
}
```

- [ ] **Step 2: Run build to verify**

Run: `go build ./dlock/etcdlock`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add dlock/etcdlock/etcdlock.go
git commit -m "refactor(dlock): use etcd concurrency.Mutex for robust locking"
```

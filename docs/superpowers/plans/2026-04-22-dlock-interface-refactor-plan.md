# dlock Interface Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the `Lock` interface to support idempotent `Unlock`, non-blocking `Valid`, and lock loss notification `Done`, and implement background watchdog renewals for Redis and MongoDB backends.

**Architecture:** 
- The core `dlock.Lock` interface changes to `Unlock`, `Valid`, and `Done`.
- `etcdlock` uses its existing `concurrency.Session` to provide `Done()` and `Valid()`.
- `redislock` and `mongolock` start a background watchdog goroutine during `TryLock`/`Lock`. The watchdog auto-renews the TTL every `TTL / 3`. If it fails, it closes the `doneCh` to notify callers that the lock is lost.
- All backends use `sync.Once` to ensure `Unlock` is safe and idempotent.

**Tech Stack:** Go, Redis, etcd, MongoDB

---

### Task 1: Update Core Interfaces

**Files:**
- Modify: `dlock/dlock.go`

- [ ] **Step 1: Write the updated interface**

Update `dlock/dlock.go` to match the new design:
```go
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
```

- [ ] **Step 2: Commit**

Run:
```bash
git add dlock/dlock.go
git commit -m "refactor(dlock): update Lock interface with Valid and Done methods"
```

---

### Task 2: Refactor etcdlock

**Files:**
- Modify: `dlock/etcdlock/etcdlock.go`

- [ ] **Step 1: Implement new interface methods**

In `dlock/etcdlock/etcdlock.go`, update the `etcdLock` struct and its methods:
```go
// ... (keep package, imports, etcdLocker, getTTLSeconds, TryLock, Lock exactly as they are) ...

// (Inside TryLock and Lock, make sure to add `once: sync.Once{}` to the returned struct if needed, or initialize it).
// Update the etcdLock struct:
import "sync" // ensure sync is imported

type etcdLock struct {
	key     string
	token   string
	session *concurrency.Session
	mutex   *concurrency.Mutex
	once    sync.Once
	err     error
}

// Replace Key(), Token(), Refresh() with the new interface methods:

func (l *etcdLock) Unlock(ctx context.Context) error {
	l.once.Do(func() {
		defer func() {
			_ = l.session.Close()
		}()
		l.err = l.mutex.Unlock(ctx)
	})
	return l.err
}

func (l *etcdLock) Valid() bool {
	select {
	case <-l.session.Done():
		return false
	default:
		return true
	}
}

func (l *etcdLock) Done() <-chan struct{} {
	return l.session.Done()
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./dlock/etcdlock`
Expected: PASS

- [ ] **Step 3: Commit**

Run:
```bash
git add dlock/etcdlock/etcdlock.go
git commit -m "refactor(dlock): implement new Lock interface for etcdlock"
```

---

### Task 3: Refactor redislock

**Files:**
- Modify: `dlock/redislock/redislock.go`

- [ ] **Step 1: Implement watchdog and new interface**

In `dlock/redislock/redislock.go`, update the `redisLock` struct and methods:
```go
// Add "sync" to imports

type redisLock struct {
	client *redis.Client
	key    string
	token  string
	ttl    time.Duration
	doneCh chan struct{}
	stopCh chan struct{}
	once   sync.Once
	err    error
}

// Update TryLock to initialize channels and start watchdog
func (l *redisLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	o := dlock.NewOptions(opts...)
	
	ok, err := l.client.SetNX(ctx, key, o.Token, o.TTL).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	lock := &redisLock{
		client: l.client,
		key:    key,
		token:  o.Token,
		ttl:    o.TTL,
		doneCh: make(chan struct{}),
		stopCh: make(chan struct{}),
	}
	
	lock.startWatchdog()
	return lock, true, nil
}

func (l *redisLock) startWatchdog() {
	go func() {
		interval := l.ttl / 3
		if interval < 1 {
			interval = 1 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-l.stopCh:
				return
			case <-ticker.C:
				res, err := l.client.Eval(context.Background(), refreshScript, []string{l.key}, l.token, l.ttl.Milliseconds()).Int64()
				if err != nil || res == 0 {
					// Lock lost or expired
					close(l.doneCh)
					return
				}
			}
		}
	}()
}

// Replace old methods with new ones
func (l *redisLock) Unlock(ctx context.Context) error {
	l.once.Do(func() {
		close(l.stopCh)
		res, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.token).Int64()
		if err != nil {
			l.err = err
		} else if res == 0 {
			l.err = dlock.ErrInvalidToken
		}
		close(l.doneCh)
	})
	return l.err
}

func (l *redisLock) Valid() bool {
	select {
	case <-l.doneCh:
		return false
	default:
		return true
	}
}

func (l *redisLock) Done() <-chan struct{} {
	return l.doneCh
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./dlock/redislock`
Expected: PASS

- [ ] **Step 3: Commit**

Run:
```bash
git add dlock/redislock/redislock.go
git commit -m "refactor(dlock): implement watchdog and new Lock interface for redislock"
```

---

### Task 4: Refactor mongolock

**Files:**
- Modify: `dlock/mongolock/mongolock.go`

- [ ] **Step 1: Implement watchdog and new interface**

In `dlock/mongolock/mongolock.go`, update the `mongoLock` struct and methods:
```go
// Add "sync" to imports

type mongoLock struct {
	coll   *mongo.Collection
	key    string
	token  string
	ttl    time.Duration
	doneCh chan struct{}
	stopCh chan struct{}
	once   sync.Once
	err    error
}

// Update TryLock to initialize channels and start watchdog
func (l *mongoLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	lockOpts := dlock.NewOptions(opts...)
	now := time.Now()
	expiresAt := now.Add(lockOpts.TTL)

	filter := bson.M{
		"_id": key,
		"expiresAt": bson.M{"$lte": now},
	}
	
	update := bson.M{
		"$set": bson.M{
			"token":     lockOpts.Token,
			"expiresAt": expiresAt,
		},
	}

	updateOpts := options.Update().SetUpsert(true)

	_, err := l.coll.UpdateOne(ctx, filter, update, updateOpts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	lock := &mongoLock{
		coll:   l.coll,
		key:    key,
		token:  lockOpts.Token,
		ttl:    lockOpts.TTL,
		doneCh: make(chan struct{}),
		stopCh: make(chan struct{}),
	}
	
	lock.startWatchdog()
	return lock, true, nil
}

func (l *mongoLock) startWatchdog() {
	go func() {
		interval := l.ttl / 3
		if interval < 1 {
			interval = 1 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-l.stopCh:
				return
			case <-ticker.C:
				filter := bson.M{
					"_id":   l.key,
					"token": l.token,
				}
				update := bson.M{
					"$set": bson.M{
						"expiresAt": time.Now().Add(l.ttl),
					},
				}
				res, err := l.coll.UpdateOne(context.Background(), filter, update)
				if err != nil || res.MatchedCount == 0 {
					// Lock lost or expired
					close(l.doneCh)
					return
				}
			}
		}
	}()
}

// Replace old methods with new ones
func (l *mongoLock) Unlock(ctx context.Context) error {
	l.once.Do(func() {
		close(l.stopCh)
		filter := bson.M{
			"_id":   l.key,
			"token": l.token,
		}
		res, err := l.coll.DeleteOne(ctx, filter)
		if err != nil {
			l.err = err
		} else if res.DeletedCount == 0 {
			l.err = dlock.ErrInvalidToken
		}
		close(l.doneCh)
	})
	return l.err
}

func (l *mongoLock) Valid() bool {
	select {
	case <-l.doneCh:
		return false
	default:
		return true
	}
}

func (l *mongoLock) Done() <-chan struct{} {
	return l.doneCh
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./dlock/mongolock`
Expected: PASS

- [ ] **Step 3: Commit**

Run:
```bash
git add dlock/mongolock/mongolock.go
git commit -m "refactor(dlock): implement watchdog and new Lock interface for mongolock"
```

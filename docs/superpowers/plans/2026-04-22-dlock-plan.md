# dlock Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a distributed lock package (`dlock`) with unified interfaces and Redis, etcd, and MongoDB backends.

**Architecture:** The core package defines `Locker`, `Lock`, and functional `Option` types. Each backend is implemented in its own subpackage (`redislock`, `etcdlock`, `mongolock`) that takes its specific client and common options.

**Tech Stack:** Go, `github.com/redis/go-redis/v9`, `go.etcd.io/etcd/client/v3`, `go.mongodb.org/mongo-driver/mongo`, `github.com/google/uuid`.

---

### Task 1: Setup Core Package and Options

**Files:**
- Create: `dlock/dlock.go`
- Create: `dlock/option.go`
- Create: `dlock/errors.go`
- Test: `dlock/option_test.go`

- [ ] **Step 1: Install google/uuid**

Run: `go get github.com/google/uuid`
Expected: PASS

- [ ] **Step 2: Create Core Interfaces and Errors**

Create `dlock/dlock.go` and `dlock/errors.go`.

`dlock/dlock.go`:
```go
package dlock

import (
	"context"
)

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
```

`dlock/errors.go`:
```go
package dlock

import "errors"

var (
	ErrLockFailed   = errors.New("failed to acquire lock")
	ErrInvalidToken = errors.New("invalid token or lock expired")
)
```

- [ ] **Step 3: Create Options**

Create `dlock/option.go`:
```go
package dlock

import (
	"time"

	"github.com/google/uuid"
)

type Options struct {
	TTL        time.Duration
	RetryDelay time.Duration
	Token      string
}

type Option func(*Options)

func WithTTL(d time.Duration) Option {
	return func(o *Options) {
		o.TTL = d
	}
}

func WithRetryDelay(d time.Duration) Option {
	return func(o *Options) {
		o.RetryDelay = d
	}
}

func WithToken(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}

func NewOptions(opts ...Option) *Options {
	o := &Options{
		TTL:        10 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		Token:      uuid.NewString(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
```

- [ ] **Step 4: Write and run Options test**

Create `dlock/option_test.go`:
```go
package dlock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	opts := NewOptions(WithTTL(5*time.Second), WithRetryDelay(100*time.Millisecond), WithToken("test-token"))
	assert.Equal(t, 5*time.Second, opts.TTL)
	assert.Equal(t, 100*time.Millisecond, opts.RetryDelay)
	assert.Equal(t, "test-token", opts.Token)

	defaultOpts := NewOptions()
	assert.Equal(t, 10*time.Second, defaultOpts.TTL)
	assert.Equal(t, 50*time.Millisecond, defaultOpts.RetryDelay)
	assert.NotEmpty(t, defaultOpts.Token)
}
```

Run: `go test ./dlock -v`
Expected: PASS

- [ ] **Step 5: Commit**

Run:
```bash
git add go.mod go.sum dlock/
git commit -m "feat: add dlock core interfaces and options"
```

---

### Task 2: Implement Redis Backend

**Files:**
- Create: `dlock/redislock/redislock.go`

- [ ] **Step 1: Install Redis dependency**

Run: `go get github.com/redis/go-redis/v9`
Expected: PASS

- [ ] **Step 2: Create Redis Locker Implementation**

Create `dlock/redislock/redislock.go`:
```go
package redislock

import (
	"context"
	"time"

	"github.com/8liang/kit/dlock"
	"github.com/redis/go-redis/v9"
)

const (
	unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`
	refreshScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end
`
)

type redisLocker struct {
	client *redis.Client
}

type redisLock struct {
	client *redis.Client
	key    string
	token  string
	ttl    time.Duration
}

func New(client *redis.Client) dlock.Locker {
	return &redisLocker{client: client}
}

func (l *redisLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	o := dlock.NewOptions(opts...)
	
	ok, err := l.client.SetNX(ctx, key, o.Token, o.TTL).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	return &redisLock{
		client: l.client,
		key:    key,
		token:  o.Token,
		ttl:    o.TTL,
	}, true, nil
}

func (l *redisLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	o := dlock.NewOptions(opts...)
	
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			lock, ok, err := l.TryLock(ctx, key, opts...)
			if err != nil {
				return nil, err
			}
			if ok {
				return lock, nil
			}
			timer.Reset(o.RetryDelay)
		}
	}
}

func (l *redisLock) Key() string {
	return l.key
}

func (l *redisLock) Token() string {
	return l.token
}

func (l *redisLock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.token).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return dlock.ErrInvalidToken
	}
	return nil
}

func (l *redisLock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, refreshScript, []string{l.key}, l.token, l.ttl.Milliseconds()).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return dlock.ErrInvalidToken
	}
	return nil
}
```

- [ ] **Step 3: Run build**

Run: `go build ./dlock/redislock`
Expected: PASS

- [ ] **Step 4: Commit**

Run:
```bash
git add go.mod go.sum dlock/redislock/
git commit -m "feat: implement redis distributed lock backend"
```

---

### Task 3: Implement etcd Backend

**Files:**
- Create: `dlock/etcdlock/etcdlock.go`

- [ ] **Step 1: Install etcd dependency**

Run: `go get go.etcd.io/etcd/client/v3`
Expected: PASS

- [ ] **Step 2: Create etcd Locker Implementation**

Create `dlock/etcdlock/etcdlock.go`:
```go
package etcdlock

import (
	"context"
	"time"

	"github.com/8liang/kit/dlock"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdLocker struct {
	client *clientv3.Client
}

type etcdLock struct {
	client  *clientv3.Client
	key     string
	token   string
	leaseID clientv3.LeaseID
}

func New(client *clientv3.Client) dlock.Locker {
	return &etcdLocker{client: client}
}

func (l *etcdLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	o := dlock.NewOptions(opts...)
	
	// Grant lease
	ttlSeconds := int64(o.TTL.Seconds())
	if ttlSeconds == 0 {
		ttlSeconds = 1
	}
	lease, err := l.client.Grant(ctx, ttlSeconds)
	if err != nil {
		return nil, false, err
	}

	// Try to acquire lock
	cmp := clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
	put := clientv3.OpPut(key, o.Token, clientv3.WithLease(lease.ID))
	
	txn, err := l.client.Txn(ctx).If(cmp).Then(put).Commit()
	if err != nil {
		l.client.Revoke(context.Background(), lease.ID)
		return nil, false, err
	}

	if !txn.Succeeded {
		l.client.Revoke(context.Background(), lease.ID)
		return nil, false, nil
	}

	return &etcdLock{
		client:  l.client,
		key:     key,
		token:   o.Token,
		leaseID: lease.ID,
	}, true, nil
}

func (l *etcdLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	o := dlock.NewOptions(opts...)
	
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			lock, ok, err := l.TryLock(ctx, key, opts...)
			if err != nil {
				return nil, err
			}
			if ok {
				return lock, nil
			}
			timer.Reset(o.RetryDelay)
		}
	}
}

func (l *etcdLock) Key() string {
	return l.key
}

func (l *etcdLock) Token() string {
	return l.token
}

func (l *etcdLock) Unlock(ctx context.Context) error {
	cmp := clientv3.Compare(clientv3.Value(l.key), "=", l.token)
	del := clientv3.OpDelete(l.key)
	
	txn, err := l.client.Txn(ctx).If(cmp).Then(del).Commit()
	if err != nil {
		return err
	}
	
	if !txn.Succeeded {
		return dlock.ErrInvalidToken
	}
	
	// Revoke lease in background
	_, _ = l.client.Revoke(context.Background(), l.leaseID)
	return nil
}

func (l *etcdLock) Refresh(ctx context.Context) error {
	_, err := l.client.KeepAliveOnce(ctx, l.leaseID)
	if err != nil {
		return dlock.ErrInvalidToken
	}
	return nil
}
```

- [ ] **Step 3: Run build**

Run: `go build ./dlock/etcdlock`
Expected: PASS

- [ ] **Step 4: Commit**

Run:
```bash
git add go.mod go.sum dlock/etcdlock/
git commit -m "feat: implement etcd distributed lock backend"
```

---

### Task 4: Implement MongoDB Backend

**Files:**
- Create: `dlock/mongolock/mongolock.go`

- [ ] **Step 1: Install mongo dependency**

Run: `go get go.mongodb.org/mongo-driver/mongo`
Expected: PASS

- [ ] **Step 2: Create MongoDB Locker Implementation**

Create `dlock/mongolock/mongolock.go`:
```go
package mongolock

import (
	"context"
	"time"

	"github.com/8liang/kit/dlock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoLocker struct {
	coll *mongo.Collection
}

type mongoLock struct {
	coll  *mongo.Collection
	key   string
	token string
	ttl   time.Duration
}

type lockDoc struct {
	ID        string    `bson:"_id"`
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

func New(collection *mongo.Collection) dlock.Locker {
	return &mongoLocker{coll: collection}
}

func (l *mongoLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	o := dlock.NewOptions(opts...)
	now := time.Now()
	expiresAt := now.Add(o.TTL)

	filter := bson.M{
		"_id": key,
		"$or": []bson.M{
			{"expiresAt": bson.M{"$lte": now}},
			{"_id": bson.M{"$exists": false}},
		},
	}
	
	update := bson.M{
		"$set": bson.M{
			"token":     o.Token,
			"expiresAt": expiresAt,
		},
	}

	upsert := true
	opt := &options.FindOneAndUpdateOptions{
		Upsert: &upsert,
	}

	err := l.coll.FindOneAndUpdate(ctx, filter, update, opt).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments || mongo.IsDuplicateKeyError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &mongoLock{
		coll:  l.coll,
		key:   key,
		token: o.Token,
		ttl:   o.TTL,
	}, true, nil
}

func (l *mongoLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	o := dlock.NewOptions(opts...)
	
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			lock, ok, err := l.TryLock(ctx, key, opts...)
			if err != nil {
				return nil, err
			}
			if ok {
				return lock, nil
			}
			timer.Reset(o.RetryDelay)
		}
	}
}

func (l *mongoLock) Key() string {
	return l.key
}

func (l *mongoLock) Token() string {
	return l.token
}

func (l *mongoLock) Unlock(ctx context.Context) error {
	filter := bson.M{
		"_id":   l.key,
		"token": l.token,
	}

	res, err := l.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return dlock.ErrInvalidToken
	}
	return nil
}

func (l *mongoLock) Refresh(ctx context.Context) error {
	filter := bson.M{
		"_id":   l.key,
		"token": l.token,
	}
	
	update := bson.M{
		"$set": bson.M{
			"expiresAt": time.Now().Add(l.ttl),
		},
	}

	res, err := l.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return dlock.ErrInvalidToken
	}
	return nil
}
```

- [ ] **Step 3: Run build**

Run: `go build ./dlock/mongolock`
Expected: PASS

- [ ] **Step 4: Commit**

Run:
```bash
git add go.mod go.sum dlock/mongolock/
git commit -m "feat: implement mongodb distributed lock backend"
```

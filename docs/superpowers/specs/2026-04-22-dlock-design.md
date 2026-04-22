# Distributed Lock (`dlock`) Design Spec

## Overview
A generic distributed lock library (`dlock`) for Go, providing a unified interface for multiple backend storage solutions: Redis, etcd, and MongoDB. The package leverages the Functional Options pattern to configure generic behavior across all backends while allowing clean separation of dependencies.

## Architecture

### 1. Core Interfaces
Located in `dlock/dlock.go` and `dlock/option.go`:

```go
package dlock

import (
	"context"
	"time"
)

// Locker is the unified interface for distributed locks.
type Locker interface {
	Lock(ctx context.Context, key string, opts ...Option) (Lock, error)
	TryLock(ctx context.Context, key string, opts ...Option) (Lock, bool, error)
}

// Lock represents an acquired distributed lock.
type Lock interface {
	Key() string
	Token() string
	Unlock(ctx context.Context) error
	Refresh(ctx context.Context) error
}

// Option configures lock behavior.
type Option func(*Options)

// Options holds configuration for the lock.
type Options struct {
	TTL        time.Duration
	RetryDelay time.Duration
	Token      string
}
```

### 2. Functional Options
Located in `dlock/option.go`:
- `WithTTL(d time.Duration)`: Sets the lock expiration time (default: 10s).
- `WithRetryDelay(d time.Duration)`: Delay between retry attempts for `Lock` (default: 50ms).
- `WithToken(token string)`: Custom unique identifier for the lock (default: auto-generated UUID).

### 3. Implementations (Subpackages)
By placing implementations in subpackages, users only import the specific driver they need, preventing bloated dependencies.

#### Redis (`dlock/redislock`)
- **Dependencies**: `github.com/redis/go-redis/v9`
- **Constructor**: `func New(client *redis.Client) dlock.Locker`
- **Implementation**:
  - `TryLock`: Uses `SET <key> <token> NX PX <ttl>`.
  - `Unlock`: Uses a Lua script to check the token and delete the key atomically.
  - `Refresh`: Uses a Lua script to check the token and PEXPIRE the key atomically.

#### etcd (`dlock/etcdlock`)
- **Dependencies**: `go.etcd.io/etcd/client/v3`
- **Constructor**: `func New(client *clientv3.Client) dlock.Locker`
- **Implementation**:
  - Leverages etcd Leases for TTL and transactions (`Txn`) to compare revision and put the lock atomically.
  - `TryLock`: Creates a lease with the TTL, uses a transaction to ensure `create_revision == 0` for the key, and puts the token.
  - `Unlock`: Compares the token and deletes the key, then revokes the lease.
  - `Refresh`: Uses `KeepAliveOnce` on the lease to extend the TTL.

#### MongoDB (`dlock/mongolock`)
- **Dependencies**: `go.mongodb.org/mongo-driver/mongo`
- **Constructor**: `func New(collection *mongo.Collection) dlock.Locker`
- **Implementation**:
  - `TryLock`: Uses `FindOneAndUpdate` with `upsert: true` and a query condition that either the key doesn't exist or its expiration time has passed.
  - The document structure will be: `{ "_id": key, "token": token, "expiresAt": time.Time }`.
  - `Unlock`: Uses `DeleteOne` matching both `_id` and `token`.
  - `Refresh`: Uses `UpdateOne` matching `_id` and `token` to update `expiresAt`.

### 4. Behavior Details
- **Lock**: Enters a loop attempting `TryLock`. If it fails, it waits for `RetryDelay` (using a `time.Timer` or `select`) before trying again. It respects `ctx.Done()` for timeout or cancellation.
- **TryLock**: Attempts to acquire the lock exactly once. Returns `(Lock, true, nil)` on success, `(nil, false, nil)` if the lock is held by someone else, or an error if a backend issue occurs.
- **Unlock**: Guarantees safety by checking if the token matches the lock's token before deleting.
- **Refresh**: Extends the lock's TTL if the token matches, resetting the expiration to the configured TTL.

## Error Handling
- Standardized error returns for common scenarios (e.g., `ErrLockFailed`, `ErrInvalidToken`).
- Context cancellation errors (`context.Canceled`, `context.DeadlineExceeded`) are propagated correctly.

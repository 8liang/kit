# dlock - Distributed Lock Package

The `dlock` package provides a unified and robust interface for managing distributed locks in Go. It abstracts away the complexity of different backends, allowing you to easily switch between Redis, etcd, and MongoDB.

## Features
- **Unified Interface**: Use the same `Locker` and `Lock` interface regardless of the underlying storage.
- **Context Aware**: Fully respects `context.Context` for cancellation and timeouts.
- **Auto-Retry**: The `Lock` method automatically retries acquiring the lock until it succeeds or the context expires.
- **Auto-Renewal (Watchdog)**: Automatically extends the lifetime of an acquired lock in the background while it is held.
- **Lock Loss Awareness**: Provides a `Done()` channel and `Valid()` method to immediately detect if the lock is lost due to network partitions or backend failures.
- **Safe & Idempotent Unlocking**: Unlocking is strictly idempotent and safely cleans up background watchdog goroutines or sessions.

## Installation & Import

Import the core `dlock` package and your desired backend implementation:

```go
import (
    "github.com/8liang/kit/dlock"
    
    // Import the backend you need:
    "github.com/8liang/kit/dlock/redislock"
    // "github.com/8liang/kit/dlock/etcdlock"
    // "github.com/8liang/kit/dlock/mongolock"
)
```

---

## Core Usage Guide

The `Locker` interface provides two main ways to acquire a lock:
1. `TryLock`: Non-blocking. Returns immediately if the lock is held by someone else.
2. `Lock`: Blocking. Retries automatically with a configurable delay until successful or the context is canceled.

### 1. Blocking Lock (`Lock`)

```go
ctx := context.Background()
key := "resource-key"

// Will block and retry every 100ms until acquired
lock, err := locker.Lock(ctx, key, 
    dlock.WithTTL(10*time.Second), 
    dlock.WithRetryDelay(100*time.Millisecond),
)
if err != nil {
    // Handle context cancellation or backend error
    return err
}

// Ensure the lock is released when done
defer lock.Unlock(ctx)

// Do your critical section work here...
// You can use a select block to ensure you abort if the lock is lost:
select {
case <-lock.Done():
    // The lock was lost (e.g. etcd session expired, Redis disconnected)
    return fmt.Errorf("lock lost during processing")
default:
    // Safe to process
}
```

### 2. Non-blocking Lock (`TryLock`)

```go
ctx := context.Background()
key := "resource-key"

lock, ok, err := locker.TryLock(ctx, key, dlock.WithTTL(5*time.Second))
if err != nil {
    return err // Backend error
}
if !ok {
    return fmt.Errorf("lock is already held by another process")
}
defer lock.Unlock(ctx)

// Fast non-blocking check
if !lock.Valid() {
    return fmt.Errorf("lock became invalid")
}

// Do your critical section work here...
```

---

## Supported Backends

### 1. Redis Backend

The Redis implementation relies on the `SETNX` command and Lua scripts for atomic operations, accompanied by a background watchdog goroutine for auto-renewal.

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/8liang/kit/dlock/redislock"
)

// Initialize redis client
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Create locker
var locker dlock.Locker = redislock.New(client)
```

### 2. etcd Backend

The etcd implementation utilizes `concurrency.Session` and `concurrency.Mutex`, which natively support auto-renewal and highly efficient watch-based lock queueing.

```go
import (
    clientv3 "go.etcd.io/etcd/client/v3"
    "github.com/8liang/kit/dlock/etcdlock"
)

// Initialize etcd client
client, err := clientv3.New(clientv3.Config{
    Endpoints: []string{"localhost:2379"},
})

// Create locker
var locker dlock.Locker = etcdlock.New(client)
```

### 3. MongoDB Backend

The MongoDB backend uses the `_id` field for uniqueness and an upsert pattern, accompanied by a background watchdog goroutine for auto-renewal.

**Important:** To prevent expired locks from accumulating and wasting space in your database, it is highly recommended to create a TTL index on the `expiresAt` field on your lock collection:

```javascript
// Run this in your mongo shell
db.distributed_locks.createIndex({ "expiresAt": 1 }, { expireAfterSeconds: 0 })
```

```go
import (
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/8liang/kit/dlock/mongolock"
)

// Assuming `db` is your *mongo.Database instance
coll := db.Collection("distributed_locks")

// Create locker
var locker dlock.Locker = mongolock.New(coll)
```
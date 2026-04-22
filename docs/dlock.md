# dlock - Distributed Lock Package

The `dlock` package provides a unified and robust interface for managing distributed locks in Go. It abstracts away the complexity of different backends, allowing you to easily switch between Redis, etcd, and MongoDB.

## Features
- **Unified Interface**: Use the same `Locker` and `Lock` interface regardless of the underlying storage.
- **Context Aware**: Fully respects `context.Context` for cancellation and timeouts.
- **Auto-Retry**: The `Lock` method automatically retries acquiring the lock until it succeeds or the context expires.
- **Refreshable TTL**: Extend the lifetime of an acquired lock easily via the `Refresh` method.
- **Safe Unlocking**: Uses unique tokens per lock to guarantee that a lock can only be released or refreshed by its owner.

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

// Do your critical section work here...
```

### 3. Refreshing a Lock

If your task takes longer than expected, you can refresh the lock's TTL:

```go
// Reset the TTL back to the original duration (e.g., 5 seconds)
err := lock.Refresh(ctx)
if err != nil {
    // Handle error (e.g., lock already expired or invalid token)
}
```

---

## Supported Backends

### 1. Redis Backend

The Redis implementation relies on the `SETNX` command and Lua scripts for atomic operations.

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

The etcd implementation utilizes etcd's native leases and transactional (Txn) capabilities.

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

The MongoDB backend uses the `_id` field for uniqueness and an upsert pattern. 

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

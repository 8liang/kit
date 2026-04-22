package mongolock

import (
	"context"
	"sync"
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
	coll   *mongo.Collection
	key    string
	token  string
	ttl    time.Duration
	doneCh chan struct{}
	stopCh chan struct{}
	once   sync.Once
	err    error
}

type lockDoc struct {
	ID        string    `bson:"_id"`
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

// New creates a MongoDB-backed Locker.
// It relies on MongoDB's unique index on `_id` and an upsert pattern.
// Note: To prevent expired locks from accumulating in the database,
// it is highly recommended to create a TTL index on the "expiresAt" field.
// Example: db.collection.createIndex({ "expiresAt": 1 }, { expireAfterSeconds: 0 })
func New(collection *mongo.Collection) dlock.Locker {
	return &mongoLocker{coll: collection}
}

func (l *mongoLocker) TryLock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, bool, error) {
	lockOpts := dlock.NewOptions(opts...)
	now := time.Now()
	expiresAt := now.Add(lockOpts.TTL)

	filter := bson.M{
		"_id":       key,
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

func (l *mongoLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	lockOpts := dlock.NewOptions(opts...)
	retryOpts := append(opts, dlock.WithToken(lockOpts.Token))

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			lock, ok, err := l.TryLock(ctx, key, retryOpts...)
			if err != nil {
				return nil, err
			}
			if ok {
				return lock, nil
			}
			timer.Reset(lockOpts.RetryDelay)
		}
	}
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

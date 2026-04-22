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
	retryOpts := append(opts, dlock.WithToken(o.Token))
	
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
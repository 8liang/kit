package redislock

import (
	"context"
	"sync"
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
	doneCh chan struct{}
	stopCh chan struct{}
	once   sync.Once
	err    error
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

func (l *redisLocker) Lock(ctx context.Context, key string, opts ...dlock.Option) (dlock.Lock, error) {
	o := dlock.NewOptions(opts...)

	timer := time.NewTimer(0)
	defer timer.Stop()

	// Append the generated token to options so we don't generate a new UUID on every retry
	retryOpts := append(opts, dlock.WithToken(o.Token))

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

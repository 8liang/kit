package etcdlock

import (
	"context"
	"fmt"

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
	return l.token
}

func (l *etcdLock) Unlock(ctx context.Context) error {
	defer func() {
		_ = l.session.Close()
	}()

	err := l.mutex.Unlock(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *etcdLock) Refresh(ctx context.Context) error {
	select {
	case <-l.session.Done():
		return fmt.Errorf("%w: etcd session expired or closed", dlock.ErrInvalidToken)
	default:
		return nil
	}
}

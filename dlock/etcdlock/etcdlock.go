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

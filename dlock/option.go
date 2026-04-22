package dlock

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultTTL        = 10 * time.Second
	DefaultRetryDelay = 50 * time.Millisecond
)

type Options struct {
	TTL        time.Duration
	RetryDelay time.Duration
	Token      string
}

type Option func(*Options)

func WithTTL(d time.Duration) Option {
	return func(o *Options) { o.TTL = d }
}

func WithRetryDelay(d time.Duration) Option {
	return func(o *Options) { o.RetryDelay = d }
}

func WithToken(token string) Option {
	return func(o *Options) { o.Token = token }
}

func NewOptions(opts ...Option) *Options {
	o := &Options{
		TTL:        DefaultTTL,
		RetryDelay: DefaultRetryDelay,
		Token:      uuid.NewString(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

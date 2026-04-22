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

package common

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/sethvargo/go-retry"
)

func TestExecuteRetryDisabledByDefault(t *testing.T) {
	r := &Retry{}
	var calls atomic.Int64
	err := r.ExecuteRetry(t.Context(), func(ctx context.Context) error {
		calls.Add(1)
		return assertError{}
	})
	assert.Error(t, err)
	assert.Equal(t, int64(1), calls.Load())
}

func TestExecuteRetryEnabledWhenRetriesSet(t *testing.T) {
	r := &Retry{}
	r.SetMaxRetry(2)
	r.SetRetryDelay(time.Millisecond)
	var calls atomic.Int64
	err := r.ExecuteRetry(t.Context(), func(ctx context.Context) error {
		calls.Add(1)
		return retry.RetryableError(assertError{})
	})
	assert.Error(t, err)
	assert.Equal(t, int64(3), calls.Load())
}

type assertError struct{}

func (assertError) Error() string { return "boom" }

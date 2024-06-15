package common

import (
	"context"
	"testing"
	"time"

	"github.com/FMotalleb/crontab-go/ctxutils"
)

func TestSetMaxRetry(t *testing.T) {
	r := &Retry{}
	r.SetMaxRetry(5)
	if r.maxRetries != 5 {
		t.Errorf("expected maxRetries to be 5, got %d", r.maxRetries)
	}
}

func TestSetRetryDelay(t *testing.T) {
	r := &Retry{}
	delay := 2 * time.Second
	r.SetRetryDelay(delay)
	if r.retryDelay != delay {
		t.Errorf("expected retryDelay to be %v, got %v", delay, r.retryDelay)
	}
}

func TestWaitForRetry(t *testing.T) {
	r := &Retry{maxRetries: 3, retryDelay: 1 * time.Second}
	ctx := context.WithValue(context.Background(), ctxutils.RetryCountKey, uint(2))

	start := time.Now()
	err := r.WaitForRetry(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if elapsed < 2*time.Second {
		t.Errorf("expected at least 2 seconds delay, got %v", elapsed)
	}
}

func TestWaitForRetryMaxExceeded(t *testing.T) {
	r := &Retry{maxRetries: 3, retryDelay: 1 * time.Second}
	ctx := context.WithValue(context.Background(), ctxutils.RetryCountKey, uint(5))

	err := r.WaitForRetry(ctx)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestIncreaseRetry(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxutils.RetryCountKey, uint(2))
	newCtx := IncreaseRetry(ctx)
	if GetRetry(newCtx) != 3 {
		t.Errorf("expected retry count to be 3, got %d", GetRetry(newCtx))
	}
}

func TestZeroValueRetry(t *testing.T) {
	ctx := context.Background()

	if GetRetry(ctx) != 0 {
		t.Errorf("expected retry count to be 3, got %d", GetRetry(ctx))
	}
}

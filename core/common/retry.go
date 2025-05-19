// Package common provides implementation of some of the basic functionalities to be used in application.
package common

import (
	"context"
	"fmt"
	"time"

	"github.com/FMotalleb/crontab-go/ctxutils"
)

type Retry struct {
	maxRetries int64
	retryDelay time.Duration
}

func (r *Retry) SetMaxRetry(retries int64) {
	r.maxRetries = retries
}

func (r *Retry) SetRetryDelay(retryDelay time.Duration) {
	r.retryDelay = retryDelay
}

func (r *Retry) WaitForRetry(ctx context.Context) error {
	tries := GetRetry(ctx)
	if tries > (r.maxRetries) {
		return fmt.Errorf("max retry of %d exceeded, tries: %d", r.maxRetries, tries)
	}
	if tries != 0 {
		time.Sleep(time.Duration(tries) * r.retryDelay)
	}
	return nil
}

func GetRetry(ctx context.Context) int64 {
	if result, ok := ctx.Value(ctxutils.RetryCountKey).(int64); ok {
		return result
	}
	return 0
}

func IncreaseRetry(ctx context.Context) context.Context {
	current := GetRetry(ctx)
	return context.WithValue(ctx, ctxutils.RetryCountKey, current+1)
}

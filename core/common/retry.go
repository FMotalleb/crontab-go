// Package common provides implementation of some of the basic functionalities to be used in application.
package common

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sethvargo/go-retry"

	"github.com/fmotalleb/crontab-go/config"
)

var retryTracer = otel.Tracer("crontab-go/retry")

type (
	RetryCount    = uint64
	DelayModifier string
)

const (
	RetryConstant    = DelayModifier("cons")
	RetryExponential = DelayModifier("expo")
	RetryFibonacci   = DelayModifier("fibo")
)

type Retry struct {
	maxRetries    RetryCount
	maxTimeout    time.Duration
	retryDelay    time.Duration
	maxDelay      time.Duration
	jitter        time.Duration
	delayModifier DelayModifier
}

func (r *Retry) SetMaxRetry(retries uint64) {
	r.maxRetries = retries
}

func (r *Retry) SetRetryDelay(retryDelay time.Duration) {
	r.retryDelay = retryDelay
}

func (r *Retry) SetMaxTimeout(d time.Duration) {
	r.maxTimeout = d
}

func (r *Retry) SetMaxDelay(d time.Duration) {
	r.maxDelay = d
}

func (r *Retry) SetJitter(d time.Duration) {
	r.jitter = d
}

func (r *Retry) SetDelayModifierFromString(s string) {
	s = strings.ToLower(s)
	switch s {
	case "const", "cons", "constant":
		r.delayModifier = RetryConstant
	case "expo", "exponential":
		r.delayModifier = RetryExponential
	case "fibo", "fibonacci":
		r.delayModifier = RetryFibonacci
	default:
		r.delayModifier = RetryExponential
	}
}

func (r *Retry) ConfigRetryFrom(t *config.Task) {
	r.SetMaxRetry(t.Retries)
	r.SetRetryDelay(t.RetryDelay)
	r.SetMaxTimeout(t.RetryTimeout)
	r.SetMaxDelay(t.RetryMaxDelay)
	r.SetJitter(t.RetryJitter)
	r.SetDelayModifierFromString(t.RetryModifier)
}

func (r *Retry) ExecuteRetry(ctx context.Context, fn func(context.Context) error) error {
	if r.maxRetries == 0 {
		if cause := context.Cause(ctx); cause != nil {
			return cause
		}
		return fn(ctx)
	}
	ctx, span := retryTracer.Start(ctx, "task.retry",
		trace.WithAttributes(
			attribute.Int64("retry.max_retries", int64(r.maxRetries)),
			attribute.Int64("retry.delay_nanos", int64(r.retryDelay)),
			attribute.String("retry.mode", string(r.delayModifier)),
		),
	)
	defer span.End()
	var attempt atomic.Uint64
	countedFn := func(ctx context.Context) error {
		n := attempt.Add(1)
		span.SetAttributes(attribute.Int64("retry.attempt", int64(n)))
		return fn(ctx)
	}
	err := retry.Do(ctx, r.buildBackoff(), countedFn)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	return err
}

func (r *Retry) buildBackoff() retry.Backoff {
	retryDelay := r.retryDelay
	if retryDelay == 0 {
		retryDelay = time.Second
	}
	var backoff retry.Backoff
	switch r.delayModifier {
	case RetryConstant:
		backoff = retry.NewConstant(retryDelay)
	case RetryExponential:
		backoff = retry.NewExponential(retryDelay)
	case RetryFibonacci:
		backoff = retry.NewFibonacci(retryDelay)
	default:
		backoff = retry.NewExponential(retryDelay)
	}
	if r.maxDelay != 0 {
		backoff = retry.WithCappedDuration(r.maxDelay, backoff)
	}
	if r.maxTimeout != 0 {
		backoff = retry.WithMaxDuration(r.maxTimeout, backoff)
	}
	backoff = retry.WithMaxRetries(r.maxRetries, backoff)
	if r.jitter != 0 {
		backoff = retry.WithJitter(r.jitter, backoff)
	}
	return backoff
}

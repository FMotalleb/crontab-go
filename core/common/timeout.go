package common

import (
	"context"
	"time"
)

type Timeout struct {
	timeout time.Duration
}

func (t *Timeout) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

func (t *Timeout) ApplyTimeout(ctx context.Context) (context.Context, func()) {
	if t.timeout != 0 {
		return context.WithTimeout(ctx, t.timeout)
	}
	return context.WithCancel(ctx)
}

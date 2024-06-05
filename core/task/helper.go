package task

import (
	"context"

	"github.com/FMotalleb/crontab-go/ctxutils"
)

func getRetry(ctx context.Context) uint {
	if result, ok := ctx.Value(ctxutils.RetryCount).(uint); ok {
		return result
	}
	return 0
}

func increaseRetry(ctx context.Context) context.Context {
	current := getRetry(ctx)

	return context.WithValue(ctx, ctxutils.RetryCount, current+1)
}

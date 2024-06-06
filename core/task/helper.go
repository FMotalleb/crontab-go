package task

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/ctxutils"
)

func getRetry(ctx context.Context) uint {
	if result, ok := ctx.Value(ctxutils.RetryCountKey).(uint); ok {
		return result
	}
	return 0
}

func increaseRetry(ctx context.Context) context.Context {
	current := getRetry(ctx)

	return context.WithValue(ctx, ctxutils.RetryCountKey, current+1)
}

func logResponse(r *http.Response) logrus.LogFunction {
	return func() []any {
		result := &ResponseWriter{}
		err := r.Write(result)
		return []any{
			fmt.Sprintf("error: %v", err),
			"\n",
			result.String(),
		}
	}
}

type ResponseWriter struct {
	buffer []byte
}

// Write implements io.Writer.
func (r *ResponseWriter) Write(p []byte) (n int, err error) {
	initial := len(r.buffer)
	r.buffer = append(r.buffer, p...)
	return len(r.buffer) - initial, nil
}

func (r *ResponseWriter) String() string {
	return string(r.buffer)
}

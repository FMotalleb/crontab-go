package task

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func logHTTPResponse(r *http.Response) logrus.LogFunction {
	return func() []any {
		result := bytes.NewBuffer([]byte{})
		err := r.Write(result)
		return []any{
			fmt.Sprintf("error: %v", err),
			"\n",
			result.String(),
		}
	}
}

func getFailedConnections(ctx context.Context) []config.TaskConnection {
	items := ctx.Value(ctxutils.FailedRemotes)
	if items != nil {
		return items.([]config.TaskConnection)
	}
	return []config.TaskConnection{}
}

func addFailedConnections(ctx context.Context, con config.TaskConnection) context.Context {
	current := getFailedConnections(ctx)
	return context.WithValue(ctx, ctxutils.FailedRemotes, append(current, con))
}

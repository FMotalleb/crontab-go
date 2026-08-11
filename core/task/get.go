package task

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/common"
)

func init() {
	tg.Register(NewGet)
}

func NewGet(logger *zap.Logger, task *config.Task) (abstraction.Executable, bool) {
	if task.Get == "" {
		return nil, false
	}
	get := &Get{
		address: task.Get,
		task:    task,
		headers: &task.Headers,
		log: logger.With(
			zap.String("url", task.Get),
			zap.String("method", "get"),
		),
	}
	get.ConfigRetryFrom(task)
	get.SetTimeout(task.Timeout)
	get.SetMetaName("get: " + task.Get)
	get.Action = get
	return get, true
}

type Get struct {
	common.Executable
	common.Cancelable
	common.Timeout
	task    *config.Task
	address string
	headers *map[string]string
	log     *zap.Logger
}

// Do implements common.Action.
func (g *Get) Do(ctx context.Context) (e error) {
	ctx = populateVars(ctx, g.task)
	_, span := taskTracer.Start(ctx, "http.get",
		trace.WithAttributes(
			attribute.String("url.full", g.address),
			attribute.String("http.request.method", "GET"),
		),
	)
	defer span.End()
	log := g.log.With(
		zap.Time("start", time.Now()),
	)
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error("panic recovered", zap.Error(err))
				e = err
			} else {
				log.Error("panic recovered", zap.Any("error", r))
				e = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	localCtx, cancel := g.ApplyTimeout(ctx)
	defer cancel()
	g.SetCancel(cancel)

	req, err := http.NewRequestWithContext(localCtx, http.MethodGet, g.address, nil)
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return err
	}
	prefix := outputPrefix(jobName(ctx), g.address)
	return doHTTP(newHTTPClient(g.task.Insecure), req, g.headers, NewPrefixWriter(os.Stdout, prefix), log)
}

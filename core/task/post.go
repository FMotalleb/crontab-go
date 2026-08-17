package task

import (
	"bytes"
	"context"
	"encoding/json"
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
	tg.Register(NewPost)
}

func NewPost(logger *zap.Logger, task *config.Task) (abstraction.Executable, bool) {
	if task.Post == "" {
		return nil, false
	}
	post := &Post{
		address: task.Post,
		headers: &task.Headers,
		data:    &task.Data,
		task:    task,
		log: logger.With(
			zap.String("url", task.Post),
			zap.String("method", "post"),
		),
	}
	post.ConfigRetryFrom(task)
	post.SetTimeout(task.Timeout)
	post.SetMetaName("post: " + task.Post)
	post.Action = post
	return post, true
}

type Post struct {
	common.Executable
	common.Cancelable
	common.Timeout
	task *config.Task

	address string
	headers *map[string]string
	data    *any
	log     *zap.Logger
}

// Do implements common.Action.
func (p *Post) Do(ctx context.Context) (e error) {
	execID := newExecutionID()
	ctx = populateVars(ctx, p.task)
	_, span := taskTracer.Start(ctx, "http.post",
		trace.WithAttributes(
			attribute.String("url.full", p.address),
			attribute.String("http.request.method", "POST"),
			attribute.String("execution.id", execID),
		),
	)
	defer span.End()
	log := p.log.With(
		zap.String("id", execID),
		zap.Time("start", time.Now()),
	)
	log.Info("post started")
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

	localCtx, cancel := p.ApplyTimeout(ctx)
	defer cancel()
	p.SetCancel(cancel)

	var dataReader *bytes.Reader
	if p.data != nil {
		data, err := json.Marshal(p.data)
		if err != nil {
			log.Warn("cannot marshal the given body (pre-send)", zap.Error(err))
			return err
		}
		dataReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(localCtx, http.MethodPost, p.address, dataReader)
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return err
	}
	return doHTTP(newHTTPClient(p.task.Insecure), req, p.headers, NewPrefixWriter(os.Stderr, executionPrefix(execID)), log)
}

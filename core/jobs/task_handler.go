package jobs

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/global"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

var tracer = otel.Tracer("crontab-go/task_handler")

func taskHandler(
	logger *zap.Logger,
	ed abstraction.EventDispatcher,
	tasks []abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
	jobName string,
) {
	logger.Debug("Spawning task handler")
	ed.AddListener(func(ctx context.Context, e abstraction.Event) {
		logger.Debug("Signal Received")
		for _, task := range tasks {
			ctxInternal := context.WithValue(ctx, ctxutils.EventData, e)
			go executeTask(ctxInternal, task, doneHooks, failHooks, lock, jobName)
		}
	})
}

func executeTask(
	c context.Context,
	task abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
	jobName string,
) {
	defer func() {
		if r := recover(); r != nil {
			global.Logger("task_handler").Error("task panicked", zap.Any("recover", r))
		}
	}()

	ctx, span := tracer.Start(c, "job."+jobName+"/task.execute")
	defer span.End()

	lock.Lock()
	defer lock.Unlock()
	ctx = context.WithValue(ctx, ctxutils.TaskKey, task)
	ctx = context.WithValue(ctx, ctxutils.JobKey, jobName)
	err := task.Execute(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		for _, hook := range failHooks {
			_ = hook.Execute(ctx)
		}
	} else {
		span.SetStatus(codes.Ok, "")
		for _, hook := range doneHooks {
			_ = hook.Execute(ctx)
		}
	}
}

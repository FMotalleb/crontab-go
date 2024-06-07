package jobs

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func taskHandler(
	c context.Context,
	logger *logrus.Entry,
	signal <-chan any,
	tasks []abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
) {
	logger.Debug("Spawning task handler")
	for range signal {
		logger.Trace("Signal Received")
		for _, task := range tasks {
			go executeTask(c, task, doneHooks, failHooks, lock)
		}
	}
}

func executeTask(
	c context.Context,
	task abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
) {
	lock.Lock()
	defer lock.Unlock()
	ctx := context.WithValue(c, ctxutils.TaskKey, task)
	err := task.Execute(ctx)
	switch err {
	case nil:
		for _, task := range doneHooks {
			_ = task.Execute(ctx)
		}
	default:
		for _, task := range failHooks {
			_ = task.Execute(ctx)
		}
	}
}

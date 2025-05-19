// Package cfgcompiler is deprecated and will be removed in future releases.
package cfgcompiler

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/task"
)

func CompileTask(ctx context.Context, t *config.Task, logger *logrus.Entry) abstraction.Executable {
	var exe abstraction.Executable
	switch {
	case t.Command != "":
		exe = task.NewCommand(t, logger)
	case t.Get != "":
		exe = task.NewGet(t, logger)
	case t.Post != "":
		exe = task.NewPost(t, logger)
	default:
		logger.Panic("cannot handle given task config", t)
	}
	if exe == nil {
		logger.Panic("did not received any executable action from given task", t)
	}
	onDone := []abstraction.Executable{}
	for _, d := range t.OnDone {
		onDone = append(onDone, CompileTask(ctx, &d, logger))
	}
	exe.SetDoneHooks(ctx, onDone)
	onFail := []abstraction.Executable{}
	for _, d := range t.OnFail {
		onFail = append(onFail, CompileTask(ctx, &d, logger))
	}
	exe.SetFailHooks(ctx, onFail)
	return exe
}

package cfgcompiler

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/task"
)

func CompileTask(t *config.Task, logger *logrus.Entry) abstraction.Executable {
	var exe abstraction.Executable
	switch {
	case t.Command != "":
		exe = task.NewCommand(t, logger)
	case t.Get != "":
		exe = task.NewGet(t, logger)
	case t.Post != "":
		exe = task.NewPost(t, logger)
	default:
		logger.Fatalln("cannot handle given task config", t)
	}

	onDone := []abstraction.Executable{}
	for _, d := range t.OnDone {
		onDone = append(onDone, CompileTask(&d, logger))
	}
	exe.SetDoneHooks(onDone)
	onFail := []abstraction.Executable{}
	for _, d := range t.OnFail {
		onFail = append(onFail, CompileTask(&d, logger))
	}
	exe.SetFailHooks(onFail)

	return exe
}

package cfgcompiler

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/schedule"
	"github.com/FMotalleb/crontab-go/core/task"
)

func CompileScheduler(sh *config.JobScheduler, cr *cron.Cron, logger *logrus.Entry) abstraction.Scheduler {
	switch {
	case sh.Cron != "":
		scheduler := schedule.NewCron(
			sh.Cron,
			cr,
			logger,
		)
		return &scheduler
	case sh.Interval != 0:
		scheduler := schedule.NewInterval(
			sh.Interval,
			logger,
		)
		return &scheduler

	case sh.OnInit:
		scheduler := schedule.Init{}
		return &scheduler
	}

	return nil
}

func CompileTask(sh *config.Task, logger *logrus.Entry) abstraction.Executable {
	var t abstraction.Executable
	switch {
	case sh.Command != "":
		t = task.NewCommand(sh, logger)
	case sh.Get != "":
		t = task.NewGet(sh, logger)
	case sh.Post != "":
		t = task.NewPost(sh, logger)
	default:
		logger.Fatalln("cannot handle given task config", sh)
	}

	onDone := []abstraction.Executable{}
	for _, d := range sh.OnDone {
		onDone = append(onDone, CompileTask(&d, logger))
	}
	t.SetDoneHooks(onDone)
	onFail := []abstraction.Executable{}
	for _, d := range sh.OnFail {
		onFail = append(onFail, CompileTask(&d, logger))
	}
	t.SetFailHooks(onFail)

	return t
}

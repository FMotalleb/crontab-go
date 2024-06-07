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
	}

	return nil
}

func CompileTask(sh *config.Task, logger *logrus.Entry) abstraction.Executable {
	switch {
	case sh.Command != "":
		return task.NewCommand(sh, logger)
	case sh.Get != "":
		return task.NewGet(sh, logger)
	case sh.Post != "":
		return task.NewPost(sh, logger)
	}
	return nil
}

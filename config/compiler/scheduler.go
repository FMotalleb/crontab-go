package cfgcompiler

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/schedule"
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

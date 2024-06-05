package config

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/core/schedule"
)

func (sh *JobScheduler) Compile(cr *cron.Cron, logger logrus.Entry) abstraction.Scheduler {
	switch {
	case sh.At != nil:
		scheduler := schedule.NewAt(
			*sh.At,
			cr,
			logger,
		)
		return &scheduler
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

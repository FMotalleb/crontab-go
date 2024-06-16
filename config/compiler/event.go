package cfgcompiler

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/schedule"
)

func CompileEvents(sh *config.JobEvents, cr *cron.Cron, logger *logrus.Entry) abstraction.Events {
	switch {
	case sh.Cron != "":
		events := schedule.NewCron(
			sh.Cron,
			cr,
			logger,
		)
		return &events
	case sh.Interval != 0:
		events := schedule.NewInterval(
			sh.Interval,
			logger,
		)
		return &events

	case sh.OnInit:
		events := schedule.Init{}
		return &events
	}

	return nil
}

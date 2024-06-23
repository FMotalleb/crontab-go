// Package cfgcompiler provides mapper functions for the config structs
package cfgcompiler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/event"
)

func CompileEvent(ctx context.Context, sh *config.JobEvent, cr *cron.Cron, logger *logrus.Entry) abstraction.Event {
	switch {
	case sh.Cron != "":
		events := event.NewCron(
			sh.Cron,
			cr,
			logger,
		)
		return &events
	case sh.Interval != 0:
		events := event.NewInterval(
			sh.Interval,
			logger,
		)
		return &events

	case sh.OnInit:
		events := event.Init{}
		return &events
	}

	return nil
}

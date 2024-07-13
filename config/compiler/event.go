// Package cfgcompiler provides mapper functions for the config structs
package cfgcompiler

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/event"
	"github.com/FMotalleb/crontab-go/core/utils"
)

func CompileEvent(sh *config.JobEvent, cr *cron.Cron, logger *logrus.Entry) abstraction.Event {
	switch {
	case sh.Cron != "":
		event := event.NewCron(
			sh.Cron,
			cr,
			logger,
		)
		return &event
	case sh.WebEvent != "":
		event := event.NewEventListener(sh.WebEvent)
		return &event
	case sh.Interval != 0:
		event := event.NewInterval(
			sh.Interval,
			logger,
		)
		return &event

	case sh.OnInit:
		event := event.Init{}
		return &event

	case sh.Docker != nil:
		d := sh.Docker
		con := utils.MayFirstNonZero(d.Connection,
			"unix:///var/run/docker.sock",
		)
		event := event.NewDockerEvent(
			con,
			d.Name,
			d.Image,
			d.Actions,
			d.Labels,
			utils.MayFirstNonZero(d.ErrorLimit, 1),
			utils.MayFirstNonZero(d.ErrorLimitPolicy, event.Reconnect),
			utils.MayFirstNonZero(d.ErrorThrottle, time.Second*5),
			logger,
		)
		return event
	}

	return nil
}

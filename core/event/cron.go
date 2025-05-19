// Package event contains all event emitters supported by this package.
package event

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/global"
)

func init() {
	eg.Register(newCronGenerator)
}

func newCronGenerator(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if cfg.Cron != "" {
		return NewCron(cfg.Cron, global.Get[*cron.Cron](), log)
	}
	return nil
}

type Cron struct {
	cronSchedule string
	logger       *logrus.Entry
	cron         *cron.Cron
	entry        *cron.EntryID
}

func NewCron(schedule string, c *cron.Cron, logger *logrus.Entry) abstraction.EventGenerator {
	cron := &Cron{
		cronSchedule: schedule,
		cron:         c,
		logger: logger.
			WithFields(
				logrus.Fields{
					"scheduler": "cron",
					"cron":      schedule,
				},
			),
	}
	return cron
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Cron) BuildTickChannel() <-chan abstraction.Event {
	if c.entry != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	notifyChan := make(chan abstraction.Event)
	schedule, err := config.DefaultCronParser.Parse(c.cronSchedule)
	if err != nil {
		c.logger.Warnln("cannot initialize cron: ", err)
	} else {
		entry := c.cron.Schedule(
			schedule,
			&cronJob{
				logger:    c.logger,
				scheduler: c.cronSchedule,
				notify:    notifyChan,
			},
		)
		c.entry = &entry
	}
	return notifyChan
}

type cronJob struct {
	logger    *logrus.Entry
	scheduler string
	notify    chan<- abstraction.Event
}

func (j *cronJob) Run() {
	j.logger.Debugln("cron tick received")
	j.notify <- []string{"cron", j.scheduler}
}

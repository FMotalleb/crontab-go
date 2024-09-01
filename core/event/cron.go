// Package event contains all event emitters supported by this package.
package event

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var CronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type Cron struct {
	cronSchedule string
	logger       *logrus.Entry
	cron         *cron.Cron
	entry        *cron.EntryID
}

func NewCron(schedule string, c *cron.Cron, logger *logrus.Entry) Cron {
	return Cron{
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
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Cron) BuildTickChannel() <-chan []string {
	if c.entry != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	notifyChan := make(chan []string)
	schedule, err := CronParser.Parse(c.cronSchedule)
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
	notify    chan<- []string
}

func (j *cronJob) Run() {
	j.logger.Debugln("cron tick received")
	j.notify <- []string{"cron", j.scheduler}
}

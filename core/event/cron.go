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
	notifyChan   chan any
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
func (c *Cron) BuildTickChannel() <-chan any {
	if c.entry != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	c.notifyChan = make(chan any)
	schedule, err := CronParser.Parse(c.cronSchedule)
	if err != nil {
		c.logger.Warnln("cannot initialize cron: ", err)
	} else {
		entry := c.cron.Schedule(schedule, c)
		c.entry = &entry
	}
	return c.notifyChan
}

func (c *Cron) Run() {
	c.logger.Debugln("cron tick received")
	c.notifyChan <- false
}

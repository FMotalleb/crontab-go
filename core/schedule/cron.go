package schedule

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

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

// buildTickChannel implements abstraction.Scheduler.
func (c *Cron) BuildTickChannel() <-chan any {
	c.Cancel()
	c.notifyChan = make(chan any)
	entry, err := c.cron.AddFunc(c.cronSchedule, func() {
		c.logger.Debugln("cron tick received")
		c.notifyChan <- false
	})
	if err != nil {
		c.logger.Warnln("cannot initialize cron: ", err)
	} else {
		c.entry = &entry
	}
	return c.notifyChan
}

// cancel implements abstraction.Scheduler.
func (c *Cron) Cancel() {
	if c.entry != nil {
		c.logger.Debugln("scheduler cancel signal received for an active instance")
		c.cron.Remove(*c.entry)
		close(c.notifyChan)
	}
}

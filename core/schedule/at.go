package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func timeToCron(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d %d *", t.Second(), t.Minute(), t.Hour(), t.Day(), t.Month())
}

type At struct {
	time       time.Time
	logger     *logrus.Entry
	cron       *cron.Cron
	notifyChan chan any
	entry      *cron.EntryID
}

func NewAt(schedule time.Time, c *cron.Cron, logger *logrus.Entry) At {
	return At{
		time: schedule,
		cron: c,
		logger: logger.
			WithFields(
				logrus.Fields{
					"scheduler": "at",
					"time":      schedule,
				},
			),
	}
}

// buildTickChannel implements abstraction.Scheduler.
func (c *At) BuildTickChannel() <-chan any {
	c.Cancel()
	c.notifyChan = make(chan any, 1)
	entry, err := c.cron.AddFunc(timeToCron(c.time), func() {
		c.logger.Debugln("cron tick received for `at` scheduler")
		c.notifyChan <- false
		c.Cancel()
	})
	if err != nil {
		c.logger.Warnln("cannot initialize cron for `at` scheduler: ", err)
	} else {
		c.entry = &entry
	}
	return c.notifyChan
}

// cancel implements abstraction.Scheduler.
func (c *At) Cancel() {
	if c.entry != nil {
		c.logger.Debugln("scheduler cancel signal received for an active instance")
		c.cron.Remove(*c.entry)
		close(c.notifyChan)
	}
}

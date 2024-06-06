package schedule

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
)

func test() {
	sh := &Interval{}
	tester(sh)
}

func tester(sh abstraction.Scheduler) {}

type Interval struct {
	duration   time.Duration
	logger     *logrus.Entry
	ticker     *time.Ticker
	notifyChan chan any
}

func NewInterval(schedule time.Duration, logger *logrus.Entry) Interval {
	return Interval{
		duration: schedule,
		logger: logger.
			WithFields(
				logrus.Fields{
					"scheduler": "interval",
					"interval":  schedule,
				},
			),
	}
}

// buildTickChannel implements abstraction.Scheduler.
func (c *Interval) BuildTickChannel() <-chan any {
	c.Cancel()
	c.notifyChan = make(chan any)

	c.ticker = time.NewTicker(c.duration)
	go func() {
		c.notifyChan <- false
		for range c.ticker.C {
			c.notifyChan <- false
		}
	}()

	return c.notifyChan
}

// cancel implements abstraction.Scheduler.
func (c *Interval) Cancel() {
	if c.ticker != nil {
		c.logger.Debugln("scheduler cancel signal received for an active instance")
		c.ticker.Stop()
		close(c.notifyChan)
	}
}

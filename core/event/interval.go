package event

import (
	"time"

	"github.com/sirupsen/logrus"
)

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

// BuildTickChannel implements abstraction.Scheduler.
func (c *Interval) BuildTickChannel() <-chan any {
	if c.ticker != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	c.notifyChan = make(chan any)
	c.ticker = time.NewTicker(c.duration)
	go func() {
		// c.notifyChan <- false
		for range c.ticker.C {
			c.notifyChan <- false
		}
	}()

	return c.notifyChan
}

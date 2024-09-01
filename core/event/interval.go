package event

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Interval struct {
	duration time.Duration
	logger   *logrus.Entry
	ticker   *time.Ticker
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
func (c *Interval) BuildTickChannel() <-chan []string {
	if c.ticker != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	notifyChan := make(chan []string)
	c.ticker = time.NewTicker(c.duration)
	go func() {
		// c.notifyChan <- false

		for i := range c.ticker.C {
			notifyChan <- []string{"interval", c.duration.String(), i.Format(time.RFC3339)}
		}
	}()

	return notifyChan
}

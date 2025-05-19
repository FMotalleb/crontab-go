package event

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

func init() {
	registerGenerator(newIntervalGenerator)
}

func newIntervalGenerator(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if cfg.Interval != 0 {
		return NewInterval(cfg.Interval, log)
	}
	return nil
}

type Interval struct {
	duration time.Duration
	logger   *logrus.Entry
	ticker   *time.Ticker
}

func NewInterval(schedule time.Duration, logger *logrus.Entry) abstraction.EventGenerator {
	return &Interval{
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

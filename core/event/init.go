package event

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

func init() {
	registerGenerator(newInitGenerator)
}

func newInitGenerator(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if cfg.OnInit {
		return &Init{}
	}
	return nil
}

type Init struct{}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel() <-chan []string {
	notifyChan := make(chan []string)

	go func() {
		notifyChan <- []string{"init"}
		close(notifyChan)
	}()

	return notifyChan
}

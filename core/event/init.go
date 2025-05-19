package event

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

func init() {
	eg.Register(newInitGenerator)
}

func newInitGenerator(log *logrus.Entry, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.OnInit {
		return &Init{}, true
	}
	return nil, false
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

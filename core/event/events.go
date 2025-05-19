package event

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var eg = generator.New[*config.JobEvent, abstraction.EventGenerator]()

func Build(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if g, ok := eg.Get(log, cfg); ok {
		return g
	}
	err := fmt.Errorf("no event generator matched %+v", *cfg)
	log.WithError(err).Warn("event.Build: generator not found")
	return nil
}

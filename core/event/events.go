package event

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var eg = generator.New[*config.JobEvent, abstraction.EventGenerator]()

func Build(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if i, ok := eg.Get(log, cfg); ok {
		return i
	}
	return nil
}

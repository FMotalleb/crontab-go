package event

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

var generators = make([]abstraction.GeneratorMaker, 0)

func registerGenerator(maker abstraction.GeneratorMaker) {
	generators = append(generators, maker)
}

func EventGeneratorOf(log *logrus.Entry, config *config.JobEvent) abstraction.EventGenerator {
	for _, maker := range generators {
		generator := maker(log, config)
		if generator != nil {
			return generator
		}
	}
	return nil
}

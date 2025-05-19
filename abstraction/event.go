// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

type EventGenerator interface {
	BuildTickChannel() <-chan Event
}

type (
	GeneratorMaker func(*logrus.Entry, *config.JobEvent) EventGenerator
	Event          = []string
	EventChannel   = <-chan Event
)

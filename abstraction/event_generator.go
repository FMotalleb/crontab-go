package abstraction

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

type EventGenerator interface {
	CanHandle(*config.JobEvent) bool
	Validatable(*config.JobEvent) error
	Generate(*config.JobEvent, *logrus.Entry) Event
}

// Package generator defines event generators
package generator

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/event"
)

var cronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type CronGenerator struct {
	cron *cron.Cron
}

func NewCronGenerator(cr *cron.Cron) *CronGenerator {
	return &CronGenerator{cron: cr}
}

// CanHandle implements abstraction.EventGenerator.
func (c *CronGenerator) CanHandle(e *config.JobEvent) bool {
	return e.Cron != ""
}

// Validatable implements abstraction.EventGenerator.
func (c *CronGenerator) Validatable(e *config.JobEvent) error {
	_, err := cronParser.Parse(e.Cron)
	return err
}

// Generate implements abstraction.EventGenerator.
func (c *CronGenerator) Generate(e *config.JobEvent, logger *logrus.Entry) abstraction.Event {
	return event.NewCron(e.Cron, c.cron, logger)
}

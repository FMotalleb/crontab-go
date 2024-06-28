package config

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/core/event"
)

func (c *JobConfig) Validate(log *logrus.Entry) error {
	// Log the start of validation
	log.Tracef("Validating JobConfig: %s", c.Name)

	// Check if the job is disabled
	if c.Disabled {
		// Log the disabled job
		log.Debugf("JobConfig %s is disabled", c.Name)
		return nil
	}
	checkList := []func(*JobConfig, *logrus.Entry) error{
		validateEvents,
		validateTasks,
		validateJobHooks,
	}
	for _, check := range checkList {
		if err := check(c, log); err != nil {
			return err
		}
	}

	// Log the successful validation
	log.Tracef("Validation successful for JobConfig: %s", c.Name)
	return nil
}

func validateTasks(c *JobConfig, log *logrus.Entry) error {
	for _, t := range c.Tasks {
		if err := t.Validate(log); err != nil {
			log.Errorf("Validation error in task for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
	return nil
}

func validateEvents(c *JobConfig, log *logrus.Entry) error {
	for _, s := range c.Events {
		if err := s.Validate(log); err != nil {
			log.Errorf("Validation error in event for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
	return nil
}

func validateJobHooks(c *JobConfig, log *logrus.Entry) error {
	for _, t := range c.Hooks.Failed {
		if err := t.Validate(log); err != nil {
			log.Errorf("Validation error in failed hook for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
	for _, t := range c.Hooks.Done {
		if err := t.Validate(log); err != nil {
			log.Errorf("Validation error in done hook for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
	return nil
}

// Validate checks the validity of a JobEvent configuration.
// It ensures that the event has a valid interval or cron expression, and only one of on_init, interval, or cron is set.
// It returns an error if the validation fails, otherwise, it returns nil.
func (s *JobEvent) Validate(log *logrus.Entry) error {
	// Check if the interval is a negative value
	if s.Interval < 0 {
		err := fmt.Errorf("received a negative time in interval: `%v`", s.Interval)
		log.WithError(err).Warn("Validation failed for JobEvent")
		return err
	} else if _, err := event.CronParser.Parse(s.Cron); s.Cron != "" && err != nil {
		log.WithError(err).Warn("Validation failed for JobEvent")
		return err
	}

	// Check the active events to ensure only one of on_init, interval, or cron is set
	events := []bool{
		s.Interval != 0,
		s.Cron != "",
		s.WebEvent != "",
		s.OnInit,
	}
	activeEvents := 0
	for _, t := range events {
		if t {
			activeEvents++
		}
	}
	if activeEvents != 1 {
		err := fmt.Errorf(
			"a single event must have one of (on-init: true,interval,cron,web-event) field, received:(on_init: %t,cron: `%s`, interval: `%s`, web_event: `%s`)",
			s.OnInit,
			s.Cron,
			s.Interval,
			s.WebEvent,
		)
		log.WithError(err).Warn("Validation failed for JobEvent")
		return err
	}

	// Log the successful validation
	log.Tracef("Validation successful for JobEvent: %+v", s)
	return nil
}

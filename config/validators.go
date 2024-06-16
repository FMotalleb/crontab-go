package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	credential "github.com/FMotalleb/crontab-go/core/os_credential"
	"github.com/FMotalleb/crontab-go/core/schedule"
)

// Validate checks the validity of the Config struct.
// It ensures that the log format and log level are valid, and all jobs within the config are also valid.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (cfg *Config) Validate(log *logrus.Entry) error {
	// Validate log format
	if err := cfg.LogFormat.Validate(); err != nil {
		return err
	}

	// Validate log level
	if err := cfg.LogLevel.Validate(); err != nil {
		return err
	}

	// Validate each job in the config
	for _, job := range cfg.Jobs {
		if err := job.Validate(log); err != nil {
			return err
		}
	}

	// All validations passed
	return nil
}

// Validate checks the validity of a JobConfig.
// It ensures that the job is not disabled and all its events, tasks, done hooks, and failed hooks are valid.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
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
		validateDoneHooks,
		validateFailedHooks,
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

func validateFailedHooks(c *JobConfig, log *logrus.Entry) error {
	for _, t := range c.Hooks.Failed {
		if err := t.Validate(log); err != nil {

			log.Errorf("Validation error in failed hook for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
	return nil
}

func validateDoneHooks(c *JobConfig, log *logrus.Entry) error {
	for _, t := range c.Hooks.Done {
		if err := t.Validate(log); err != nil {

			log.Errorf("Validation error in done hook for JobConfig %s: %v", c.Name, err)
			return err
		}
	}
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

// Validate checks the validity of a Task.
// It ensures that the task has exactly one of the Get, Post, or Command fields, and validates other fields based on the specified action.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (t *Task) Validate(log *logrus.Entry) error {
	// Log the start of validation
	log.Tracef("Validating Task: %+v", t)

	// Check the number of action fields
	actions := []bool{
		t.Get != "",
		t.Command != "",
		t.Post != "",
	}
	activeActions := 0
	for _, t := range actions {
		if t {
			activeActions++
		}
	}
	if activeActions != 1 {
		err := fmt.Errorf(
			"a single task should have one of (Get, Post, Command) fields, received:(Command: `%s`, Get: `%s`, Post: `%s`)",
			t.Command,
			t.Get,
			t.Post,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}

	// Validate credentials
	if err := credential.Validate(log, t.UserName, t.GroupName); err != nil {
		log.WithError(err).Warn("Be careful when using credentials, in local mode you can't use credentials unless running as root")
		// return err
	}

	// Validate command-specific fields
	if t.Command != "" && (t.Data != nil || t.Headers != nil) {
		err := fmt.Errorf("command cannot have data or headers field, violating command: `%s`", t.Command)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}

	// Validate GET-specific fields
	if t.Get != "" && t.Data != nil {
		err := fmt.Errorf("GET request cannot have data field, violating GET URI: `%s`", t.Get)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}

	// Validate timeout
	if t.Timeout < 0 {
		err := fmt.Errorf(
			"timeout for tasks cannot be negative received `%d` for %+v",
			t.Timeout,
			t,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}

	// Validate data
	if t.Data != nil {
		_, err := json.Marshal(t.Data)
		if err != nil {
			log.WithError(err).Warn("Validation failed for Task")
			return err
		}
	}

	// Validate retry delay
	if t.RetryDelay < 0 {
		err := fmt.Errorf(
			"retry delay for tasks cannot be negative received `%d` for %+v",
			t.RetryDelay,
			t,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}

	// Validate hooks
	for _, task := range append(t.OnDone, t.OnFail...) {
		if err := task.Validate(log); err != nil {
			joinedErr := errors.Join(errors.New("hook: failed to validate"), err)
			log.WithError(joinedErr).Warn("Validation failed for Task")
			return joinedErr
		}
	}

	// Log the successful validation
	log.Tracef("Validation successful for Task: %+v", t)
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
	} else if _, err := schedule.CronParser.Parse(s.Cron); s.Cron != "" && err != nil {
		log.WithError(err).Warn("Validation failed for JobEvent")
		return err
	}

	// Check the active schedules to ensure only one of on_init, interval, or cron is set
	schedules := []bool{
		s.Interval != 0,
		s.Cron != "",
		s.OnInit,
	}
	activeSchedules := 0
	for _, t := range schedules {
		if t {
			activeSchedules++
		}
	}
	if activeSchedules != 1 {
		err := fmt.Errorf(
			"a single event must have one of (on_init: true,interval,cron) field, received:(on_init: %t,cron: `%s`, interval: `%s`)",
			s.OnInit,
			s.Cron,
			s.Interval,
		)
		log.WithError(err).Warn("Validation failed for JobEvent")
		return err
	}

	// Log the successful validation
	log.Tracef("Validation successful for JobEvent: %+v", s)
	return nil
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	credential "github.com/FMotalleb/crontab-go/core/os_credential"
	"github.com/FMotalleb/crontab-go/core/schedule"
)

func (cfg *Config) Validate(log *logrus.Entry) error {
	if err := cfg.LogFormat.Validate(); err != nil {
		return err
	}
	if err := cfg.LogLevel.Validate(); err != nil {
		return err
	}
	for _, job := range cfg.Jobs {
		if err := job.Validate(log); err != nil {
			return err
		}
	}
	return nil
}

func (c *JobConfig) Validate(log *logrus.Entry) error {
	if c.Disabled == true {
		return nil
	}

	for _, s := range c.Schedulers {
		if err := s.Validate(log); err != nil {
			return err
		}
	}
	for _, t := range c.Tasks {
		if err := t.Validate(log); err != nil {
			return err
		}
	}
	for _, t := range c.Hooks.Done {
		if err := t.Validate(log); err != nil {
			return err
		}
	}
	for _, t := range c.Hooks.Failed {
		if err := t.Validate(log); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) Validate(log *logrus.Entry) error {
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
		return fmt.Errorf(
			"a single task should have one of (get,post,command) fields, received:(command: `%s`, get: `%s`, post: `%s`)",
			t.Command,
			t.Get,
			t.Post,
		)
	}
	if err := credential.Validate(log, t.UserName, t.GroupName); err != nil {
		log.WithError(err).Warn("Be careful when using credentials, in local mode you cant use credentials unless running as root")
	}
	if t.Command != "" && (t.Data != nil || t.Headers != nil) {
		return fmt.Errorf("command cannot have data or headers field, violating command: `%s`", t.Command)
	}
	if t.Get != "" && t.Data != nil {
		return fmt.Errorf("get request cannot have data field, violating get uri: `%s`", t.Get)
	}
	if t.Timeout < 0 {
		return fmt.Errorf(
			"timeout for jobs cannot be negative received `%s` for `%v`",
			t.Timeout,
			t,
		)
	}
	if t.Data != nil {
		_, err := json.Marshal(t.Data)
		if err != nil {
			return fmt.Errorf("cannot marshal the given data: %sr", err)
		}
	}

	if t.RetryDelay < 0 {
		return fmt.Errorf(
			"retry delay for jobs cannot be negative received `%s` for `%v`",
			t.RetryDelay,
			t,
		)
	}
	for _, task := range append(t.OnDone, t.OnFail...) {
		if err := task.Validate(log); err != nil {
			return errors.Join(errors.New("hook: failed to validate"), err)
		}
	}
	return nil
}

// Validate checks the validity of a JobScheduler configuration.
// It ensures that the scheduler has a valid interval or cron expression, and only one of on_init, interval, or cron is set.
// It returns an error if the validation fails, otherwise, it returns nil.
func (s *JobScheduler) Validate(log *logrus.Entry) (_ error) {
	// Check if the interval is a negative value
	if s.Interval < 0 {
		return fmt.Errorf("received a negative time in interval: `%v`", s.Interval)
	} else if _, err := schedule.CronParser.Parse(s.Cron); s.Cron != "" && err != nil {
		return err
	}

	// Check the active schedules to ensure only one of on_init, interval, or cron is set
	schedules := []bool{
		s.Interval != 0,
		s.Cron != "",
		s.OnInit == true,
	}
	activeSchedules := 0
	for _, t := range schedules {
		if t {
			activeSchedules++
		}
	}
	if activeSchedules != 1 {
		return fmt.Errorf(
			"a single scheduler must have one of (on_init: true,interval,cron) field, received:(on_init: %t,cron: `%s`, interval: `%s`)",
			s.OnInit,
			s.Cron,
			s.Interval,
		)
	}
	return
}

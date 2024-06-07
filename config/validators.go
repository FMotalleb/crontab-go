package config

import (
	"encoding/json"
	"fmt"

	"github.com/FMotalleb/crontab-go/core/schedule"
)

func (cfg *Config) Validate() error {
	if err := cfg.LogFormat.Validate(); err != nil {
		return err
	}
	if err := cfg.LogLevel.Validate(); err != nil {
		return err
	}
	for _, job := range cfg.Jobs {
		if err := job.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *JobConfig) Validate() error {
	if c.Disabled == true {
		return nil
	}

	for _, s := range c.Schedulers {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	for _, t := range c.Tasks {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	for _, t := range c.Hooks.Done {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	for _, t := range c.Hooks.Failed {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) Validate() error {
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
	return nil
}

func (s *JobScheduler) Validate() error {
	if s.Interval < 0 {
		return fmt.Errorf("received a negative time in interval: `%v`", s.Interval)
	} else if _, err := schedule.CronParser.Parse(s.Cron); s.Cron != "" && err != nil {
		return err
	}
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
	return nil
}

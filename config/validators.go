package config

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(cron.SecondOptional)

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
	if c.Enabled == false {
		return nil
	}
	if c.Timeout < 0 {
		return fmt.Errorf(
			"timeout for jobs cannot be negative received `%s` for `%s`",
			c.Timeout,
			c.Name,
		)
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf(
			"retry delay for jobs cannot be negative received `%s` for `%s`",
			c.RetryDelay,
			c.Name,
		)
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

func (c *Task) Validate() error {
	actions := []bool{
		c.Get != "",
		c.Command != "",
		c.Post != "",
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
			c.Command,
			c.Get,
			c.Post,
		)
	}
	return nil
}

func (s *JobScheduler) Validate() error {
	if s.At != nil {
		if s.At.Before(time.Now()) {
			fmt.Println("you've set the time in the scheduler that is before now, received:", s.At, "Given time will be ignored")
		}
	} else if s.Interval < 0 {
		return fmt.Errorf("received a negative time in interval: `%v`", s.Interval)
	} else if _, err := cronParser.Parse(s.Cron); err != nil {
		return err
	}
	schedules := []bool{
		s.At != nil,
		s.Interval != 0,
		s.Cron != "",
	}
	activeSchedules := 0
	for _, t := range schedules {
		if t {
			activeSchedules++
		}
	}
	if activeSchedules != 1 {
		return fmt.Errorf(
			"a single scheduler must have one of (at,interval,cron) field, received:(cron: `%s`, interval: `%s`, at: `%s`)",
			s.Cron,
			s.Interval,
			s.At,
		)
	}
	return nil
}

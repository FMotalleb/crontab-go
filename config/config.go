// Package config contains structured representation of config.yaml file
package config

import "time"

type (
	EnvVariables = map[string]string
	JobMetadata  = map[string]string
	Config       struct {
		Log  LogConfig            `yaml:"log"`
		Jobs map[string]JobConfig `yaml:"jobs"`
	}
	LogConfig struct {
		TimeStampFormat string `yaml:"timestamp_format"`
		Format          string `yaml:"format"`
		File            string `yaml:"file"`
		Stdout          bool   `yaml:"stdout"`
		Level           string `yaml:"level"`
	}
	JobConfig struct {
		Description string        `yaml:"description"`
		Enabled     bool          `yaml:"enabled"`
		Exe         JobExe        `yaml:"exe"`
		Scheduler   JobScheduler  `yaml:"scheduler"`
		Retries     int           `yaml:"retries"`
		RetryDelay  time.Duration `yaml:"retry_delay"`
		OnFailure   JobOnFailure  `yaml:"on_failure"`
		Env         EnvVariables  `yaml:"env"`
		Metadata    JobMetadata   `yaml:"metadata"`
	}
	JobExe struct {
		Command string   `yaml:"command"`
		Args    []string `yaml:"args"`
	}
	JobScheduler struct {
		Cron     string        `yaml:"cron"`
		Interval time.Duration `yaml:"interval"`
	}
	JobOnFailure struct {
		Webhooks   []WebHook `yaml:"webhooks"`
		ShouldExit bool      `yaml:"exit"`
	}
	WebHook struct {
		Address string            `yaml:"address"`
		Headers map[string]string `yaml:"headers"`
		Data    map[string]any    `yaml:"data"`
	}
)

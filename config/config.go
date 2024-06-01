// Package config contains structured representation of config.mapstructure file
package config

import (
	"time"
)

type (
	// EnvVariables is a map of environment variables that can be used in the configuration.
	EnvVariables = map[string]string
	// JobMetadata is a map of metadata that can be associated with a job.
	JobMetadata = map[string]string

	// Config is the main configuration struct for the exporter.
	Config struct {
		Log  LogConfig            `mapstructure:"log"`
		Jobs map[string]JobConfig `mapstructure:"jobs"`
	}

	// LogConfig contains the configuration for the logging system.
	LogConfig struct {
		TimeStampFormat string `mapstructure:"timestamp-format"`
		Format          string `mapstructure:"format"`
		File            string `mapstructure:"file"`
		Stdout          bool   `mapstructure:"stdout"`
		Level           string `mapstructure:"level"`
	}

	// JobConfig contains the configuration for a single job.
	JobConfig struct {
		Description string        `mapstructure:"description"`
		Enabled     bool          `mapstructure:"enabled"`
		Exe         JobExe        `mapstructure:"exe"`
		Scheduler   JobScheduler  `mapstructure:"scheduler"`
		Retries     int           `mapstructure:"retries"`
		RetryDelay  time.Duration `mapstructure:"retry-delay"`
		Timeout     time.Duration `mapstructure:"timeout"`
		Hooks       JobHooks      `mapstructure:"hooks"`
		Env         EnvVariables  `mapstructure:"env"`
		Metadata    JobMetadata   `mapstructure:"metadata"`
	}

	// JobExe contains the configuration for the executable that a job runs.
	JobExe struct {
		Command string   `mapstructure:"command"`
		Args    []string `mapstructure:"args"`
	}

	// JobScheduler contains the configuration for a job's scheduling.
	JobScheduler struct {
		Cron     string        `mapstructure:"cron"`
		Interval time.Duration `mapstructure:"interval"`
	}

	// JobHooks contains the configuration for a job's hooks.
	JobHooks struct {
		PreRun []JobHookItem `mapstructure:"pre-run"`
		Done   []JobHookItem `mapstructure:"done"`
		Failed []JobHookItem `mapstructure:"failed"`
	}

	// JobHookItem contains the configuration for a single job hook.
	JobHookItem struct {
		// Webhooks is a list of webhook configurations for the hook.
		Webhooks []WebHook `mapstructure:"webhooks"`
	}

	// WebHook contains the configuration for a single webhook.
	WebHook struct {
		Address string            `mapstructure:"address"`
		Headers map[string]string `mapstructure:"headers"`
		Data    map[string]any    `mapstructure:"data"`
	}
)

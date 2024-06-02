package config

import (
	"time"

	"github.com/FMotalleb/crontab-go/enums"
)

type (
	EnvVariables = map[string]string
	JobMetadata  = map[string]interface{}

	Config struct {
		LogTimestampFormat string                 `mapstructure:"LOG_TIMESTAMP_FORMAT"`
		LogFormat          enums.LoggerFormatType `mapstructure:"LOG_FORMAT"`
		LogFile            string                 `mapstructure:"LOG_FILE"`
		LogStdout          bool                   `mapstructure:"LOG_STDOUT"`
		LogLevel           string                 `mapstructure:"LOG_LEVEL"`
		Jobs               []JobConfig            `mapstructure:"jobs"`
	}

	JobConfig struct {
		Name        string        `mapstructure:"name"`
		Description string        `mapstructure:"description"`
		Enabled     bool          `mapstructure:"enabled"`
		Exe         []Task        `mapstructure:"exe"`
		Scheduler   JobScheduler  `mapstructure:"scheduler"`
		Retries     int           `mapstructure:"retries"`
		RetryDelay  time.Duration `mapstructure:"retry-delay"`
		Timeout     time.Duration `mapstructure:"timeout"`
		Hooks       JobHooks      `mapstructure:"hooks"`
		Env         EnvVariables  `mapstructure:"env"`
		Metadata    JobMetadata   `mapstructure:"metadata"`
	}

	JobScheduler struct {
		Cron     string        `mapstructure:"cron"`
		Interval time.Duration `mapstructure:"interval"`
	}

	JobHooks struct {
		Done   []Task `mapstructure:"done"`
		Failed []Task `mapstructure:"failed"`
	}

	Task struct {
		Post    string            `mapstructure:"get"`
		Get     string            `mapstructure:"post"`
		Command string            `mapstructure:"command"`
		Args    []string          `mapstructure:"args"`
		Headers map[string]string `mapstructure:"headers"`
		Data    map[string]any    `mapstructure:"data"`
	}
)

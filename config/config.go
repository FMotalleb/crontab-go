package config

import (
	"time"

	"github.com/FMotalleb/crontab-go/enums"
)

type (
	EnvVariables = map[string]string
	JobMetadata  = map[string]interface{}

	Config struct {
		LogTimestampFormat string                 `mapstructure:"log_timestamp_format" json:"log_timestamp_format"`
		LogFormat          enums.LoggerFormatType `mapstructure:"log_format" json:"log_format"`
		LogFile            string                 `mapstructure:"log_file" json:"log_file,omitempty"`
		LogStdout          bool                   `mapstructure:"log_stdout" json:"log_stdout"`
		LogLevel           enums.LogLevel         `mapstructure:"log_level" json:"log_level"`
		Jobs               []JobConfig            `mapstructure:"jobs" json:"jobs"`
	}

	JobConfig struct {
		Name        string        `mapstructure:"name" json:"name,omitempty"`
		Description string        `mapstructure:"description" json:"description,omitempty"`
		Enabled     bool          `mapstructure:"enabled" json:"enabled,omitempty"`
		Tasks       []Task        `mapstructure:"tasks" json:"tasks,omitempty"`
		Scheduler   JobScheduler  `mapstructure:"scheduler" json:"scheduler"`
		Retries     int           `mapstructure:"retries" json:"retries,omitempty"`
		RetryDelay  time.Duration `mapstructure:"retry-delay" json:"retry_delay,omitempty"`
		Timeout     time.Duration `mapstructure:"timeout" json:"timeout,omitempty"`
		Hooks       JobHooks      `mapstructure:"hooks" json:"hooks,omitempty"`
		Env         EnvVariables  `mapstructure:"env" json:"env,omitempty"`
		Metadata    JobMetadata   `mapstructure:"metadata" json:"metadata,omitempty"`
	}

	JobScheduler struct {
		Cron     string        `mapstructure:"cron" json:"cron,omitempty"`
		Interval time.Duration `mapstructure:"interval" json:"interval,omitempty"`
		At       time.Time     `mapstructure:"at" json:"at,omitempty"`
	}

	JobHooks struct {
		Done   []Task `mapstructure:"done" json:"done,omitempty"`
		Failed []Task `mapstructure:"failed" json:"failed,omitempty"`
	}

	Task struct {
		Post    string            `mapstructure:"get" json:"post,omitempty"`
		Get     string            `mapstructure:"post" json:"get,omitempty"`
		Command string            `mapstructure:"command" json:"command,omitempty"`
		Args    []string          `mapstructure:"args" json:"args,omitempty"`
		Headers map[string]string `mapstructure:"headers" json:"headers,omitempty"`
		Data    map[string]any    `mapstructure:"data" json:"data,omitempty"`
	}
)

func (t *Task) Initialize() error {
	return nil
}

func (h *JobHooks) Initialize() error {
	return nil
}

func (s *JobScheduler) Initialize() error {
	return nil
}

func (j *JobConfig) Initialize() error {
	return nil
}

func (cfg *Config) Initialize() error {
	if cfg.LogFormat == "" {
		cfg.LogFormat = enums.AnsiLogger
	}
	if err := cfg.LogFormat.Validate(); err != nil {
		return err
	}

	return nil
}

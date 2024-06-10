package config

import (
	"time"

	"github.com/FMotalleb/crontab-go/enums"
)

type (
	Config struct {
		LogTimestampFormat string                 `mapstructure:"log_timestamp_format" json:"log_timestamp_format"`
		LogFormat          enums.LoggerFormatType `mapstructure:"log_format" json:"log_format"`
		LogFile            string                 `mapstructure:"log_file" json:"log_file,omitempty"`
		LogStdout          bool                   `mapstructure:"log_stdout" json:"log_stdout"`
		LogLevel           enums.LogLevel         `mapstructure:"log_level" json:"log_level"`
		Shell              string                 `mapstructure:"shell" json:"shell"`
		ShellArgs          []string               `mapstructure:"shell_args" json:"shell_args"`
		Jobs               []JobConfig            `mapstructure:"jobs" json:"jobs"`
	}

	JobConfig struct {
		Name        string         `mapstructure:"name" json:"name,omitempty"`
		Description string         `mapstructure:"description" json:"description,omitempty"`
		Disabled    bool           `mapstructure:"disabled" json:"disabled,omitempty"`
		Concurrency uint           `mapstructure:"concurrency" json:"concurrency,omitempty"`
		Tasks       []Task         `mapstructure:"tasks" json:"tasks,omitempty"`
		Schedulers  []JobScheduler `mapstructure:"schedulers" json:"schedulers"`
		Hooks       JobHooks       `mapstructure:"hooks" json:"hooks,omitempty"`
	}

	JobScheduler struct {
		Cron     string        `mapstructure:"cron" json:"cron,omitempty"`
		Interval time.Duration `mapstructure:"interval" json:"interval,omitempty"`
		OnInit   bool          `mapstructure:"on-init" json:"on_init,omitempty"`
	}

	JobHooks struct {
		Done   []Task `mapstructure:"done" json:"done,omitempty"`
		Failed []Task `mapstructure:"failed" json:"failed,omitempty"`
	}

	Task struct {
		// Http Requests
		Post    string            `mapstructure:"post" json:"post,omitempty"`
		Get     string            `mapstructure:"get" json:"get,omitempty"`
		Headers map[string]string `mapstructure:"headers" json:"headers,omitempty"`
		Data    any               `mapstructure:"data" json:"data,omitempty"`

		// Command params
		Command          string            `mapstructure:"command" json:"command,omitempty"`
		WorkingDirectory string            `mapstructure:"working-dir" json:"working_directory,omitempty"`
		UserName         string            `mapstructure:"user" json:"user,omitempty"`
		GroupName        string            `mapstructure:"group" json:"group,omitempty"`
		Env              map[string]string `mapstructure:"env" json:"env,omitempty"`
		Connections      []TaskConnection  `mapstructure:"connections" json:"connections,omitempty"`

		// Retry & Timeout config
		Retries    uint          `mapstructure:"retries" json:"retries,omitempty"`
		RetryDelay time.Duration `mapstructure:"retry-delay" json:"retry_delay,omitempty"`
		Timeout    time.Duration `mapstructure:"timeout" json:"timeout,omitempty"`

		// Hooks
		OnDone []Task `mapstructure:"on-done" json:"on_done,omitempty"`
		OnFail []Task `mapstructure:"on-fail" json:"on_fail,omitempty"`
	}
	TaskConnection struct {
		Local            bool   `mapstructure:"local" json:"local,omitempty"`
		DockerConnection string `mapstructure:"docker" json:"docker,omitempty"`
		ContainerName    string `mapstructure:"container" json:"container,omitempty"`
	}
)

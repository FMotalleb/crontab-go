// Package config contains structured representation of config.yaml file
package config

import "time"

type Config struct {
	Log  LogConfig            `yaml:"log"`
	Jobs map[string]JobConfig `yaml:"jobs"`
}

type LogConfig struct {
	TimeStampFormat string `yaml:"timestamp_format"`
	Format          string `yaml:"format"`
	File            string `yaml:"file"`
	Stdout          bool   `yaml:"stdout"`
	Level           string `yaml:"level"`
}

type JobConfig struct {
	Description string            `yaml:"description"`
	Enabled     bool              `yaml:"enabled"`
	Exe         JobExe            `yaml:"exe"`
	Scheduler   JobScheduler      `yaml:"scheduler"`
	Retries     int               `yaml:"retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay"`
	OnFailure   JobOnFailure      `yaml:"on_failure"`
	Env         map[string]string `yaml:"env"`
	Metadata    JobMetadata       `yaml:"metadata"`
}

type JobExe struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

type JobScheduler struct {
	Cron     string `yaml:"cron"`
	Interval string `yaml:"interval"`
}

type JobOnFailure struct {
	Webhooks   []WebHook `yaml:"webhooks"`
	ShouldExit bool      `yaml:"exit"`
}
type WebHook struct {
	Address string            `yaml:"address"`
	Headers map[string]string `yaml:"headers"`
	Data    map[string]any    `yaml:"data"`
}

type JobMetadata = map[string]string

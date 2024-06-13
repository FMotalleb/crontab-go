package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/enums"
)

var failJob config.JobConfig = config.JobConfig{
	Disabled: false,
	Tasks: []config.Task{
		{},
	},
}

var okJob config.JobConfig = config.JobConfig{
	Disabled: false,
	Tasks: []config.Task{
		{
			Post: "https://localhost",
		},
	},
}

func TestConfig_Validate_LogFormatFails(t *testing.T) {
	cfg := &config.Config{
		LogFormat: enums.LoggerFormatType("unknown"),
		LogLevel:  enums.DebugLevel,
		Jobs:      []config.JobConfig{},
	}
	log := logrus.NewEntry(logrus.New())
	err := cfg.Validate(log)
	assert.Error(t, err)
	assert.Equal(t, "Given Logger format: `unknown`", err.Error())
}

func TestConfig_Validate_LogLevelFails(t *testing.T) {
	cfg := &config.Config{
		LogFormat: enums.AnsiLogger,
		LogLevel:  enums.LogLevel("unknown"),
		Jobs:      []config.JobConfig{},
	}
	log := logrus.NewEntry(logrus.New())
	err := cfg.Validate(log)
	assert.Error(t, err)
	assert.Equal(t, "LogLevel must be one of (trace,debug,info,warn,fatal,panic) but received `unknown`", err.Error())
}

func TestConfig_Validate_JobFails(t *testing.T) {
	cfg := &config.Config{
		LogFormat: enums.JsonLogger,
		LogLevel:  enums.FatalLevel,
		Jobs:      []config.JobConfig{failJob},
	}
	log := logrus.NewEntry(logrus.New())
	err := cfg.Validate(log)
	assert.Error(t, err)
}

func TestConfig_Validate_AllValidationsPass(t *testing.T) {
	cfg := &config.Config{
		LogFormat: enums.JsonLogger,
		LogLevel:  enums.DebugLevel,
		Jobs: []config.JobConfig{
			okJob,
		},
	}
	log := logrus.NewEntry(logrus.New())
	err := cfg.Validate(log)
	assert.NoError(t, err)
}

func TestConfig_Validate_NoJobs(t *testing.T) {
	cfg := &config.Config{
		LogFormat: enums.JsonLogger,
		LogLevel:  enums.DebugLevel,
		Jobs:      []config.JobConfig{},
	}
	log := logrus.NewEntry(logrus.New())
	err := cfg.Validate(log)
	assert.NoError(t, err)
}

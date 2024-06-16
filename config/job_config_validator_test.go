package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

func TestJobConfig_Validate_Disabled(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	jobConfig := &config.JobConfig{
		Disabled: true,
	}

	err := jobConfig.Validate(log)
	assert.NoError(t, err, "Expected no error when job is disabled")
}

func TestJobConfig_Validate_Eventss(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	jobConfig := &config.JobConfig{
		Eventss: []config.JobEvents{
			{Interval: -1}, // Invalid interval
		},
	}

	err := jobConfig.Validate(log)
	assert.Error(t, err, "Expected error due to invalid events interval")
}

func TestJobConfig_Validate_Tasks(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	jobConfig := &config.JobConfig{
		Tasks: []config.Task{
			{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
		},
	}

	err := jobConfig.Validate(log)
	assert.Error(t, err, "Expected error due to invalid task configuration")
}

func TestJobConfig_Validate_HooksDone(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	jobConfig := &config.JobConfig{
		Hooks: config.JobHooks{
			Done: []config.Task{
				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
			},
		},
	}

	err := jobConfig.Validate(log)
	assert.Error(t, err, "Expected error due to invalid done hook task configuration")
}

func TestJobConfig_Validate_HooksFailed(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	jobConfig := &config.JobConfig{
		Hooks: config.JobHooks{
			Failed: []config.Task{
				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
			},
		},
	}

	err := jobConfig.Validate(log)
	assert.Error(t, err, "Expected error due to invalid failed hook task configuration")
}

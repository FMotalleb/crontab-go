package cfgcompiler_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
)

func TestCompileTask_NonExistingTask(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	taskConfig := &config.Task{}
	assert.Panics(
		t,
		func() {
			cfgcompiler.CompileTask(taskConfig, logger)
		},
	)
}

func TestCompileTask_GetTask(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	taskConfig := &config.Task{
		Get: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, logger)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_CommandTask(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	taskConfig := &config.Task{
		Command: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, logger)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_PostTask(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	taskConfig := &config.Task{
		Post: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, logger)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_WithHooks(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	taskConfig := &config.Task{
		Command: "test",
		OnDone: []config.Task{
			{
				Command: "test",
			},
		},
		OnFail: []config.Task{
			{
				Command: "test",
			},
		},
	}
	exe := cfgcompiler.CompileTask(taskConfig, logger)
	assert.NotEqual(t, exe, nil)
}

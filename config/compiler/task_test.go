package cfgcompiler_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
)

func TestCompileTask_NonExistingTask(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
	taskConfig := &config.Task{}
	assert.Panics(
		t,
		func() {
			cfgcompiler.CompileTask(taskConfig, log)
		},
	)
}

func TestCompileTask_GetTask(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
	taskConfig := &config.Task{
		Get: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, log)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_CommandTask(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
	taskConfig := &config.Task{
		Command: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, log)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_PostTask(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
	taskConfig := &config.Task{
		Post: "test",
	}
	exe := cfgcompiler.CompileTask(taskConfig, log)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_WithHooks(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
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
	exe := cfgcompiler.CompileTask(taskConfig, log)
	assert.NotEqual(t, exe, nil)
}

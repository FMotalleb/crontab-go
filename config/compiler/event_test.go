package cfgcompiler_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/event"
	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
)

// TestCompileEvent_IntervalZero tests that CompileEvents returns nil when Interval is zero
func TestCompileEvent_IntervalZero(t *testing.T) {
	sh := &config.JobEvent{Interval: 0}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)
	event := cfgcompiler.CompileEvent(sh, cr, log)
	assert.Equal(t, event, nil)
}

func TestCompileEvent_IntervalNonZero(t *testing.T) {
	sh := &config.JobEvent{Interval: 15}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	sch := cfgcompiler.CompileEvent(sh, cr, log)
	_, ok := sch.(*event.Interval)
	assert.Equal(t, ok, true)
}

// TestCompileEvent_IntervalZeroWithCronSet tests CompileEvents with Interval zero but Cron expression set
func TestCompileEvent_IntervalZeroWithCronSet(t *testing.T) {
	sh := &config.JobEvent{Cron: "0 * * * *", Interval: 0}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	e := cfgcompiler.CompileEvent(sh, cr, log)
	if _, ok := e.(*event.Cron); !ok {
		t.Errorf("Expected Cron events, got %T", e)
	}
}

// TestCompileEvent_IntervalZeroWithOnInitSet tests CompileEvents with Interval zero and OnInit set
func TestCompileEvent_IntervalZeroWithOnInitSet(t *testing.T) {
	sh := &config.JobEvent{OnInit: true, Interval: 0}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	e := cfgcompiler.CompileEvent(sh, cr, log)
	if _, ok := e.(*event.Init); !ok {
		t.Errorf("Expected Init events, got %T", e)
	}
}

// TestCompileEvent_IntervalZeroWithAllFieldsEmpty tests CompileEvents with Interval zero and all other fields empty
func TestCompileEvent_IntervalZeroWithAllFieldsEmpty(t *testing.T) {
	sh := &config.JobEvent{Interval: 0}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	e := cfgcompiler.CompileEvent(sh, cr, log)
	if e != nil {
		t.Errorf("Expected nil, got %v", e)
	}
}

// TestCompileEvent_IntervalZeroWithCronAndOnInitSet tests CompileEvent with Interval zero, Cron expression, and OnInit set
func TestCompileEvent_IntervalZeroWithCronAndOnInitSet(t *testing.T) {
	sh := &config.JobEvent{Cron: "0 * * * *", OnInit: true, Interval: 0}
	cr := cron.New()
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	e := cfgcompiler.CompileEvent(sh, cr, log)
	if _, ok := e.(*event.Cron); !ok {
		t.Errorf("Expected Cron event, got %T", e)
	}
}

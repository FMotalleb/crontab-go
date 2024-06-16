package cfgcompiler_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/schedule"
)

// TestCompileEvents_IntervalZero tests that CompileEvents returns nil when Interval is zero
func TestCompileEvents_IntervalZero(t *testing.T) {
	sh := &config.JobEvents{Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	events := cfgcompiler.CompileEvents(sh, cr, logger)
	assert.Equal(t, events, nil)
}

func TestCompileEvents_IntervalNonZero(t *testing.T) {
	sh := &config.JobEvents{Interval: 15}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	sch := cfgcompiler.CompileEvents(sh, cr, logger)
	_, ok := sch.(*schedule.Interval)
	assert.Equal(t, ok, true)
}

// TestCompileEvents_IntervalZeroWithCronSet tests CompileEvents with Interval zero but Cron expression set
func TestCompileEvents_IntervalZeroWithCronSet(t *testing.T) {
	sh := &config.JobEvents{Cron: "0 * * * *", Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	events := cfgcompiler.CompileEvents(sh, cr, logger)
	if _, ok := events.(*schedule.Cron); !ok {
		t.Errorf("Expected Cron events, got %T", events)
	}
}

// TestCompileEvents_IntervalZeroWithOnInitSet tests CompileEvents with Interval zero and OnInit set
func TestCompileEvents_IntervalZeroWithOnInitSet(t *testing.T) {
	sh := &config.JobEvents{OnInit: true, Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	events := cfgcompiler.CompileEvents(sh, cr, logger)
	if _, ok := events.(*schedule.Init); !ok {
		t.Errorf("Expected Init events, got %T", events)
	}
}

// TestCompileEvents_IntervalZeroWithAllFieldsEmpty tests CompileEvents with Interval zero and all other fields empty
func TestCompileEvents_IntervalZeroWithAllFieldsEmpty(t *testing.T) {
	sh := &config.JobEvents{Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	events := cfgcompiler.CompileEvents(sh, cr, logger)
	if events != nil {
		t.Errorf("Expected nil, got %v", events)
	}
}

// TestCompileEvents_IntervalZeroWithCronAndOnInitSet tests CompileEvents with Interval zero, Cron expression, and OnInit set
func TestCompileEvents_IntervalZeroWithCronAndOnInitSet(t *testing.T) {
	sh := &config.JobEvents{Cron: "0 * * * *", OnInit: true, Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	events := cfgcompiler.CompileEvents(sh, cr, logger)
	if _, ok := events.(*schedule.Cron); !ok {
		t.Errorf("Expected Cron events, got %T", events)
	}
}

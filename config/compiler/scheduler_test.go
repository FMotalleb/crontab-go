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

// TestCompileScheduler_IntervalZero tests that CompileScheduler returns nil when Interval is zero
func TestCompileScheduler_IntervalZero(t *testing.T) {
	sh := &config.JobScheduler{Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	scheduler := cfgcompiler.CompileScheduler(sh, cr, logger)
	assert.Equal(t, scheduler, nil)
}

func TestCompileScheduler_IntervalNonZero(t *testing.T) {
	sh := &config.JobScheduler{Interval: 15}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	sch := cfgcompiler.CompileScheduler(sh, cr, logger)
	_, ok := sch.(*schedule.Interval)
	assert.Equal(t, ok, true)
}

// TestCompileScheduler_IntervalZeroWithCronSet tests CompileScheduler with Interval zero but Cron expression set
func TestCompileScheduler_IntervalZeroWithCronSet(t *testing.T) {
	sh := &config.JobScheduler{Cron: "0 * * * *", Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	scheduler := cfgcompiler.CompileScheduler(sh, cr, logger)
	if _, ok := scheduler.(*schedule.Cron); !ok {
		t.Errorf("Expected Cron scheduler, got %T", scheduler)
	}
}

// TestCompileScheduler_IntervalZeroWithOnInitSet tests CompileScheduler with Interval zero and OnInit set
func TestCompileScheduler_IntervalZeroWithOnInitSet(t *testing.T) {
	sh := &config.JobScheduler{OnInit: true, Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	scheduler := cfgcompiler.CompileScheduler(sh, cr, logger)
	if _, ok := scheduler.(*schedule.Init); !ok {
		t.Errorf("Expected Init scheduler, got %T", scheduler)
	}
}

// TestCompileScheduler_IntervalZeroWithAllFieldsEmpty tests CompileScheduler with Interval zero and all other fields empty
func TestCompileScheduler_IntervalZeroWithAllFieldsEmpty(t *testing.T) {
	sh := &config.JobScheduler{Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	scheduler := cfgcompiler.CompileScheduler(sh, cr, logger)
	if scheduler != nil {
		t.Errorf("Expected nil, got %v", scheduler)
	}
}

// TestCompileScheduler_IntervalZeroWithCronAndOnInitSet tests CompileScheduler with Interval zero, Cron expression, and OnInit set
func TestCompileScheduler_IntervalZeroWithCronAndOnInitSet(t *testing.T) {
	sh := &config.JobScheduler{Cron: "0 * * * *", OnInit: true, Interval: 0}
	cr := cron.New()
	logger := logrus.NewEntry(logrus.StandardLogger())

	scheduler := cfgcompiler.CompileScheduler(sh, cr, logger)
	if _, ok := scheduler.(*schedule.Cron); !ok {
		t.Errorf("Expected Cron scheduler, got %T", scheduler)
	}
}

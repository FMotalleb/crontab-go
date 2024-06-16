package config_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/alecthomas/assert/v2"

	"github.com/FMotalleb/crontab-go/config"
)

func TestJobEvents_Validate_PositiveInterval(t *testing.T) {
	events := config.JobEvents{
		Interval: 10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_events_validate")

	err := events.Validate(log)

	assert.NoError(t, err)
}

func TestJobEvents_Validate_CorrectCron(t *testing.T) {
	events := config.JobEvents{
		Interval: 0,
		Cron:     "* * * * *",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_events_validate")

	err := events.Validate(log)
	assert.NoError(t, err)
}

func TestJobEvents_Validate_NegativeInterval(t *testing.T) {
	events := config.JobEvents{
		Interval: -10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_events_validate")

	err := events.Validate(log)

	expectedErr := "received a negative time in interval: `-10ns`"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestJobEvents_Validate_InvalidCronExpression(t *testing.T) {
	events := config.JobEvents{
		Interval: 0,
		Cron:     "invalid_cron_expression",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_events_validate")

	err := events.Validate(log)

	assert.Error(t, err)
}

func TestJobEvents_Validate_MultipleActiveSchedules(t *testing.T) {
	events := config.JobEvents{
		Interval: 60,
		Cron:     "0 0 * * *",
		OnInit:   true,
	}
	log := logrus.New().WithField("test", "job_events_validate")

	err := events.Validate(log)

	expectedErr := "a single events must have one of (on_init: true,interval,cron) field, received:(on_init: true,cron: `0 0 * * *`, interval: `60ns`)"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

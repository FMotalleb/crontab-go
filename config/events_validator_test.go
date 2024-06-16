package config_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/alecthomas/assert/v2"

	"github.com/FMotalleb/crontab-go/config"
)

func TestJobEvent_Validate_PositiveInterval(t *testing.T) {
	event := config.JobEvent{
		Interval: 10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_event_validate")

	err := event.Validate(log)

	assert.NoError(t, err)
}

func TestJobEvent_Validate_CorrectCron(t *testing.T) {
	event := config.JobEvent{
		Interval: 0,
		Cron:     "* * * * *",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_event_validate")

	err := event.Validate(log)
	assert.NoError(t, err)
}

func TestJobEvent_Validate_NegativeInterval(t *testing.T) {
	event := config.JobEvent{
		Interval: -10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_event_validate")

	err := event.Validate(log)

	expectedErr := "received a negative time in interval: `-10ns`"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestJobEvent_Validate_InvalidCronExpression(t *testing.T) {
	event := config.JobEvent{
		Interval: 0,
		Cron:     "invalid_cron_expression",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_event_validate")

	err := event.Validate(log)

	assert.Error(t, err)
}

func TestJobEvent_Validate_MultipleActiveSchedules(t *testing.T) {
	event := config.JobEvent{
		Interval: 60,
		Cron:     "0 0 * * *",
		OnInit:   true,
	}
	log := logrus.New().WithField("test", "job_event_validate")

	err := event.Validate(log)

	expectedErr := "a single event must have one of (on_init: true,interval,cron) field, received:(on_init: true,cron: `0 0 * * *`, interval: `60ns`)"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

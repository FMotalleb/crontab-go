package config_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/alecthomas/assert/v2"

	"github.com/FMotalleb/crontab-go/config"
)

func TestJobScheduler_Validate_PositiveInterval(t *testing.T) {
	scheduler := config.JobScheduler{
		Interval: 10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_scheduler_validate")

	err := scheduler.Validate(log)

	assert.NoError(t, err)
}

func TestJobScheduler_Validate_CorrectCron(t *testing.T) {
	scheduler := config.JobScheduler{
		Interval: 0,
		Cron:     "* * * * *",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_scheduler_validate")

	err := scheduler.Validate(log)
	assert.NoError(t, err)
}

func TestJobScheduler_Validate_NegativeInterval(t *testing.T) {
	scheduler := config.JobScheduler{
		Interval: -10,
		Cron:     "",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_scheduler_validate")

	err := scheduler.Validate(log)

	expectedErr := "received a negative time in interval: `-10ns`"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestJobScheduler_Validate_InvalidCronExpression(t *testing.T) {
	scheduler := config.JobScheduler{
		Interval: 0,
		Cron:     "invalid_cron_expression",
		OnInit:   false,
	}
	log := logrus.New().WithField("test", "job_scheduler_validate")

	err := scheduler.Validate(log)

	assert.Error(t, err)
}

func TestJobScheduler_Validate_MultipleActiveSchedules(t *testing.T) {
	scheduler := config.JobScheduler{
		Interval: 60,
		Cron:     "0 0 * * *",
		OnInit:   true,
	}
	log := logrus.New().WithField("test", "job_scheduler_validate")

	err := scheduler.Validate(log)

	expectedErr := "a single scheduler must have one of (on_init: true,interval,cron) field, received:(on_init: true,cron: `0 0 * * *`, interval: `60ns`)"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

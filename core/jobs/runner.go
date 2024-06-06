package jobs

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func InitializeJobs(log *logrus.Entry, cronInstance *cron.Cron) {
	for _, job := range cmd.CFG.Jobs {
		if job.Disabled {
			log.Warnf("job %s is disabled", job.Name)
			continue
		}

		c := context.Background()
		c = context.WithValue(c, ctxutils.JobKey, job)

		logger := initLogger(c, log, job)

		if err := job.Validate(); err != nil {
			log.Panicf("failed to validate job (%s): %v", job.Name, err)
		}

		schedulers := initSchedulers(job, cronInstance, logger)
		logger.Trace("Schedulers initialized")

		tasks, doneHooks, failHooks := initTasks(job, logger)
		logger.Trace("Tasks initialized")

		signal := initEventSignal(schedulers, logger)

		go taskHandler(c, logger, signal, tasks, doneHooks, failHooks)
		logger.Trace("EventLoop initialized")
	}
	log.Debugln("Jobs Are Ready")
}

func initLogger(c context.Context, log *logrus.Entry, job config.JobConfig) *logrus.Entry {
	logger := log.WithContext(c).WithField("job.name", job.Name)
	logger.Trace("Initializing Job")
	return logger
}

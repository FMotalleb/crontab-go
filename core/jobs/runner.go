package jobs

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func InitializeJobs(log *logrus.Entry, cronInstance *cron.Cron) {
	for _, job := range cmd.CFG.Jobs {
		if job.Disabled {
			log.Warnf("job %s is disabled", job.Name)
			continue
		}
		// Setting default value of concurrency
		if job.Concurrency == 0 {
			job.Concurrency = 1
		}

		c := context.Background()
		c = context.WithValue(c, ctxutils.JobKey, job)

		var lock sync.Locker = concurrency.NewConcurrentPool(job.Concurrency)

		logger := initLogger(c, log, job)
		logger = logger.WithField("concurrency", job.Concurrency)
		if err := job.Validate(log); err != nil {
			log.Panicf("failed to validate job (%s): %v", job.Name, err)
		}

		signal := buildSignal(job, cronInstance, logger)

		tasks, doneHooks, failHooks := initTasks(job, logger)
		logger.Trace("Tasks initialized")

		go taskHandler(c, logger, signal, tasks, doneHooks, failHooks, lock)
		logger.Trace("EventLoop initialized")
	}
	log.Debugln("Jobs Are Ready")
}

func buildSignal(job config.JobConfig, cronInstance *cron.Cron, logger *logrus.Entry) <-chan any {
	schedulers := initSchedulers(job, cronInstance, logger)
	logger.Trace("Schedulers initialized")

	signal := initEventSignal(schedulers, logger)

	return signal
}

func initLogger(c context.Context, log *logrus.Entry, job config.JobConfig) *logrus.Entry {
	logger := log.WithContext(c).WithField("job.name", job.Name)
	logger.Trace("Initializing Job")
	return logger
}

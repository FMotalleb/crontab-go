package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/core/global"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func InitializeJobs(log *logrus.Entry) {
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
		c = context.WithValue(c, ctxutils.JobKey, job.Name)

		lock, err := concurrency.NewConcurrentPool(job.Concurrency)
		if err != nil {
			log.Panicf("failed to validate job (%s): %v", job.Name, err)
		}
		logger := initLogger(c, log, job)
		logger = logger.WithField("concurrency", job.Concurrency)
		if err := job.Validate(log); err != nil {
			log.Panicf("failed to validate job (%s): %v", job.Name, err)
		}

		signal := buildSignal(*job, logger)
		signal = global.CTX().CountSignals(c, "events", signal, "amount of events dispatched for this job", prometheus.Labels{})
		tasks, doneHooks, failHooks := initTasks(*job, logger)
		logger.Trace("Tasks initialized")

		go taskHandler(c, logger, signal, tasks, doneHooks, failHooks, lock)
		logger.Trace("EventLoop initialized")
	}
	log.Debugln("Jobs Are Ready")
}

func buildSignal(job config.JobConfig, logger *logrus.Entry) abstraction.EventChannel {
	events := initEvents(job, logger)
	logger.Trace("Events initialized")

	signal := initEventSignal(events, logger)

	return signal
}

func initLogger(c context.Context, log *logrus.Entry, job *config.JobConfig) *logrus.Entry {
	logger := log.WithContext(c).WithField("job.name", job.Name)
	logger.Trace("Initializing Job")
	return logger
}

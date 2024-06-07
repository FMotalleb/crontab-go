package jobs

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/goutils"
)

func initEventSignal(schedulers []abstraction.Scheduler, logger *logrus.Entry) <-chan any {
	signals := make([]<-chan any, 0, len(schedulers))
	for _, sh := range schedulers {
		signals = append(signals, sh.BuildTickChannel())
	}
	logger.Trace("Signals Built")
	signal := goutils.ZipChannels(signals...)
	return signal
}

func initTasks(job config.JobConfig, logger *logrus.Entry) ([]abstraction.Executable, []abstraction.Executable, []abstraction.Executable) {
	tasks := make([]abstraction.Executable, 0, len(job.Tasks))
	doneHooks := make([]abstraction.Executable, 0, len(job.Hooks.Done))
	failHooks := make([]abstraction.Executable, 0, len(job.Hooks.Failed))
	for _, t := range job.Tasks {
		tasks = append(tasks, cfgcompiler.CompileTask(&t, logger))
	}
	logger.Trace("Compiled Tasks")
	for _, t := range job.Hooks.Done {
		doneHooks = append(doneHooks, cfgcompiler.CompileTask(&t, logger))
	}
	logger.Trace("Compiled Hooks.Done")
	for _, t := range job.Hooks.Failed {
		failHooks = append(failHooks, cfgcompiler.CompileTask(&t, logger))
	}
	logger.Trace("Compiled Hooks.Fail")
	return tasks, doneHooks, failHooks
}

func initSchedulers(job config.JobConfig, cronInstance *cron.Cron, logger *logrus.Entry) []abstraction.Scheduler {
	schedulers := make([]abstraction.Scheduler, 0, len(job.Schedulers))
	for _, sh := range job.Schedulers {
		schedulers = append(schedulers, cfgcompiler.CompileScheduler(&sh, cronInstance, logger))
	}
	return schedulers
}

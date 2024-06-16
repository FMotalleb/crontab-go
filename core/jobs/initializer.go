package jobs

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/goutils"
)

func initEventSignal(eventss []abstraction.Events, logger *logrus.Entry) <-chan any {
	signals := make([]<-chan any, 0, len(eventss))
	for _, sh := range eventss {
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

func initEventss(job config.JobConfig, cronInstance *cron.Cron, logger *logrus.Entry) []abstraction.Events {
	eventss := make([]abstraction.Events, 0, len(job.Eventss))
	for _, sh := range job.Eventss {
		eventss = append(eventss, cfgcompiler.CompileEvents(&sh, cronInstance, logger))
	}
	return eventss
}

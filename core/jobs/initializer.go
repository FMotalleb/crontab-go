// Package jobs implements the main functionality for the jobs in the application
package jobs

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/utils"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func initEventSignal(events []abstraction.Event, logger *logrus.Entry) abstraction.EventChannel {
	signals := make([]<-chan []string, 0, len(events))
	for _, sh := range events {
		signals = append(signals, sh.BuildTickChannel())
	}
	logger.Trace("Signals Built")
	signal := utils.ZipChannels(signals...)
	return signal
}

func initTasks(job config.JobConfig, logger *logrus.Entry) ([]abstraction.Executable, []abstraction.Executable, []abstraction.Executable) {
	tasks := make([]abstraction.Executable, 0, len(job.Tasks))
	doneHooks := make([]abstraction.Executable, 0, len(job.Hooks.Done))
	failHooks := make([]abstraction.Executable, 0, len(job.Hooks.Failed))

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxutils.JobKey, job.Name)
	for _, t := range job.Tasks {
		tasks = append(tasks, cfgcompiler.CompileTask(ctx, &t, logger))
	}
	logger.Trace("Compiled Tasks")
	for _, t := range job.Hooks.Done {
		doneHooks = append(doneHooks, cfgcompiler.CompileTask(ctx, &t, logger))
	}
	logger.Trace("Compiled Hooks.Done")
	for _, t := range job.Hooks.Failed {
		failHooks = append(failHooks, cfgcompiler.CompileTask(ctx, &t, logger))
	}
	logger.Trace("Compiled Hooks.Fail")
	return tasks, doneHooks, failHooks
}

func initEvents(job config.JobConfig, cronInstance *cron.Cron, logger *logrus.Entry) []abstraction.Event {
	events := make([]abstraction.Event, 0, len(job.Events))
	for _, sh := range job.Events {
		events = append(events, cfgcompiler.CompileEvent(&sh, cronInstance, logger))
	}
	return events
}

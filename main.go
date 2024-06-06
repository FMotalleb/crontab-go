/*
Copyright © 2024 Motalleb Fallahnezhad

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
	cfgcompiler "github.com/FMotalleb/crontab-go/config/compiler"
	"github.com/FMotalleb/crontab-go/core/goutils"
	"github.com/FMotalleb/crontab-go/ctxutils"
	cx "github.com/FMotalleb/crontab-go/ctxutils"
	"github.com/FMotalleb/crontab-go/logger"
)

var (
	log *logrus.Entry
	ctx cx.Context
)

func main() {
	cmd.Execute()
	ctx = cx.NewContext("core")
	logger.InitFromConfig()
	log = logger.SetupLogger("Crontab-GO")
	cronInstance := cron.New(cron.WithSeconds())

	log.Info("Booting up")
	log.Infoln(cmd.CFG)
	for _, job := range cmd.CFG.Jobs {
		if !job.Enabled {
			log.Warn("job %s is disabled", job.Name)
			continue
		}
		c := context.Background()
		c = context.WithValue(c, ctxutils.JobKey, job)
		logger := log.WithContext(c).WithField("job.name", job.Name)
		logger.Trace("Initializing Job")
		if err := job.Validate(); err != nil {
			log.Panicln("failed to validate job: ", err)
		}
		schedulers := make([]abstraction.Scheduler, 0, len(job.Schedulers))
		for _, sh := range job.Schedulers {
			schedulers = append(schedulers, cfgcompiler.CompileScheduler(&sh, cronInstance, logger))
		}
		logger.Trace("Compiled Schedulers")
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

		signals := make([]<-chan any, 0, len(schedulers))

		for _, sh := range schedulers {
			signals = append(signals, sh.BuildTickChannel())
		}
		logger.Trace("Signals Built")
		signal := goutils.Zip(signals...)

		logger.Infof("Zipping Signals")
		go func() {
			logger.Debug("Spawned work goroutine")
			for range signal {
				logger.Debug("Signal Received")
				for _, task := range tasks {
					ctx := context.Background()
					err := task.Execute(ctx)
					switch err {
					case nil:
						for _, task := range doneHooks {
							_ = task.Execute(ctx)
						}
					default:
						for _, task := range failHooks {
							_ = task.Execute(ctx)
						}
					}
				}
			}
		}()
	}
	cronInstance.Start()
	<-make(chan any)
}

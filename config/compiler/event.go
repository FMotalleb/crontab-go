// Package cfgcompiler provides mapper functions for the config structs
package cfgcompiler

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/event"
)

func CompileEvent(sh *config.JobEvent, logger *logrus.Entry) abstraction.EventGenerator {
	return event.EventGeneratorOf(logger, *sh)
	// switch {

	// case sh.WebEvent != "":
	// 	event := event.NewEventListener(sh.WebEvent)

	// case sh.LogFile != "":
	// 	e, err := event.NewLogFile(
	// 		sh.LogFile,
	// 		sh.LogLineBreaker,
	// 		sh.LogMatcher,
	// 		sh.LogCheckCycle,
	// 		logger,
	// 	)
	// 	if err != nil {
	// 		logger.Error("Error creating LogFile: ", err)
	// 		return nil
	// 	}
	// 	return e
	// }
	// return nil
}

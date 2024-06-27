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
	"github.com/robfig/cron/v3"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/core/global"
	"github.com/FMotalleb/crontab-go/core/jobs"
	"github.com/FMotalleb/crontab-go/core/webserver"
	"github.com/FMotalleb/crontab-go/logger"
)

func main() {
	cmd.Execute()
	logger.InitFromConfig()
	log := logger.SetupLogger("Crontab-GO")
	// TODO: move somewhere else
	cronInstance := cron.New(cron.WithSeconds())
	log.Info("Booting up")
	jobs.InitializeJobs(log, cronInstance)
	if cmd.CFG.WebServerAddress != "" {
		go webserver.
			NewWebServer(
				global.CTX,
				cmd.CFG.WebServerAddress,
				cmd.CFG.WebServerPort,
				cmd.CFG.WebserverUsername,
				cmd.CFG.WebServerPassword,
			).
			Serve()
	}
	cronInstance.Start()
	<-make(chan any)
}

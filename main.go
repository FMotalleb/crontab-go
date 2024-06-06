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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/task"
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
	lvl, _ := cmd.CFG.LogLevel.ToLogrusLevel()
	logrus.SetLevel(lvl)
	log = logger.SetupLogger("Crontab-GO")
	task.NewPost(
		&config.Task{
			Post:    "http://127.0.0.1:9085/",
			Retries: 5,
			Headers: map[string]string{
				"echo": "1",
			},
			Data: map[string]any{
				"test": true,
			},
			RetryDelay: time.Second * 2,
		},
		log,
	).Execute(context.Background())

	// j, _ := json.MarshalIndent(cmd.CFG, "", "  ")
	// fmt.Println(strings.Replace(string(j), `\n`, "\n", -1))
}

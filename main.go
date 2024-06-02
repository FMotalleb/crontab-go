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
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	cx "github.com/FMotalleb/crontab-go/context"
	"github.com/FMotalleb/crontab-go/logger"
)

var (
	log logrus.Entry
	ctx cx.Context
)

func main() {
	cmd.Execute()
	ctx = cx.NewContext("core")
	logger.InitFromConfig()
	log = *logger.SetupLogger("Crontab-GO")

	j, _ := json.Marshal(cmd.CFG)
	logrus.Infoln(string(j))
}

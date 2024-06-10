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
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/core/jobs"
	"github.com/FMotalleb/crontab-go/logger"
)

func main() {

	cli,err:=client.NewClientWithOpts(client.WithHost("unix:///run/docker.sock"))
	if(err!=nil){
		panic(err)
	}
	defer cli.Close()
	
    // Specify the container ID or name
    containerID := "alp"

    // Create an exec configuration
    execConfig := types.ExecConfig{
        AttachStdout: true,
        AttachStderr: true,
        Cmd:          []string{"/bin/sh","-c","ls -l /"},
    }

    // Create the exec instance
    exec, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
    if err != nil {
        panic(err)
    }

    // Attach to the exec instance
    resp, err := cli.ContainerExecAttach(context.Background(), exec.ID, types.ExecStartCheck{})
    if err != nil {
        panic(err)
    }
    defer resp.Close()

    // Print the command output
    _, err = io.Copy(os.Stdout, resp.Reader)
    if err != nil {
        panic(err)
    }
	return 
	cmd.Execute()
	logger.InitFromConfig()
	log := logger.SetupLogger("Crontab-GO")
	cronInstance := cron.New(cron.WithSeconds())
	log.Info("Booting up")
	jobs.InitializeJobs(log, cronInstance)
	cronInstance.Start()
	<-make(chan any)
}

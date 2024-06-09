package task

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/common"
	credential "github.com/FMotalleb/crontab-go/core/os_credential"
)

type Command struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	exe              string
	envVars          *[]string
	workingDirectory string
	log              *logrus.Entry

	user  string
	group string

	shell     string
	shellArgs []string
}

// Execute implements abstraction.Executable.
func (c *Command) Execute(ctx context.Context) error {
	r := common.GetRetry(ctx)
	log := c.log.WithField("retry", r)

	if err := c.WaitForRetry(ctx); err != nil {
		c.DoFailHooks(ctx)
		return err
	}
	ctx = common.IncreaseRetry(ctx)
	procCtx, cancel := c.ApplyTimeout(ctx)
	c.SetCancel(cancel)

	proc := exec.CommandContext(
		procCtx,
		c.shell,
		append(c.shellArgs, c.exe)...,
	)
	credential.SetUser(log, proc, c.user, c.group)
	proc.Env = *c.envVars
	proc.Dir = c.workingDirectory
	var res bytes.Buffer
	proc.Stdout = &res
	proc.Stderr = &res
	if err := proc.Start(); err != nil {
		log.Warn("failed to start the command ", err)
		return c.Execute(ctx)
	} else if err := proc.Wait(); err != nil {
		log.Warnf("command failed with answer: %s", strings.TrimSpace(res.String()))
		log.Warn("failed to execute the command", err)
		return c.Execute(ctx)
	} else {
		log.Warnf("command finished with answer: %s", strings.TrimSpace(res.String()))
	}

	if errs := c.DoDoneHooks(ctx); len(errs) != 0 {
		log.Warn("command finished successfully but its hooks failed")
	}
	return nil
}

func NewCommand(
	task *config.Task,
	logger *logrus.Entry,
) abstraction.Executable {
	log := logger.WithField("command", task.Command)
	env := os.Environ()

	shell := cmd.CFG.Shell
	shellArgs := cmd.CFG.ShellArgs
	for key, val := range task.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
		switch strings.ToLower(key) {
		case "shell":
			shell = val
		case "shell_args":
			shellArgs = strings.Split(val, ";")
		}
	}
	wd := task.WorkingDirectory
	if wd == "" {
		var e error
		wd, e = os.Getwd()
		if e != nil {
			logger.Fatalln("cannot get current working directory: ", e)
		}
	}
	cmd := &Command{
		exe:              task.Command,
		envVars:          &env,
		workingDirectory: wd,
		log: log.WithFields(
			logrus.Fields{
				"working_directory": wd,
				"shell":             shell,
				"shell_args":        shellArgs,
				"command":           task.Command,
			},
		),
		shell:     shell,
		shellArgs: shellArgs,
		user:      task.UserName,
		group:     task.GroupName,
	}
	cmd.SetMaxRetry(task.Retries)
	cmd.SetRetryDelay(task.RetryDelay)
	cmd.SetTimeout(task.Timeout)

	return cmd
}

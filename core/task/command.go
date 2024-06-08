package task

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	credential "github.com/FMotalleb/crontab-go/core/os_credential"
)

type Command struct {
	exe              string
	envVars          *[]string
	workingDirectory string
	log              *logrus.Entry
	cancel           context.CancelFunc

	user  string
	group string

	shell     string
	shellArgs []string

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration

	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

// SetDoneHooks implements abstraction.Executable.
func (c *Command) SetDoneHooks(done []abstraction.Executable) {
	c.doneHooks = done
}

// SetFailHooks implements abstraction.Executable.
func (c *Command) SetFailHooks(fail []abstraction.Executable) {
	c.failHooks = fail
}

// Cancel implements abstraction.Executable.
func (c *Command) Cancel() {
	if c.cancel != nil {
		c.log.Debugln("canceling executable")
		c.cancel()
	}
}

// Execute implements abstraction.Executable.
func (c *Command) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := c.log.WithField("retry", r)
	if getRetry(ctx) > c.retries {
		log.Warn("maximum retry reached")
		runTasks(c.failHooks)
		return fmt.Errorf("maximum retries reached")
	}
	if r != 0 {
		log.Debugln("waiting", c.retryDelay, "before executing the next iteration after last fail")
		time.Sleep(c.retryDelay)
	}
	ctx = increaseRetry(ctx)
	var procCtx context.Context
	var cancel context.CancelFunc
	if c.timeout != 0 {
		procCtx, cancel = context.WithTimeout(ctx, c.timeout)
	} else {
		procCtx, cancel = context.WithCancel(ctx)
	}
	c.cancel = cancel

	proc := exec.CommandContext(
		procCtx,
		c.shell,
		append(c.shellArgs, *&c.exe)...,
	)
	credential.SetUser(log, proc, c.user, c.group)
	proc.Env = *c.envVars
	proc.Dir = c.workingDirectory
	var res bytes.Buffer
	proc.Stdout = &res
	proc.Stderr = &res
	proc.Start()
	e = proc.Wait()
	log.Infof("command finished with answer: `%s`", strings.TrimSpace(string(res.Bytes())))
	if e != nil {
		log.Warn("failed to execute the command ", e)
		return c.Execute(ctx)
	}

	runTasks(c.doneHooks)
	return
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
	return &Command{
		exe:              task.Command,
		envVars:          &env,
		workingDirectory: wd,
		log: log.WithField(
			"working_directory", wd,
		),
		shell:      shell,
		shellArgs:  shellArgs,
		retries:    task.Retries,
		retryDelay: task.RetryDelay,
		timeout:    task.Timeout,
		user:       task.UserName,
		group:      task.GroupName,
	}
}

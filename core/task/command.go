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
)

type Command struct {
	exe              string
	envVars          *[]string
	workingDirectory string
	log              *logrus.Entry
	cancel           context.CancelFunc

	shell     string
	shellArgs []string

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration
}

// Cancel implements abstraction.Executable.
func (g *Command) Cancel() {
	if g.cancel != nil {
		g.log.Debugln("canceling executable")
		g.cancel()
	}
}

// Execute implements abstraction.Executable.
func (cmmnd *Command) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := cmmnd.log.WithField("retry", r)
	if getRetry(ctx) > cmmnd.retries {
		log.Warn("maximum retry reached")
		return fmt.Errorf("maximum retries reached")
	}
	if r != 0 {
		log.Debugln("waiting", cmmnd.retryDelay, "before executing the next iteration after last fail")
		time.Sleep(cmmnd.retryDelay)
	}
	ctx = increaseRetry(ctx)
	var procCtx context.Context
	var cancel context.CancelFunc
	if cmmnd.timeout != 0 {
		procCtx, cancel = context.WithTimeout(ctx, cmmnd.timeout)
	} else {
		procCtx, cancel = context.WithCancel(ctx)
	}
	cmmnd.cancel = cancel

	proc := exec.CommandContext(
		procCtx,
		cmmnd.shell,
		append(cmmnd.shellArgs, *&cmmnd.exe)...,
	)
	proc.Env = *cmmnd.envVars
	proc.Dir = cmmnd.workingDirectory
	var res bytes.Buffer
	proc.Stdout = &res
	proc.Stderr = &res
	proc.Start()
	e = proc.Wait()
	log.Infof("command finished with answer: `%s`", strings.TrimSpace(string(res.Bytes())))
	if e != nil {
		log.Warn("failed to execute the command ", e)
		return cmmnd.Execute(ctx)
	}
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
	}
}

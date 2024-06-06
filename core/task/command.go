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
func (g *Command) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := g.log.WithField("retry", r)
	if getRetry(ctx) > g.retries {
		log.Warn("maximum retry reached")
		return fmt.Errorf("maximum retries reached")
	}
	if r != 0 {
		log.Debugln("waiting", g.retryDelay, "before executing the next iteration after last fail")
		time.Sleep(g.retryDelay)
	}
	ctx = increaseRetry(ctx)
	var procCtx context.Context
	var cancel context.CancelFunc
	if g.timeout != 0 {
		procCtx, cancel = context.WithTimeout(ctx, g.timeout)
	} else {
		procCtx, cancel = context.WithCancel(ctx)
	}
	g.cancel = cancel
	proc := exec.CommandContext(
		procCtx,
		cmd.CFG.Shell,
		append(cmd.CFG.ShellArgs, *&g.exe)...,
	)
	proc.Env = *g.envVars
	proc.Dir = g.workingDirectory
	var res bytes.Buffer
	proc.Stdout = &res
	proc.Stderr = &res
	proc.Start()
	e = proc.Wait()
	log.Infof("command finished with answer: `%s`", strings.TrimSpace(string(res.Bytes())))
	if e != nil {
		log.Warn("failed to execute the command ", e)
		return g.Execute(ctx)
	}
	return
}

func NewCommand(
	task *config.Task,
	logger *logrus.Entry,
) abstraction.Executable {
	log := logger.WithField("command", task.Command)
	env := os.Environ()
	for key, val := range task.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
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
		retries:    task.Retries,
		retryDelay: task.RetryDelay,
		timeout:    task.Timeout,
	}
}

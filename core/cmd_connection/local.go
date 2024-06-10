package connection

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
	credential "github.com/FMotalleb/crontab-go/core/os_credential"
)

type Local struct {
	log *logrus.Entry
	cmd *exec.Cmd
}

func NewLocalCMDConn(log *logrus.Entry) abstraction.CmdConnection {
	return &Local{
		log: log.WithField(
			"connection", "local",
		),
	}
}

// Prepare implements abstraction.CmdConnection.
func (l *Local) Prepare(ctx context.Context, task *config.Task) error {
	shell := cmd.CFG.Shell
	shellArgs := cmd.CFG.ShellArgs
	env := os.Environ()
	for key, val := range task.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
		switch strings.ToLower(key) {
		case "shell":
			l.log.Info("you've used `SHELL` env variable in command environments, overriding the global shell with:", val)
			shell = val
		case "shell_args":
			l.log.Info("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: ", val)
			shellArgs = strings.Split(val, ";")
		}
	}
	workingDir := task.WorkingDirectory
	if workingDir == "" {
		var e error
		workingDir, e = os.Getwd()
		if e != nil {
			return fmt.Errorf("cannot get current working directory: %s", e)
		}
	}
	l.cmd = exec.CommandContext(
		ctx,
		shell,
		append(shellArgs, task.Command)...,
	)
	l.log = l.log.WithFields(
		logrus.Fields{
			"working_directory": workingDir,
			"shell":             shell,
			"shell_args":        shellArgs,
		},
	)
	credential.SetUser(l.log, l.cmd, task.UserName, task.GroupName)
	l.cmd.Env = env
	l.cmd.Dir = workingDir

	return nil
}

// Connect implements abstraction.CmdConnection.
func (l *Local) Connect() error {
	return nil
}

// Disconnect implements abstraction.CmdConnection.
func (l *Local) Disconnect() error {
	return nil
}

// Execute implements abstraction.CmdConnection.
func (l *Local) Execute() ([]byte, error) {
	var res bytes.Buffer
	l.cmd.Stdout = &res
	l.cmd.Stderr = &res
	if err := l.cmd.Start(); err != nil {
		l.log.Warn("failed to start the command ", err)
		return []byte{}, err
	} else if err := l.cmd.Wait(); err != nil {
		l.log.Warnf("command failed with answer: %s", strings.TrimSpace(res.String()))
		l.log.Warn("failed to execute the command", err)
		return res.Bytes(), err
	} else {
		return res.Bytes(), nil
	}
}

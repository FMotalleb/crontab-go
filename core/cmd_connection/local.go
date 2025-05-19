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
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/cmd_connection/command"
	credential "github.com/FMotalleb/crontab-go/core/os_credential"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func init() {
	cg.Register(NewLocalCMDConn)
}

// Local represents a local command connection.
type Local struct {
	log *logrus.Entry
	cmd *exec.Cmd
}

// NewLocalCMDConn creates a new instance of Local command connection.
func NewLocalCMDConn(log *logrus.Entry, cfg *config.TaskConnection) (abstraction.CmdConnection, bool) {
	if !cfg.Local {
		return nil, false
	}
	res := &Local{
		log: log.WithField(
			"connection", "local",
		),
	}
	return res, true
}

// Prepare prepares the command for execution.
// It sets up the command with the provided context, task, and environment.
// It returns an error if the preparation fails.
func (l *Local) Prepare(ctx context.Context, task *config.Task) error {
	cmdCtx := command.NewCtx(ctx, task.Env, l.log)
	workingDir := task.WorkingDirectory
	if workingDir == "" {
		var e error
		workingDir, e = os.Getwd()
		if e != nil {
			return fmt.Errorf("cannot get current working directory: %s", e)
		}
	}

	event := ctx.Value(ctxutils.EventData).([]string)
	shell, commandArg, environ := cmdCtx.BuildExecuteParams(task.Command, event)
	l.cmd = exec.CommandContext(
		ctx,
		shell,
		commandArg...,
	)
	l.log = l.log.WithFields(
		logrus.Fields{
			"working_directory": workingDir,
			"shell":             shell,
			"shell_args":        commandArg,
		},
	)
	credential.SetUser(l.log, l.cmd, task.UserName, task.GroupName)
	l.cmd.Env = environ
	l.cmd.Dir = workingDir

	// Add additional logging fields if needed
	l.log.WithFields(logrus.Fields{
		"working_directory": workingDir,
		"shell":             shell,
		"shell_args":        commandArg,
		"task":              task,
	}).Debug("command prepared")

	return nil
}

// Connect establishes the command connection.
// It returns an error if the connection cannot be established.
func (l *Local) Connect() error {
	return nil
}

// Disconnect closes the command connection.
// It returns an error if the disconnection process fails.
func (l *Local) Disconnect() error {
	return nil
}

// Execute executes the command and returns the output.
// It captures the command's standard output and standard error.
// It returns the output and an error, if any.
func (l *Local) Execute() ([]byte, error) {
	var res bytes.Buffer
	l.cmd.Stdout = &res
	l.cmd.Stderr = &res
	if err := l.cmd.Start(); err != nil {
		l.log.WithError(err).Warn("failed to start the command")
		return []byte{}, err
	} else if err := l.cmd.Wait(); err != nil {
		output := res.Bytes()
		l.log.WithError(err).WithField("output", strings.TrimSpace(res.String())).Warn("command execution failed")
		l.log.WithField("output", strings.TrimSpace(res.String())).Debug("command output")
		return output, err
	} else {
		l.log.WithField("output", strings.TrimSpace(res.String())).Debug("command output")
		return res.Bytes(), nil
	}
}

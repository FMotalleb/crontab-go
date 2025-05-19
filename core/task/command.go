// Package task provides implementation of the abstraction.Executable interface for command tasks.
package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	connection "github.com/FMotalleb/crontab-go/core/cmd_connection"
	"github.com/FMotalleb/crontab-go/core/common"
	"github.com/FMotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewCommand)
}

func NewCommand(
	logger *logrus.Entry,
	task *config.Task,
) (abstraction.Executable, bool) {
	if task.Command == "" {
		return nil, false
	}
	log := logger.WithField("command", task.Command)
	cmd := &Command{
		log: log.WithField(
			"command", task.Command,
		),
		task: task,
	}
	cmd.SetMaxRetry(task.Retries)
	cmd.SetRetryDelay(task.RetryDelay)
	cmd.SetTimeout(task.Timeout)
	cmd.SetMetaName("cmd: " + task.Command)
	return cmd, true
}

type Command struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	task *config.Task
	log  *logrus.Entry
}

// Execute implements abstraction.Executable.
func (c *Command) Execute(ctx context.Context) (e error) {
	r := common.GetRetry(ctx)
	log := c.log.WithField("retry", r)
	defer func() {
		err := recover()
		if err != nil {
			if err, ok := err.(error); ok {
				log = log.WithError(err)
				e = err
			}
			log.Warnf("recovering command execution from a fatal error: %s", err)
		}
	}()

	if err := c.WaitForRetry(ctx); err != nil {
		c.DoFailHooks(ctx)
		return err
	}

	ctx = common.IncreaseRetry(ctx)
	connections := c.task.Connections
	if fc := getFailedConnections(ctx); len(fc) != 0 {
		connections = fc
	}
	if len(connections) == 0 {
		connections = []config.TaskConnection{
			{
				Local: true,
			},
		}
		log.Debug("no explicit Connection provided using local task connection by default")
	}
	for _, conn := range connections {
		l := log.WithFields(
			logrus.Fields{
				"is-local": conn.Local,
			},
		)
		connection := connection.Get(&conn, l)
		cmdCtx, cancel := c.ApplyTimeout(ctx)
		c.SetCancel(cancel)

		if err := connection.Prepare(cmdCtx, c.task); err != nil {
			l.Warn("cannot prepare command: ", err)
			ctx = addFailedConnections(ctx, conn)
			helpers.WarnOnErrIgnored(
				l,
				connection.Disconnect,
				"Cannot disconnect the command's connection: %s",
			)
			continue
		}

		if err := connection.Connect(); err != nil {
			l.Warn("error when tried to connect, exiting current remote", err)
			ctx = addFailedConnections(ctx, conn)
			continue
		}
		ans, err := connection.Execute()
		if err != nil {
			ctx = addFailedConnections(ctx, conn)
		}
		l.Infof("command finished with answer: %s, error: %s", ans, err)
		if err := connection.Disconnect(); err != nil {
			l.Warn("error when tried to disconnect", err)
			ctx = addFailedConnections(ctx, conn)
			continue
		}
	}
	if fc := getFailedConnections(ctx); len(fc) != 0 {
		return c.Execute(ctx)
	}

	if errs := c.DoDoneHooks(ctx); len(errs) != 0 {
		log.Warn("command finished successfully but its hooks failed")
	}
	return nil
}

package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	connection "github.com/FMotalleb/crontab-go/core/cmd_connection"
	"github.com/FMotalleb/crontab-go/core/common"
)

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
		log := log.WithFields(
			logrus.Fields{
				"is-local": conn.Local,
			},
		)
		connection := connection.CompileConnection(&conn, log)
		cmdCtx, cancel := c.ApplyTimeout(ctx)
		c.SetCancel(cancel)

		if err := connection.Connect(); err != nil {
			log.Warn("error when tried to connect, exiting current remote", err)
			ctx = addFailedConnections(ctx, conn)
			continue
		}
		err := connection.Prepare(cmdCtx, c.task)
		if err != nil {
			log.Warn("cannot prepare command: ", err)
			ctx = addFailedConnections(ctx, conn)
			connection.Disconnect()
			continue
		}
		ans, err := connection.Execute()
		if err != nil {
			ctx = addFailedConnections(ctx, conn)
		}
		log.Infof("command finished with answer: %s, error: %s", ans, err)
		if err := connection.Disconnect(); err != nil {
			log.Warn("error when tried to disconnect", err)
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

func NewCommand(
	task *config.Task,
	logger *logrus.Entry,
) abstraction.Executable {
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

	return cmd
}

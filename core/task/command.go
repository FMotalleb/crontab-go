// Package task provides implementation of the abstraction.Executable interface for command tasks.
package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	connection "github.com/fmotalleb/crontab-go/core/cmd_connection"
	"github.com/fmotalleb/crontab-go/core/common"
	"github.com/fmotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewCommand)
}

func NewCommand(
	logger *zap.Logger,
	task *config.Task,
) (abstraction.Executable, bool) {
	if task.Command == "" {
		return nil, false
	}
	cmd := &Command{
		log: logger.With(
			zap.String("command", task.Command),
		),

		task: task,
	}
	cmd.ConfigRetryFrom(task)
	cmd.SetTimeout(task.Timeout)
	cmd.SetMetaName("cmd: " + task.Command)
	cmd.Action = cmd
	return cmd, true
}

type Command struct {
	common.Executable
	common.Cancelable
	common.Timeout

	task *config.Task
	log  *zap.Logger
}

// Do implements common.Action.
func (c Command) Do(ctx context.Context) (e error) {
	ctx, span := taskTracer.Start(ctx, "cmd:"+c.task.Command)
	defer span.End()
	ctx = populateVars(ctx, c.task)
	log := c.log.With(
		zap.Time("start", time.Now()),
	)
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error("panic recovered", zap.Error(err))
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				e = err
			} else {
				log.Error("panic recovered", zap.Any("error", r))
				span.RecordError(fmt.Errorf("panic: %v", r))
				span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", r))
				e = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	connections := c.task.Connections
	if len(connections) == 0 {
		connections = []config.TaskConnection{
			{
				Local: true,
			},
		}
		log.Debug("no explicit Connection provided using local task connection by default")
	}
	for _, conn := range connections {
		if err := func() error {
			l := log.With(
				zap.Any("is-local", conn.Local),
			)
			cmdConn, connErr := connection.Get(&conn, l)
			if connErr != nil {
				return connErr
			}
			cmdCtx, cancel := c.ApplyTimeout(ctx)
			defer cancel()
			c.SetCancel(cancel)

			if prepErr := cmdConn.Prepare(cmdCtx, c.task); prepErr != nil {
				l.Error("cannot prepare command", zap.Error(prepErr))
				helpers.WarnOnErrIgnored(
					l,
					cmdConn.Disconnect,
					"Cannot disconnect the command's connection",
				)
				return errors.Join(errors.New("failed to prepare"), prepErr)
			}

			if connErr = cmdConn.Connect(); connErr != nil {
				l.Error("error when tried to connect, exiting current remote", zap.Error(connErr))
				return errors.Join(errors.New("failed to connect"), connErr)
			}
			ans, execErr := cmdConn.Execute()
			if execErr != nil {
				l.Error("failed to run command", zap.Error(execErr))
				return errors.Join(errors.New("failed to execute command"), execErr)
			}
			l.Info("command finished", zap.ByteString("result", ans), zap.Error(execErr))
			if discErr := cmdConn.Disconnect(); discErr != nil {
				l.Warn("error when tried to disconnect", zap.Error(discErr))
				return nil
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

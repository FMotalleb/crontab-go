// Package command contain helper methods for cmd executors
package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/core/utils"
	"github.com/FMotalleb/crontab-go/ctxutils"
	"github.com/FMotalleb/crontab-go/template"
)

type Ctx struct {
	context.Context
	logger *logrus.Entry
}

func NewCtx(
	ctx context.Context,
	taskEnviron map[string]string,
	logger *logrus.Entry,
) *Ctx {
	result := &Ctx{
		Context: ctx,
		logger:  logger,
	}
	result.init(taskEnviron)
	return result
}

func (ctx *Ctx) init(taskEnviron map[string]string) {
	osEnviron := os.Environ()
	ctx.logger.Trace("Initial environment variables: ", osEnviron)
	ctx.Context = context.WithValue(
		ctx,
		ctxutils.Environments,
		map[string]string{},
	)
	for _, pair := range osEnviron {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			ctx.envAdd(parts[0], parts[1])
		}
	}
	for key, val := range taskEnviron {
		ctx.envAdd(key, val)
		switch strings.ToLower(key) {
		case "shell":
			ctx.logger.Info("you've used `SHELL` env variable in command environments, overriding the global shell with:", val)
		case "shell_args":
			ctx.logger.Info("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: ", val)
		}
	}
}

func (ctx *Ctx) envGetAll() map[string]string {
	if env := ctx.Value(ctxutils.Environments); env != nil {
		return env.(map[string]string)
	}
	return map[string]string{}
}

func (ctx *Ctx) envGet(key string) string {
	return ctx.envGetAll()[key]
}

func (ctx *Ctx) envAdd(key string, value string) {
	oldEnv := ctx.envGetAll()
	key = strings.ToUpper(key)
	oldEnv[key] = value
	ctx.Context = context.WithValue(
		ctx,
		ctxutils.Environments,
		oldEnv,
	)
}

func (ctx *Ctx) envReshape() []string {
	env := ctx.envGetAll()
	var result []string
	for key, val := range env {
		fKey := ctx.tryTemplate(ctx.logger, key)
		fVal := ctx.tryTemplate(ctx.logger, val)
		result = append(result, fmt.Sprintf("%s=%s", strings.ToUpper(fKey), fVal))
	}
	return result
}

func (ctx *Ctx) getShell() string {
	shell := ctx.envGet("SHELL")
	var err error
	if shell, err = ctx.applyEventTemplate(ctx.logger, shell); err != nil {
		ctx.logger.WithError(err).Warn("Failed to apply event template to shell")
	}
	return shell
}

func (ctx *Ctx) getShellArg() string {
	return ctx.envGet("SHELL_ARGS")
}

func (ctx *Ctx) BuildExecuteParams(command string) (shell string, cmd []string, env []string) {
	environments := ctx.envReshape()
	var err error
	shell = ctx.getShell()
	shellArgs := utils.EscapedSplit(ctx.getShellArg(), ':')

	c := command
	if c, err = ctx.applyEventTemplate(ctx.logger, c); err != nil {
		ctx.logger.WithError(err).Warn("Failed to apply event template to shell")
	}
	shellArgs = append(shellArgs, c)

	return shell, shellArgs, environments
}

func (ctx *Ctx) applyEventTemplate(
	log *logrus.Entry,
	src string,
) (string, error) {
	if event, ok := ctx.Value(ctxutils.EventData).(abstraction.Event); ok {
		data := event.GetData()
		return applyTemplate(log, src, data)
	}
	log.Warn("Event not found in context")
	return src, nil
}

func (ctx *Ctx) tryTemplate(
	log *logrus.Entry,
	src string,
) string {
	res, _ := ctx.applyEventTemplate(log, src)
	return res
}

func applyTemplate(
	log *logrus.Entry,
	src string,
	data map[string]any,
) (string, error) {
	res, err := template.EvaluateTemplate(src, data)
	if err == nil {
		return res, nil
	}
	log.WithError(err).Warn("Failed to apply template")
	return src, err
}

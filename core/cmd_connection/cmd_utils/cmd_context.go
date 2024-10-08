// Package cmdutils contain helper methods for cmd executors
package cmdutils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/utils"
	"github.com/FMotalleb/crontab-go/ctxutils"
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
		case "shell_arg_compatibility":
			ctx.logger.Info("you've used `SHELL_ARG_COMPATIBILITY` env variable in command environments, overriding the global shell_arg_compatibility with: ", val)
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
		result = append(result, fmt.Sprintf("%s=%s", strings.ToUpper(key), val))
	}
	return result
}

func (ctx *Ctx) getShell() string {
	return ctx.envGet("SHELL")
}

func (ctx *Ctx) getShellArg() string {
	return ctx.envGet("SHELL_ARGS")
}

func (ctx *Ctx) getShellArgCompatibility() config.ShellArgCompatibilityMode {
	result := config.ShellArgCompatibilityMode(ctx.envGet("SHELL_ARG_COMPATIBILITY"))
	switch result {
	case "":
		return config.DefaultShellArgCompatibility
	default:
		return result
	}
}

func (ctx *Ctx) BuildExecuteParams(command string, eventData []string) (shell string, cmd []string, env []string) {
	environments := ctx.envReshape()
	shell = ctx.getShell()
	shellArgs := utils.EscapedSplit(ctx.getShellArg(), ':')
	shellArgs = append(shellArgs, command)
	switch ctx.getShellArgCompatibility() {
	case config.EventArgOmit:
		ctx.logger.Debug("event arguments will not be passed to the command")
	case config.EventArgPassingAsArgs:
		shellArgs = append(shellArgs, eventData...)
	case config.EventArgPassingAsEnviron:
		environments = append(
			environments,
			fmt.Sprintf("CRONTAB_GO_EVENT_ARGUMENTS=%s",
				collectEventForEnv(eventData),
			),
		)
	default:
		ctx.logger.Warn("event argument passing mode is not supported, using default mode (omitting)")
	}
	return shell, shellArgs, environments
}

func collectEventForEnv(eventData []string) string {
	builder := &strings.Builder{}
	for i, part := range eventData {
		builder.WriteString(strings.ReplaceAll(part, ":", "\\:"))
		if i < len(eventData)-1 {
			builder.WriteRune(':')
		}
	}

	return builder.String()
}

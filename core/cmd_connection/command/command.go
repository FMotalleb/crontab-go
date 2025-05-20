// Package command contains helper methods for cmd executors
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

// NewCtx initializes a new Ctx with the provided environment and logger.
func NewCtx(ctx context.Context, taskEnviron map[string]string, logger *logrus.Entry) Ctx {
	envMap := parseEnviron(os.Environ())
	mergeEnviron(envMap, taskEnviron, logger)
	newCtx := context.WithValue(ctx, ctxutils.Environments, envMap)
	return Ctx{Context: newCtx, logger: logger}
}

func parseEnviron(environ []string) map[string]string {
	env := make(map[string]string)
	for _, pair := range environ {
		if parts := strings.SplitN(pair, "=", 2); len(parts) == 2 {
			env[strings.ToUpper(parts[0])] = parts[1]
		}
	}
	return env
}

func mergeEnviron(dest map[string]string, src map[string]string, logger *logrus.Entry) {
	for key, val := range src {
		upperKey := strings.ToUpper(key)
		dest[upperKey] = val
		switch upperKey {
		case "SHELL":
			logger.Infof("you've used `SHELL` env variable in command environments, overriding the global shell with: %s", val)
		case "SHELL_ARGS":
			logger.Infof("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: %s", val)
		}
	}
}

func (ctx Ctx) getEnv() map[string]string {
	env, ok := ctx.Value(ctxutils.Environments).(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return env
}

func (ctx Ctx) envReshape() []string {
	env := ctx.getEnv()
	result := make([]string, 0, len(env))
	for key, val := range env {
		k := strings.ToUpper(ctx.tryTemplate(key))
		v := ctx.tryTemplate(val)
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func (ctx Ctx) getShell() string {
	shell, _ := ctx.applyEventTemplate(ctx.getEnv()["SHELL"])
	if shell == "" {
		shell = "/bin/sh"
	}
	return shell
}

func (ctx Ctx) getShellArg() string {
	shellArgs, _ := ctx.applyEventTemplate(ctx.getEnv()["SHELL_ARGS"])
	if shellArgs == "" {
		shellArgs = "-c"
	}
	return shellArgs
}

// BuildExecuteParams prepares shell, args and environment for command execution.
func (ctx Ctx) BuildExecuteParams(command string) (string, []string, []string) {
	envs := ctx.envReshape()
	shell := ctx.getShell()
	shellArgs := utils.EscapedSplit(ctx.getShellArg(), ':')
	for i, v := range shellArgs {
		shellArgs[i] = ctx.tryTemplate(v)
	}
	cmd, err := ctx.applyEventTemplate(command)
	if err != nil {
		ctx.logger.WithError(err).Warn("Failed to apply event template to command")
	}
	shellArgs = append(shellArgs, cmd)
	return shell, shellArgs, envs
}

func (ctx Ctx) applyEventTemplate(src string) (string, error) {
	event, ok := ctx.Value(ctxutils.EventData).(abstraction.Event)
	if !ok {
		ctx.logger.Warn("Event not found in context")
		return src, nil
	}
	return applyTemplate(ctx.logger, src, event.GetData())
}

func (ctx Ctx) tryTemplate(src string) string {
	res, _ := ctx.applyEventTemplate(src)
	return res
}

func applyTemplate(log *logrus.Entry, src string, data map[string]any) (string, error) {
	res, err := template.EvaluateTemplate(src, data)
	if err != nil {
		log.WithError(err).Warn("Failed to apply template")
		return src, err
	}
	return res, nil
}

package connection

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
)

func reshapeEnviron(task *config.Task, log *logrus.Entry) (string, []string, []string) {
	shell := cmd.CFG.Shell
	shellArgs := cmd.CFG.ShellArgs
	env := os.Environ()
	for key, val := range task.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
		switch strings.ToLower(key) {
		case "shell":
			log.Info("you've used `SHELL` env variable in command environments, overriding the global shell with:", val)
			shell = val
		case "shell_args":
			log.Info("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: ", val)
			shellArgs = strings.Split(val, ";")
		}
	}
	return shell, shellArgs, env
}

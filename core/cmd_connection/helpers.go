package connection

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
)

// reshapeEnviron modifies the environment variables for a given task.
// It allows overriding the global shell and shell arguments with task-specific values.
//
// Parameters:
//   - task: A pointer to a config.Task struct containing the task-specific environment variables.
//   - log: A logrus.Entry used for logging information about environment variable overrides.
//
// Returns:
//   - string: The shell to be used for the task, either the global shell or the overridden shell.
//   - []string: The shell arguments to be used for the task, either the global shell arguments or the overridden shell arguments.
//   - []string: The complete set of environment variables for the task, including any task-specific overrides.
func reshapeEnviron(task *config.Task, log *logrus.Entry) (string, []string, []string) {
	shell := cmd.CFG.Shell
	shellArgs := cmd.CFG.ShellArgs
	env := os.Environ()
	log.Trace("Initial environment variables: ", env)
	for key, val := range task.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
		log.Debugf("Adding environment variable: %s=%s", key, val)
		switch strings.ToLower(key) {
		case "shell":
			log.Info("you've used `SHELL` env variable in command environments, overriding the global shell with:", val)
			shell = val
		case "shell_args":
			log.Info("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: ", val)
			shellArgs = strings.Split(val, ";")
		}
	}
	log.Trace("Final environment variables: ", env)
	return shell, shellArgs, env
}

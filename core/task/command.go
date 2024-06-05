package task

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
)

type Command struct {
	exe              string
	envVars          *[]string
	workingDirectory string
	logger           *logrus.Entry
	proc             *exec.Cmd
}

// Cancel implements abstraction.Executable.
func (g *Command) Cancel() {
	if g.proc != nil {
		g.logger.Debugln("canceling executable")
		e := g.proc.Cancel()
		if e != nil {
			g.logger.Warnln("cannot stop process", e)
		}
	}
}

// Execute implements abstraction.Executable.
func (g *Command) Execute() (e error) {
	g.Cancel()

	proc := exec.Cmd{}
	proc.Args = append(cmd.CFG.ShellArgs, *&g.exe)
	proc.Env = *g.envVars
	proc.Path = cmd.CFG.Shell
	proc.Dir = g.workingDirectory

	res, e := proc.CombinedOutput()
	g.logger.Infoln("command finished with answer:\n", string(res))
	if e != nil {
		g.logger.Warn("failed to start the command", e)
		return
	}

	return
}

func NewCommand(
	exe string,
	envVars *map[string]string,
	workingDirectory string,
	logger logrus.Entry,
) abstraction.Executable {
	env := os.Environ()
	for key, val := range *envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}
	return &Command{
		exe,
		&env,
		workingDirectory,
		logger.WithFields(
			logrus.Fields{
				"exe":               exe,
				"working_directory": workingDirectory,
			},
		),
		nil,
	}
}

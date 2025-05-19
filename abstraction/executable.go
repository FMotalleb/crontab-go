// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

// Executable is an object that can be executed using a execute method and stopped using cancel method
type Executable interface {
	Execute(context.Context) error
	SetMetaName(string)
	SetDoneHooks(context.Context, []Executable)
	SetFailHooks(context.Context, []Executable)
	Cancel()
}

type ExecutableMaker func(*logrus.Entry, *config.Task) Executable

// Package connection provides implementation of the abstraction.CmdConnection interface for command tasks.
package connection

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/generator"
)

var cg = generator.New[*config.TaskConnection, abstraction.CmdConnection]()

// Get compiles the task connection based on the provided configuration and logger.
// It returns an abstraction.CmdConnection interface based on the type of connection specified in the configuration.
func Get(conn *config.TaskConnection, logger *zap.Logger) (abstraction.CmdConnection, error) {
	con, ok := cg.Get(logger, conn)
	if ok {
		return con, nil
	}
	return nil, fmt.Errorf("cannot compile task connection: %+v", conn)
}

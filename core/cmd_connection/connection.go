// Package connection provides implementation of the abstraction.CmdConnection interface for command tasks.
package connection

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var cg = generator.New[*config.TaskConnection, abstraction.CmdConnection]()

// Get compiles the task connection based on the provided configuration and logger.
// It returns an abstraction.CmdConnection interface based on the type of connection specified in the configuration.
// If the connection type is not recognized or invalid, it logs a fatal error and returns nil.
func Get(conn *config.TaskConnection, logger *logrus.Entry) abstraction.CmdConnection {
	con, ok := cg.Get(logger, conn)
	if ok {
		return con
	}
	logger.WithField("taskConnection", conn).Error("cannot compile given taskConnection")
	return nil
}

// Package connection provides implementation of the abstraction.CmdConnection interface for command tasks.
package connection

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

// CompileConnection compiles the task connection based on the provided configuration and logger.
// It returns an abstraction.CmdConnection interface based on the type of connection specified in the configuration.
// If the connection type is not recognized or invalid, it logs a fatal error and returns nil.
func CompileConnection(conn *config.TaskConnection, logger *logrus.Entry) abstraction.CmdConnection {
	logger.Warn(conn)
	switch {
	case conn.Local:
		return NewLocalCMDConn(logger)
	case conn.ContainerName != "" && conn.ImageName == "":
		return NewDockerAttachConnection(logger, conn)
	case conn.ImageName != "":
		return NewDockerCreateConnection(logger, conn)
	}

	logger.WithField("taskConnection", conn).Error("cannot compile given taskConnection")
	return nil
}

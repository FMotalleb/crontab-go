// Package connection
package connection

import (
	"log"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

func CompileConnection(conn *config.TaskConnection, logger *logrus.Entry) abstraction.CmdConnection {
	logger.Warn(conn)
	switch {
	case conn.Local:
		return NewLocalCMDConn(logger)
	case conn.ContainerName != "":
		return NewDockerConnection(logger, conn)
	}

	log.Fatalln("cannot compile given taskConnection", conn)
	return nil
}

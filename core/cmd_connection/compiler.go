package connection

import (
	"log"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

func CompileConnection(conn *config.TaskConnection, logger *logrus.Entry) abstraction.CmdConnection {
	if conn.Local {
		return NewLocalCMDConn(logger)
	}
	log.Fatalln("cannot compile given taskConnection", conn)
	return nil
}

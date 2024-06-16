// Package mocklogger provides a logrus.Logger that does not log into stdout or stderr.
package mocklogger

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

func HijackOutput(log *logrus.Logger) (*logrus.Logger, *bytes.Buffer) {
	buffer := bytes.NewBuffer([]byte{})
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(buffer)
	return log, buffer
}

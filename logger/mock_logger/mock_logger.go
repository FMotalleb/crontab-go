package mocklogger

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

func HijackOutput(log *logrus.Logger) (*logrus.Logger, *bytes.Buffer) {
	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	return log, buffer
}

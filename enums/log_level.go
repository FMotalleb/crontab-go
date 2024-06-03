package enums

import "github.com/sirupsen/logrus"

type LogLevel string

const (
	TraceLevel = LogLevel("trace")
	DebugLevel = LogLevel("debug")
	InfoLevel  = LogLevel("info")
	WarnLevel  = LogLevel("warn")
	FatalLevel = LogLevel("fatal")
	PanicLevel = LogLevel("panic")
)

func (l LogLevel) ToLogrusLevel() (logrus.Level, error) {
	return logrus.ParseLevel(string(l))
}

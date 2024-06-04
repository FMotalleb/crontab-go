package enums

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogLevel string

const (
	TraceLevel = LogLevel("trace")
	DebugLevel = LogLevel("debug")
	InfoLevel  = LogLevel("info")
	WarnLevel  = LogLevel("warn")
	FatalLevel = LogLevel("fatal")
	PanicLevel = LogLevel("panic")
)

func (l LogLevel) Validate() error {
	switch l {
	case TraceLevel, DebugLevel, InfoLevel, WarnLevel, FatalLevel, PanicLevel:
		return nil
	default:
		return fmt.Errorf("LogLevel must be one of (trace,debug,info,warn,fatal,panic) but received `%s`", l)
	}
}

func (l LogLevel) ToLogrusLevel() (logrus.Level, error) {
	return logrus.ParseLevel(string(l))
}

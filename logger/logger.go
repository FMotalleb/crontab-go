// Package logger contains basic logging logic of the application
package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
)

type loggerType = string

var (
	jsonLogger                 = "json"
	ansiLogger                 = "ansi"
	plainLogger                = "plain"
	log         *logrus.Logger = logrus.New()
)

// SetupLogger for a section will add the section name to logger's field
func SetupLogger(section string) *logrus.Entry {
	return log.WithField("section", section)
}

// SetupLoggerOf a parent log entry
func SetupLoggerOf(parent logrus.Entry, section string) *logrus.Entry {
	parentSection := parent.Data["section"]
	sectionValue := fmt.Sprintf("%s.%s", parentSection, section)
	return parent.WithField("section", sectionValue)
}

// InitFromConfig parsed using cmd.Execute()
func InitFromConfig() {
	log = logrus.New()
	switch cmd.Config.Log.Format {
	case jsonLogger:
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: cmd.Config.Log.TimeStampFormat,
		}
	case ansiLogger:
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     true,
			TimestampFormat: cmd.Config.Log.TimeStampFormat,
		}
	}
}

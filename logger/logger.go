// Package logger contains basic logging logic of the application
package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/enums"
)

var log *logrus.Logger = logrus.New()

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
	if err := cmd.CFG.LogFormat.Validate(); err != nil {
		log.Fatal(err)
	}
	switch cmd.CFG.LogFormat {
	case enums.JsonLogger:
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	case enums.AnsiLogger:
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     true,
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	case enums.PlainLogger:
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     false,
			DisableColors:   true,
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	}
}

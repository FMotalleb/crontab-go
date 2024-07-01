// Package logger contains basic logging logic of the application
package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/enums"
)

var log *logrus.Logger = logrus.New()

// SetupLogger for a section will add the section name to logger's field
func SetupLogger(section string) (l *logrus.Entry) {
	return log.WithField("section", section)
}

// SetupLoggerOf a parent log entry
func SetupLoggerOf(parent logrus.Entry, section string) *logrus.Entry {
	parentSection := parent.Data["section"]
	sectionValue := fmt.Sprintf("%s.%s", parentSection, section)
	return parent.WithField("section", sectionValue)
}

func AddHook(hook logrus.Hook) {
	log.AddHook(hook)
}

// InitFromConfig parsed using cmd.Execute()
func InitFromConfig() {
	log = logrus.New()
	wr := &writer{
		stdout: cmd.CFG.LogStdout,
	}
	if cmd.CFG.LogFile != "" {
		var err error

		wr.file, err = os.OpenFile(cmd.CFG.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.SetOutput(wr)
	if err := cmd.CFG.LogFormat.Validate(); err != nil {
		log.Fatal(err)
	}
	log.Level, _ = cmd.CFG.LogLevel.ToLogrusLevel()

	switch cmd.CFG.LogFormat {
	case enums.JSONLogger:
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	case enums.AnsiLogger:
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	case enums.PlainLogger:
		log.Formatter = &logrus.TextFormatter{
			ForceColors:     false,
			DisableColors:   true,
			TimestampFormat: cmd.CFG.LogTimestampFormat,
		}
	}

	logrus.SetLevel(log.Level)
	log.ReportCaller = log.IsLevelEnabled(logrus.TraceLevel)
	logrus.SetFormatter(log.Formatter)
	logrus.SetOutput(wr)
}

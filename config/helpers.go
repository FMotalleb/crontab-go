package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	shouldLog        = true
	internalLogLevel = -1
)

func (cfg *Config) debugLog(text string, opts ...any) {
	if internalLogLevel == -1 {
		lvl, err := cfg.LogLevel.ToLogrusLevel()
		if err != nil {
			internalLogLevel = 5
		} else {
			internalLogLevel = int(lvl)
			if internalLogLevel < int(logrus.DebugLevel) {
				shouldLog = false
			}
		}
	}
	if shouldLog {
		fmt.Printf("[LOADER-LOG] %s\n", fmt.Sprintf(text, opts...))
	}
}

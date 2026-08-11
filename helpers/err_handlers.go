// Package helpers provides helper functions.
package helpers

import "go.uber.org/zap"

func WarnOnErrIgnored(log *zap.Logger, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Warn(message, zap.Error(err))
	}
}

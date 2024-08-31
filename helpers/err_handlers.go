// Package helpers provides helper functions.
package helpers

import (
	"github.com/sirupsen/logrus"
)

func PanicOnErr(log *logrus.Entry, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Panicf(message, err)
	}
}

func FatalOnErr(log *logrus.Entry, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Fatalf(message, err)
	}
}

func WarnOnErr(log *logrus.Entry, errorCatcher func() error, message string) error {
	if err := errorCatcher(); err != nil {
		log.Warnf(message, err)
		return err
	}
	return nil
}

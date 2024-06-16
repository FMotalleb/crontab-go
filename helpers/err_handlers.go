// Package helpers provides helper functions.
package helpers

import (
	"github.com/sirupsen/logrus"
)

func PanicOnErr(log *logrus.Entry, err error, message string) {
	if err != nil {
		log.Panicf(message, err)
	}
}

func FatalOnErr(log *logrus.Entry, err error, message string) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func WarnOnErr(log *logrus.Entry, err error, message string) {
	if err != nil {
		log.Warnf(message, err)
	}
}

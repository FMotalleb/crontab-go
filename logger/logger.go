// Package logger contains basic logging logic of the application
package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func SetupLogger(parent logrus.Entry, section string) *logrus.Entry {
	parentSection := parent.Data["section"]
	sectionValue := fmt.Sprintf("%s/%s", parentSection, section)
	return parent.WithField("section", sectionValue)
}

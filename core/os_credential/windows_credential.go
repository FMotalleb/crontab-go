//go:build windows
// +build windows

package credential

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Validate(log *logrus.Entry, usr string, grp string) error {
	log.Warn("windows os does not have capability to set user thus validation will pass but will not work")
	return nil
}

func SetUser(log *logrus.Entry, _ *exec.Cmd, _ string, _ string) {
	log.Warn("cannot set user in windows platform")
}

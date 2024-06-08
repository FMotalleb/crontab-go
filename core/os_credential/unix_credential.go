//go:build !windows
// +build !windows

package credential

import (
	"errors"
	"fmt"
	"os/exec"
	osUser "os/user"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

func Validate(log *logrus.Entry, usr string, grp string) error {
	cu, err := osUser.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user error: %s", err)
	}
	if usr != "" && cu.Uid != "0" {
		return errors.New("cannot switch user of tasks without root privilege, if you need to use user in tasks run crontab-go as user root")
	}
	_, _, err = lookupUIDAndGID(usr, log)
	if err != nil {
		return fmt.Errorf("cannot get uid and gid of user `%s` error: %s", usr, err)
	}
	_, err = lookupGID(grp, log)
	if err != nil {
		return fmt.Errorf("cannot get gid of group `%s` error: %s", grp, err)
	}

	return nil
}

func SetUser(log *logrus.Entry, proc *exec.Cmd, usr string, grp string) {
	if usr == "" {
		log.Trace("no username given, running as current user")
		return
	}

	uid, gid, err := lookupUIDAndGID(usr, log)
	if err != nil {
		log.Panicf("cannot get uid and gid of user %s, error: %s", usr, err)
	}
	if grp != "" {
		gid, _ = lookupGID(grp, log)
	}

	setUID(log, proc, uint32(uid), uint32(gid))
}

func lookupGID(grp string, log *logrus.Entry) (gid int, err error) {
	g, err := osUser.LookupGroup(grp)
	if err != nil {
		log.Panicf("cannot find group with name %s in the os: %s, you've changed os users during application runtime", grp, err)
	}
	gid, err = strconv.Atoi(g.Gid)
	return
}

func lookupUIDAndGID(usr string, log *logrus.Entry) (uid int, gid int, err error) {
	u, err := osUser.Lookup(usr)
	if err != nil {
		log.Panicf("cannot find user with name %s in the os: %s, you've changed os users during application runtime", usr, err)
	}
	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return
	}
	gid, err = strconv.Atoi(u.Gid)
	return
}

func setUID(
	log *logrus.Entry,
	proc *exec.Cmd,
	uid uint32,
	gid uint32,
) {
	log.Tracef("Setting: uid(%d) and gid(%d)", uid, gid)
	attrib := &syscall.SysProcAttr{}
	proc.SysProcAttr = attrib
	proc.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}
}

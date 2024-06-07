//go:build windows
// +build windows

package task

import "os/exec"

func (g *Command) setUid(_ *exec.Cmd, uid uint, gid uint) {
	g.log.Warnf("Windows platform does not support setting uid(%d) and gid(%d) for the process", uid, gid)
}

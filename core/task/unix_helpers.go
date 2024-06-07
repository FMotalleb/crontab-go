//go:build !windows
// +build !windows

package task

import "os/exec"

func (g *Command) setUid(_ *exec.Cmd, uid uint, gid uint) {
	g.log.Warnf("Todo: uid(%d) and gid(%d)", uid, gid)
}

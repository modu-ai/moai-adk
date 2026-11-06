//go:build !windows

package cli

import (
	"errors"
	"os/exec"
	"syscall"
)

// isolateCodexProcessGroup starts the codex child in a process group of its own,
// so a group signal reaches the helpers it started and nothing else.
func isolateCodexProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// codexGroupAlive reports whether the group led by pgid still has a member. A
// permission error still means a member exists.
func codexGroupAlive(pgid int) bool {
	if pgid <= 0 {
		return false
	}
	err := syscall.Kill(-pgid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// signalCodexGroup sends SIGTERM, or SIGKILL when kill is set, to the group. The
// only expected error is "no such process": the group is already empty.
func signalCodexGroup(pgid int, kill bool) {
	if pgid <= 0 {
		return
	}
	sig := syscall.SIGTERM
	if kill {
		sig = syscall.SIGKILL
	}
	_ = syscall.Kill(-pgid, sig)
}

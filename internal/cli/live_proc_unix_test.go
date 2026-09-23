//go:build !windows

package cli

import (
	"errors"
	"os/exec"
	"syscall"
)

// setLiveProcGroup places the process in its own process group so the whole
// tree it spawns can be killed at once.
func setLiveProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killLiveProcGroup(pid int) error {
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

// liveProcGroupGone reports whether no process of the group remains.
func liveProcGroupGone(pid int) bool {
	return errors.Is(syscall.Kill(-pid, 0), syscall.ESRCH)
}

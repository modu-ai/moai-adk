//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

// afGateProcessGroup puts the gate in its own process group and makes a cancelled gate kill the whole group, so a
// grandchild the gate started (a test server, a watcher) does not outlive its timeout.
func afGateProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

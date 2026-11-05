//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

// afGateKillGroup kills the gate's process group after the gate returned, whatever it returned: a child the gate
// left running must not outlive the command. A group that is already gone is not an error.
func afGateKillGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

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

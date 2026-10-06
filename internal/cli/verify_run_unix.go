//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

// verifyRunPrepare starts the command in its own process group and makes a
// timeout terminate the whole group, so descendants (a `sh -c` wrapper's
// `go test` binary) do not outlive it as background load.
func verifyRunPrepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

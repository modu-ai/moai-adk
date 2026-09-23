//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// Windows has no POSIX process groups; the LIVE tests are not run there and
// these fall back to single-process handling.
func setLiveProcGroup(cmd *exec.Cmd) {}

func killLiveProcGroup(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	_ = p.Kill()
	return nil
}

func liveProcGroupGone(pid int) bool { return true }

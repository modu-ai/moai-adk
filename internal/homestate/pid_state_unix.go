//go:build !windows

package homestate

import (
	"errors"
	"os"
	"syscall"
)

func platformPIDState(pid int) ProcessIdentityState {
	if pid <= 0 {
		return ProcessIdentityDead
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return ProcessIdentityIndeterminate
	}
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		return ProcessIdentityLive
	}
	if errors.Is(err, syscall.ESRCH) {
		return ProcessIdentityDead
	}
	return ProcessIdentityIndeterminate
}

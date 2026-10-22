//go:build !windows

package atomicfile

import (
	"errors"
	"os"
	"syscall"
)

// PidAlive observes Unix signal 0. Only an absent/finished process is dead;
// permission and other unknown failures retain the possibly-live owner's lock.
func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false // no such process on this platform's FindProcess
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		if errors.Is(err, os.ErrProcessDone) || err == syscall.ESRCH {
			return false
		}
		// EPERM (a live process owned by someone else) and any other error
		// read as alive — the conservative direction.
		return true
	}
	return true
}

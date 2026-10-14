//go:build !windows

package toolpolicy

import (
	"errors"
	"os"
	"syscall"
)

// pidProvablyDead reports whether the recorded lockfile holder is provably
// dead (card t1630 sync-audit F2'): a non-positive pid, an unresolvable
// process, or a process whose signal-0 probe reports it done or gone. A live
// process we cannot signal (EPERM) is treated as alive — never taken over.
// ESRCH is matched explicitly because a fresh FindProcess handle has not
// observed the exit, so the os package does not translate kill(2)'s ESRCH
// into os.ErrProcessDone.
func pidProvablyDead(pid int) bool {
	if pid <= 0 {
		return true
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return true
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
	}
	return false
}

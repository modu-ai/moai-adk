//go:build !windows

package hygiene

import (
	"errors"
	"syscall"
)

// probePidAlive classifies a registry pid's liveness signal: a process
// that answers signal 0 is alive; a gone process (ESRCH) is a measured
// negative; an unreadable process (EPERM — a foreign-user owner) is still
// alive as far as exclusion goes. The fingerprint half of an affirmative
// signal needs a recorded fingerprint, which this install does not store,
// so an alive pid reads unmeasured at the caller (fail-closed, REQ-HYG-007).
func probePidAlive(pid int) Signal {
	if pid <= 0 {
		return SignalUnmeasured
	}
	err := syscall.Kill(pid, 0)
	switch {
	case err == nil:
		return SignalUnmeasured // alive; no recorded fingerprint to compare
	case errors.Is(err, syscall.ESRCH):
		return SignalNegative
	case errors.Is(err, syscall.EPERM):
		return SignalUnmeasured // alive under another account
	default:
		return SignalUnmeasured
	}
}

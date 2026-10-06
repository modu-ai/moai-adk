//go:build windows

package hygiene

import "golang.org/x/sys/windows"

// probePidAlive classifies a registry pid's liveness signal on Windows: an
// openable process is alive; a gone process is a measured negative; an
// unreadable one reads unmeasured (fail-closed, REQ-HYG-007).
func probePidAlive(pid int) Signal {
	if pid <= 0 {
		return SignalUnmeasured
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if errno, ok := err.(windows.Errno); ok && errno == windows.ERROR_INVALID_PARAMETER {
			return SignalNegative
		}
		return SignalUnmeasured
	}
	_ = windows.CloseHandle(h)
	return SignalUnmeasured // alive; no recorded fingerprint to compare
}

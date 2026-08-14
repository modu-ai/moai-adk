//go:build windows

// lock_alive_windows.go — the default liveness probe for the stale-lock
// clear (REQ-KB-023) on Windows.
//
// The earlier form ended in proc.Signal(syscall.Signal(0)), which the stdlib
// supports for Kill only — Signal returns EWINDOWS for every other signal,
// so the probe reported DEAD for every pid and the clear's live-owner
// refusal was dead code: it would unlink a LIVE holder's lock while
// reporting "terminated pid" (sync-audit F1). The Windows-valid shape is
// OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION) + GetExitCodeProcess, with
// STILL_ACTIVE (259) meaning live.
//
// Indeterminate observations resolve to LIVE — guessing "dead" is precisely
// what unlinks a live holder's lock. The one positively-dead signal is
// ERROR_INVALID_PARAMETER from OpenProcess (no such process); access-denied
// and every other open failure mean the process MAY be alive (other session,
// protected process), so they read as live.
package kanban

import (
	"golang.org/x/sys/windows"
)

// stillActiveExitCode is Windows' STILL_ACTIVE sentinel: the exit code a
// handle reports while the process has not terminated.
const stillActiveExitCode = 259

// defaultProcessAlive reports whether pid names a live process on Windows.
func defaultProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if err == windows.ERROR_INVALID_PARAMETER {
			// No such process — positively dead.
			return false
		}
		// Access denied / protected process / anything else: the process may
		// well be alive. Indeterminate resolves to LIVE.
		return true
	}
	defer func() { _ = windows.CloseHandle(h) }()

	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		// Could not read the exit code: indeterminate, resolve to LIVE.
		return true
	}
	return code == stillActiveExitCode
}

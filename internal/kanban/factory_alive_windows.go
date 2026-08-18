//go:build windows

package kanban

// factory_alive_windows.go — the Windows liveness probe behind the factory
// worker-name registry (SPEC-FACTORY-WORKER-FANOUT-001). It is the
// Windows-valid shape internal/kanban/lock_alive_windows.go established
// (sync-audit F1 there): the stdlib's proc.Signal(syscall.Signal(0)) returns
// EWINDOWS for anything but Kill, so the probe is
// OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION) + GetExitCodeProcess, with
// STILL_ACTIVE (259) meaning live.
//
// Indeterminate observations resolve to LIVE — guessing "dead" is precisely
// what would let two sessions share one worker name. The one positively-dead
// signal is ERROR_INVALID_PARAMETER from OpenProcess (no such process);
// access-denied and every other open failure mean the process MAY be alive.
//
// Moved from internal/cli (t85 lead loop) unchanged in behavior; the cli
// package keeps a package var seam over it.

import (
	"golang.org/x/sys/windows"
)

// FactoryProcessAlive reports whether pid names a live process on Windows.
// pid <= 0 is never live. The STILL_ACTIVE sentinel is lock_alive_windows.go's
// stillActiveExitCode, shared rather than restated.
func FactoryProcessAlive(pid int) bool {
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

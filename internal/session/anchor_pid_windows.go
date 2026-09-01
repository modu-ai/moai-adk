//go:build windows

// anchor_pid_windows.go — process-liveness probe for the anchor guard.
//
// The historical stub reported every PID as alive; t426 (windows census
// axis 2a) replaces it with a real OpenProcess + GetExitCodeProcess probe so
// mcp_server_runtime's liveness split, the doctor's dead-record pruning, and
// the update-cleanup stale-lock check all observe what unix observes. The
// fail-closed direction of the anchor guard is unchanged: probeProcessLiveness
// keeps "I do not know" distinguishable from "dead".
package session

import (
	"errors"

	"golang.org/x/sys/windows"
)

// stillActive is the exit code GetExitCodeProcess reports for a process that
// has not yet terminated (WAIT_TIMEOUT / STATUS_PENDING = 259). Declared
// locally rather than via the x/sys constant so the value is pinned against
// any module version.
const stillActive = 259

// probeProcessLiveness is the two-valued liveness probe backing the shared
// anchor decision (SPEC-WORKTREE-REAPER-001 REQ-WR-008, design.md §B.5).
//
//	nil    → (true, true)   the process exists (including another owner's)
//	ESRCH  → (false, true)  it is positively gone
//	other  → (false, false) UNDETERMINED — the caller must fail closed
//
// The unix EPERM analogue is ERROR_ACCESS_DENIED: the process exists but
// another owner holds it, so it is alive. GetExitCodeProcess failing is the
// one shape this platform cannot judge.
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return true, true
		}
		return false, true
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false, false
	}
	return code == stillActive, true
}

// isProcessAlive reports whether a process with the given PID is running.
func isProcessAlive(pid int) bool {
	alive, _ := probeProcessLiveness(pid)
	return alive
}

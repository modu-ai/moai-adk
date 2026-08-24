//go:build windows

// anchor_pid_windows.go — conservative process-liveness probe for the
// anchor guard, mirroring internal/cli/update_cleanup_windows.go: it
// reports every PID as alive so the guard fails toward protecting a
// possibly-live session. Stale entries age out via the
// DefaultStaleMinutes heartbeat floor instead.
package session

// isProcessAlive is conservative on Windows: always true. Proper
// Windows-native PID validation (OpenProcess + GetExitCodeProcess) is
// deferred alongside the internal/cli twin — see the follow-up noted in
// update_cleanup_windows.go.
func isProcessAlive(pid int) bool {
	_ = pid
	return true
}

// probeProcessLiveness is the two-valued liveness probe backing the shared
// anchor decision (SPEC-WORKTREE-REAPER-001 REQ-WR-008, design.md §B.5).
// Windows has no probe here at all — nothing about the process is observed —
// so it reports the result as UNDETERMINED rather than as definitively alive.
//
// The guard's outcome is identical either way: lockAnchorVerdict treats
// undetermined as anchored, so the tree is preserved on both spellings. What
// changes is truthfulness. `determined=true` asserts a fact the platform did
// not measure, and the notice downstream then reads "locked by live pid N" for
// a pid nothing ever checked. `(false, false)` keeps the fail-closed direction
// and makes the notice say what was actually known.
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	_ = pid
	return false, false
}

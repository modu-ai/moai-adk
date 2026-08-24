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
// Windows cannot assert death here — the underlying probe is unconditionally
// conservative — so it reports (true, true): existing, definitively. Reporting
// (false, ...) would widen removal on the one platform with no real probe.
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	_ = pid
	return true, true
}

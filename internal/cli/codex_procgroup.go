package cli

import "time"

// Issue #1948 (CLI package cleanup failure). A codex app-server starts helpers of
// its own (a plugin clone writes under CODEX_HOME) and the session's close used to
// wait for the direct child only, so a helper could still be writing when the
// caller removed the directory it wrote to. The session now runs in a process
// group of its own (pgid = the child's pid, so the group is ours and only ours),
// and close does not return while that group has members, within a fixed bound.

const (
	// codexGroupTermWait is how long close waits for the group to empty after
	// SIGTERM, and codexGroupKillWait how long after SIGKILL. Together they cap the
	// time close adds once the direct child is gone.
	codexGroupTermWait = time.Second
	codexGroupKillWait = time.Second

	codexGroupPoll = 10 * time.Millisecond
)

// reapCodexGroup ends whatever is still running in the session's process group.
// pgid is the pid of the child close started; a zero or negative value (a child
// that never started, or a platform without groups) reaps nothing. Only the
// group created for this session is signalled — never a name or a pattern.
func reapCodexGroup(pgid int) {
	if !codexGroupAlive(pgid) {
		return
	}
	signalCodexGroup(pgid, false)
	if waitCodexGroupGone(pgid, codexGroupTermWait) {
		return
	}
	signalCodexGroup(pgid, true)
	waitCodexGroupGone(pgid, codexGroupKillWait)
}

// waitCodexGroupGone polls until the group has no member or d has passed.
func waitCodexGroupGone(pgid int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for codexGroupAlive(pgid) {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(codexGroupPoll)
	}
	return true
}

package statusline

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/stateanchor"
)

// resolveStateAnchor resolves the project root that anchors every state read
// and write of a render (SPEC-STATE-ANCHOR-001 REQ-SA-001): the telemetry
// write (B1), the board root with its landed and github-counts consumers
// (B2/B2b), and the goal-state read (B3). It is the statusline adapter over
// the single shared seam — the precedence chain itself is fixed by
// REQ-SA-002 and lives in internal/stateanchor.
//
// The session's current directory (and the legacy cwd field) feeds only the
// git walk-up inside the seam; it is never the anchor itself. That is the
// GH #1694 repair: a session that cd'd around used to leave one telemetry
// record per visited directory. Display-name derivation stays with
// extractProjectDirectory, untouched (REQ-SA-004).
//
// An empty return means no project: callers skip the state write or read and
// the render completes normally (REQ-SA-003).
func resolveStateAnchor(input *StdinData) string {
	var s stateanchor.Session
	if input != nil {
		if input.Workspace != nil {
			s.ProjectDir = input.Workspace.ProjectDir
			s.CurrentDir = input.Workspace.CurrentDir
		}
		if input.Worktree != nil {
			s.OriginalCwd = input.Worktree.OriginalCwd
		}
		s.CWD = input.CWD
	}
	return stateanchor.Resolve(s)
}

// resolveSessionDir is the PRE-REPAIR session-directory chain (formerly
// resolveProjectDir: current_dir → legacy CWD → os.Getwd). It is retained
// only for the members whose flip to the state anchor is sequenced behind
// their own observed RED (plan §D8): B2 board root and B3 goal read flip in
// M2, B4 in M3. It is NOT an anchor — B1 (the telemetry write) already goes
// through resolveStateAnchor. Deleted once the last member flips (AC-SA-012).
func resolveSessionDir(input *StdinData) string {
	if input != nil {
		if input.Workspace != nil && input.Workspace.CurrentDir != "" {
			return input.Workspace.CurrentDir
		}
		if input.CWD != "" {
			return input.CWD
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

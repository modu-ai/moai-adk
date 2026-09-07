// Package stateanchor resolves the single state anchor — the project root
// under which .moai/state/ is read and written (SPEC-STATE-ANCHOR-001
// REQ-SA-001). Before this seam every state surface derived its own anchor
// from wherever the session happened to be: the statusline telemetry write
// anchored to workspace.current_dir, so a session that cd'd around left one
// stray .moai directory behind per visited directory (GH #1694: 226 stray
// .moai dirs in the reporter's project). Every member now resolves through
// this one seam: the statusline telemetry write (B1), the board root with its
// landed and github-counts consumers (B2/B2b), the goal-state read (B3), and
// the CLI config-cache chain (B4).
//
// The precedence chain is fixed by REQ-SA-002 (plan D1 — inserting steps or
// reordering is a requirements change): stdin workspace.project_dir first,
// then worktree.original_cwd, then a git resolution of the session's
// directory reusing gitcore.ResolveGitDirs — the git common directory's
// parent, which is one root for every checkout and worktree of the
// repository (the internal/kanban primaryCheckoutRoot shape). When nothing
// resolves, the anchor is "": callers skip the state write or read and the
// render completes normally (REQ-SA-003 — no project, no state).
//
// The session's current directory is never the anchor itself; it only feeds
// the git walk-up. Display-name derivation is a separate concern and stays
// with the statusline's extractProjectDirectory (REQ-SA-004).
package stateanchor

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// Session carries the anchor-relevant fields of a session context — the
// subset of the statusline stdin payload the chain reads. All fields are
// optional; an empty field simply falls through to the next chain step.
type Session struct {
	// ProjectDir is workspace.project_dir — the project root the runtime
	// reports for the session (chain step 1).
	ProjectDir string

	// OriginalCwd is worktree.original_cwd — the primary checkout behind a
	// worktree session (chain step 2).
	OriginalCwd string

	// CurrentDir is workspace.current_dir — where the session is right now.
	// Feeds only the git walk-up (chain step 3), never the anchor itself.
	CurrentDir string

	// CWD is the legacy top-level cwd field, used when CurrentDir is absent.
	CWD string
}

// Resolve returns the state anchor for a session context, following the
// fixed REQ-SA-002 precedence: ProjectDir → OriginalCwd → a git resolution
// of CurrentDir (or CWD). Returns "" when nothing resolves — callers skip
// the state write and carry on (REQ-SA-003). A context with no directory
// fields at all resolves to "": the process working directory is
// deliberately NOT consulted, so a render whose payload carries no location
// can never reach the operator's real checkout through the resolver —
// no directory context, no anchor.
//
// @MX:ANCHOR: [AUTO] single state-anchor seam — every state read/write member resolves through here
// @MX:REASON: SPEC-STATE-ANCHOR-001 REQ-SA-001/002; the fixed precedence chain is the repair for GH #1694 and member-specific anchors are the defect being removed
// @MX:SPEC: SPEC-STATE-ANCHOR-001
func Resolve(s Session) string {
	if s.ProjectDir != "" {
		return s.ProjectDir
	}
	if s.OriginalCwd != "" {
		return s.OriginalCwd
	}
	dir := s.CurrentDir
	if dir == "" {
		dir = s.CWD
	}
	return FromDirectory(dir)
}

// FromDirectory resolves the anchor from a directory alone — the CLI chain's
// shape, where no session payload exists. The resolution is a function of the
// REPOSITORY, never of the caller's location: the git common directory is
// shared by every checkout, and its parent is the primary checkout's root.
// Returns "" when the directory is not inside a git repository (or git
// cannot answer) — the caller's REQ-SA-003 skip path.
//
// Cost note: this spawns one `git rev-parse` per call. Session-context
// callers (statusline) reach it only when neither stdin field resolved; the
// CLI config chain calls it once per process. Both are cold paths compared
// to the render's existing git status collection.
func FromDirectory(dir string) string {
	if dir == "" {
		return ""
	}
	dirs, err := git.ResolveGitDirs(dir)
	if err != nil || dirs == nil || dirs.CommonDir == "" {
		return ""
	}
	return filepath.Dir(dirs.CommonDir)
}

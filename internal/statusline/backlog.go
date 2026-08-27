package statusline

import (
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// BacklogCounts is the kanban backlog reduced to what a glance needs: how much
// work is in flight and how much is waiting. Dropped items are deliberately not
// counted — they are history, and a number that only ever grows is noise.
type BacklogCounts struct {
	Picked    int  // state == "picked" — admitted to the board, in flight
	Queued    int  // state == "queued" — waiting for the operator to pick
	Available bool // false when no backlog file could be read
}

// resolveBoardRoot returns the directory holding the project's kanban state.
//
// This is NOT always the session's working directory. `.moai/state/` is
// gitignored, so it exists only in the primary checkout — a session working
// inside a worktree finds nothing there. Claude Code hands us the primary's
// path as `worktree.original_cwd`, so the worktree case is resolved from the
// payload rather than by shelling out to git, which would put a subprocess on
// every status render.
func resolveBoardRoot(input *StdinData) string {
	if input != nil && input.Worktree != nil && input.Worktree.OriginalCwd != "" {
		return input.Worktree.OriginalCwd
	}
	return resolveProjectDir(input)
}

// resolveBacklogCounts counts the backlog by state under boardRoot.
//
// It delegates to kanban.BacklogCountsForRoot rather than reading a file
// itself: the queue's storage is a database now, and a second reader here
// would have to be kept in step with the store by hand — the exact drift the
// single-seam rule exists to prevent. The two properties this render depends
// on are the shared helper's, and both are load-bearing: PURE (a status render
// must never perform the one-time storage cutover or directory relocation) and
// fail-open (an unreadable queue renders nothing rather than an
// authoritative-looking zero).
//
// Constant-cost per render, on either layout: it must never grow with the
// number of cards in a way that puts the render on a slow path.
func resolveBacklogCounts(boardRoot string) BacklogCounts {
	c := kanban.BacklogCountsForRoot(boardRoot)
	return BacklogCounts{Picked: c.Picked, Queued: c.Queued, Available: c.Available}
}

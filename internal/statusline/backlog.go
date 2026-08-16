package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// BacklogCounts is the kanban backlog reduced to what a glance needs: how much
// work is in flight and how much is waiting. Dropped items are deliberately not
// counted — they are history, and a number that only ever grows is noise.
type BacklogCounts struct {
	Picked    int  // state == "picked" — admitted to the board, in flight
	Queued    int  // state == "queued" — waiting for the operator to pick
	Available bool // false when no backlog file could be read
}

// backlogFileSnippet is the minimal subset of backlog.json the statusline
// reads. Only the per-item state is load-bearing.
type backlogFileSnippet struct {
	Items []struct {
		State string `json:"state"`
	} `json:"items"`
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

// resolveBacklogCounts reads .moai/state/kanban/backlog.json under boardRoot and
// counts items by state. Best-effort + fail-open: an absent file, a read error,
// or a parse error all yield Available=false and render nothing. Constant-cost
// (one read of one small file) per statusline render — it must never grow with
// the number of cards in a way that puts the render on a slow path.
func resolveBacklogCounts(boardRoot string) BacklogCounts {
	if boardRoot == "" {
		return BacklogCounts{}
	}
	path := filepath.Join(boardRoot, ".moai", "state", "kanban", "backlog.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return BacklogCounts{} // absent or unreadable → nothing to show
	}
	var s backlogFileSnippet
	if err := json.Unmarshal(data, &s); err != nil {
		return BacklogCounts{} // corrupt → nothing to show (fail-open)
	}

	counts := BacklogCounts{Available: true}
	for _, item := range s.Items {
		switch item.State {
		case "picked":
			counts.Picked++
		case "queued":
			counts.Queued++
		}
	}
	return counts
}

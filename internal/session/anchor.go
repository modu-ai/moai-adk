// anchor.go — live-session anchor detection for worktree disposal.
//
// Backs the `moai worktree done` anchor guard: before removing a worktree,
// the CLI asks whether a live session is still anchored inside it. Removing
// the tree under a live session kills that session's shell — Claude Code's
// native worktree-isolation guard then refuses every Bash call whose working
// directory can no longer resolve inside the (deleted) tree.
package session

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LiveAnchoredSessions returns registry entries that describe LIVE sessions
// anchored inside treePath: registered in the tree-LOCAL registry at
// <treePath>/.moai/state/active-sessions.json with a working directory at
// or below treePath, on this host, whose owning process is still running.
//
// Why the tree-LOCAL registry: hooks resolve the registry path from the
// session's own project dir (CLAUDE_PROJECT_DIR), so a lane launched via
// `moai cc -w <name>` registers inside the worktree, not in the primary
// checkout. A disposal invoked from the primary therefore only sees the
// lane when it reads the target tree's registry.
//
// Liveness is judged from the PID probe, with a heartbeat-freshness floor
// (DefaultStaleMinutes) as the conservative fallback for platforms where
// the probe cannot prove death. The heartbeat alone is NOT sufficient: no
// per-turn heartbeat driver exists today, so a long-running live lane
// carries a heartbeat hours old — the PID is what distinguishes it from a
// zombie.
//
// Fail-open: an absent or unreadable registry returns nil (disposal must
// not wedge on a corrupt registry file).
//
// Known limitation: a session that started elsewhere and entered the tree
// mid-session (Claude Code's EnterWorktree) keeps its original registry
// entry — its CWD field never updates on CwdChanged — so it is invisible
// here. Only sessions that STARTED inside the tree are detected.
func LiveAnchoredSessions(treePath string, now time.Time) []Entry {
	if treePath == "" {
		return nil
	}
	reg := NewRegistry(filepath.Join(treePath, DefaultRegistryPath), nil)
	entries, err := reg.Query("")
	if err != nil {
		return nil
	}
	host, _ := os.Hostname()
	stale := DefaultStaleMinutes * time.Minute
	var anchored []Entry
	for _, e := range entries {
		if e.Host != "" && e.Host != host {
			continue
		}
		if !pathWithinTree(e.CWD, treePath) {
			continue
		}
		alive := e.PID > 0 && isProcessAlive(e.PID)
		if alive || now.Sub(e.LastHeartbeat) < stale {
			anchored = append(anchored, e)
		}
	}
	return anchored
}

// pathWithinTree reports whether the recorded working directory child lies
// at or below root. Compares cleaned literal paths first, then best-effort
// symlink-resolved forms (the session may have been launched via an alias
// of the tree path). EvalSymlinks requires the path to exist; a missing
// recorded cwd falls back to the literal comparison above.
func pathWithinTree(child, root string) bool {
	if child == "" {
		return false
	}
	if underPath(child, root) {
		return true
	}
	rc, errC := filepath.EvalSymlinks(child)
	rr, errR := filepath.EvalSymlinks(root)
	if errC == nil && errR == nil {
		return underPath(rc, rr)
	}
	return false
}

// underPath is the literal prefix test on cleaned paths.
func underPath(child, root string) bool {
	c := filepath.Clean(child)
	r := filepath.Clean(root)
	if c == r {
		return true
	}
	return strings.HasPrefix(c, r+string(filepath.Separator))
}

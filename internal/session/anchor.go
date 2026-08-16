// anchor.go — live-session anchor detection for worktree disposal.
//
// Backs the `moai worktree done` anchor guard: before removing a worktree,
// the CLI asks whether a live session is still anchored inside it. Removing
// the tree under a live session kills that session's shell — Claude Code's
// native worktree-isolation guard then refuses every Bash call whose working
// directory can no longer resolve inside the (deleted) tree.
package session

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// LiveAnchoredSessions returns registry entries that describe LIVE sessions
// anchored inside treePath: a working directory at or below treePath, on
// this host, whose owning process is still running.
//
// TWO registries are consulted, because the two lane launch shapes register
// in different places (measured 2026-08-17):
//
//  1. the tree-LOCAL registry (<treePath>/.moai/state/active-sessions.json)
//     — sessions whose own project root IS the tree;
//  2. the CALLER's project registry (CLAUDE_PROJECT_DIR, else CWD) — where
//     `moai cc -w` lanes actually register: the launcher pins the session's
//     project to the checkout it launched from, so the entry carrying
//     cwd=<tree> lives in that checkout's registry while the worktree
//     itself carries no local registry at all.
//
// Liveness is judged from the PID probe, with a heartbeat-freshness floor
// (DefaultStaleMinutes) as the conservative fallback for platforms where
// the probe cannot prove death. The heartbeat alone is NOT sufficient: no
// per-turn heartbeat driver exists today, so a long-running live lane
// carries a heartbeat hours old — the PID is what distinguishes it from a
// zombie.
//
// Fail-open: absent or unreadable registries contribute nothing (disposal
// must not wedge on a corrupt registry file).
//
// Sessions that entered the tree mid-session (Claude Code's EnterWorktree)
// are covered once the CwdChanged hook relocates their entry CWD
// (RelocateSession); before that relocation runs they keep their
// launch-time CWD and stay invisible here.
func LiveAnchoredSessions(treePath string, now time.Time) []Entry {
	if treePath == "" {
		return nil
	}
	roots := []string{treePath, callerProjectRoot()}
	host, _ := os.Hostname()
	stale := DefaultStaleMinutes * time.Minute
	seen := make(map[string]bool)
	var anchored []Entry
	for i, root := range roots {
		if root == "" {
			continue
		}
		if i > 0 && root == roots[0] {
			continue // same tree as the tree-local source
		}
		reg := NewRegistry(filepath.Join(root, DefaultRegistryPath), nil)
		entries, err := reg.Query("")
		if err != nil {
			continue
		}
		for _, e := range entries {
			if seen[e.SessionID] {
				continue
			}
			if e.Host != "" && e.Host != host {
				continue
			}
			if !pathWithinTree(e.CWD, treePath) {
				continue
			}
			alive := e.PID > 0 && isProcessAlive(e.PID)
			if alive || now.Sub(e.LastHeartbeat) < stale {
				seen[e.SessionID] = true
				anchored = append(anchored, e)
			}
		}
	}
	return anchored
}

// callerProjectRoot resolves the project root of the process asking about
// anchors: CLAUDE_PROJECT_DIR when the runtime set it, else the working
// directory. This is where a `moai cc -w` lane's entry lives when the
// disposal command runs from the launching checkout.
func callerProjectRoot() string {
	if v := os.Getenv(config.EnvClaudeProjectDir); v != "" {
		return v
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// RelocateSession updates the CWD field of the matching entry. It backs the
// CwdChanged hook (t74): a session that enters a worktree mid-session keeps
// its launch-time CWD otherwise, which makes it invisible to anchor
// detection. Idempotent on a missing entry — mirrors Heartbeat
// (REQ-COORD-004): no error when the session is not registered anywhere.
func (r *Registry) RelocateSession(sessionID, newCwd string) error {
	if sessionID == "" {
		return errors.New("session registry: sessionID cannot be empty")
	}
	if newCwd == "" {
		return errors.New("session registry: newCwd cannot be empty")
	}
	return r.withLock(func(entries []Entry) ([]Entry, error) {
		for i := range entries {
			if entries[i].SessionID == sessionID {
				entries[i].CWD = newCwd
				return entries, nil
			}
		}
		return entries, nil
	})
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

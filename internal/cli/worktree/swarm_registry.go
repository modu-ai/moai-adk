// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M4 swarm registry writer.
//
// Persists the per-SPEC active session state to
// .moai/state/swarm/<SPEC-ID>.json after a successful --team launch. The
// orchestrator and downstream tools read this file to discover live worktree
// sessions, their dispatch pattern (P1/P2/P3), and (for P1/P2) the tmux pane
// they were spawned into.
//
// Schema versioning: the 7 fields below are the canonical schema for
// REQ-WTL-008. Any new field must be additive (omitempty) so old consumers
// keep parsing successfully.
//
// File mode: 0o600 (per-user). The registry path lives under the user's home
// (~/.moai/worktrees/<project>/<spec>/.moai/state/swarm/) or, for in-tree
// worktrees, under the repo root, so 0o600 is sufficient — no shared-system
// access pattern exists.

package worktree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SwarmEntry captures the per-SPEC active session state written to
// .moai/state/swarm/<SPEC-ID>.json after a successful team launch.
//
// PaneID is empty for PatternP3InProgress (the in-process syscall.Exec path
// has no tmux pane). For PatternP1TmuxGLM / PatternP2TmuxCC the field carries
// the pane_id returned by `tmux new-window -P -F '#{pane_id}'` (typically
// "%5", "%7", etc.).
//
// CreatedAt is recorded in UTC (RFC3339 via encoding/json default). The
// orchestrator uses this timestamp to detect stale entries.
type SwarmEntry struct {
	SpecID       string    `json:"spec_id"`
	WorktreePath string    `json:"worktree_path"`
	Branch       string    `json:"branch"`
	PaneID       string    `json:"pane_id"`        // empty for P3; populated for P1/P2
	Mode         string    `json:"mode"`           // "tmux-glm" | "tmux-cc" | "in-progress-glm" | "in-progress-cc"
	CreatedAt    time.Time `json:"created_at"`     // RFC3339 UTC
	CreatedByPID int       `json:"created_by_pid"` // os.Getpid() at write time
}

// swarmRegistryDir returns the absolute path to the per-project
// .moai/state/swarm/ directory under repoRoot.
func swarmRegistryDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".moai", "state", "swarm")
}

// WriteSwarmEntry persists the SwarmEntry to
// <repoRoot>/.moai/state/swarm/<SpecID>.json. The parent directory is
// auto-created with 0o755 (per-user state dir). The file itself is written
// with 0o600 (per-user only).
//
// REQ-WTL-008 (registry schema): 7-field canonical layout.
//
// Failure modes:
//
//   - os.MkdirAll error: filesystem permission or disk-full. The caller
//     should treat this as non-fatal for the team launch itself (the tmux
//     window or syscall.Exec has already succeeded or is about to succeed)
//     and surface a stderr warning so the user is aware the registry is
//     incomplete.
//   - json.MarshalIndent error: only possible if SwarmEntry gains a field
//     with an un-marshalable type — a structural bug, not a runtime error.
//   - os.WriteFile error: filesystem permission, disk-full, or read-only
//     mount. Same non-fatal handling as MkdirAll.
//
// @MX:ANCHOR: [AUTO] WriteSwarmEntry — single canonical writer for the
// .moai/state/swarm/ registry. Fan-in: dispatchTeamLaunch (P1/P2 success
// branch + P3 pre-launch branch).
// @MX:REASON: Centralizing the schema + filename convention here prevents
// drift across the P1/P2/P3 dispatch branches. Any future writer
// (orchestrator, hook, etc.) MUST call this function rather than re-marshal
// independently.
func WriteSwarmEntry(repoRoot string, entry SwarmEntry) error {
	dir := swarmRegistryDir(repoRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir swarm registry dir: %w", err)
	}
	path := filepath.Join(dir, entry.SpecID+".json")
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal swarm entry: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write swarm entry: %w", err)
	}
	return nil
}

// patternToMode maps a Pattern + LLM string into the canonical
// SwarmEntry.Mode value. P4Handoff returns the empty string because the P4
// path never writes a registry entry (no spawn occurs); callers MUST
// short-circuit on PatternP4Handoff before invoking this function.
//
// The mapping table:
//
//	PatternP1TmuxGLM   → "tmux-glm"
//	PatternP2TmuxCC    → "tmux-cc"
//	PatternP3InProgress + llm="glm" → "in-progress-glm"
//	PatternP3InProgress + llm="cc"  → "in-progress-cc"
//	PatternP4Handoff   → "" (no registry written; defensive return)
func patternToMode(p Pattern, llm string) string {
	switch p {
	case PatternP1TmuxGLM:
		return "tmux-glm"
	case PatternP2TmuxCC:
		return "tmux-cc"
	case PatternP3InProgress:
		if llm == "glm" {
			return "in-progress-glm"
		}
		return "in-progress-cc"
	default:
		return ""
	}
}

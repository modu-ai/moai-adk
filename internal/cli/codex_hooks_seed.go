package cli

// codex_hooks_seed.go — SPEC-HANDOFF-NEUTRAL-001 M1.3 (REQ-HN-005/006/007).
//
// The fail-open wrapper around codexwiring.SeedHooksIfMissing shared by the
// worktree materializer (new trees) and the launcher entry path (backfill for
// trees created before this wiring existed). Seeding is additive, never a
// gate: a failure emits a diagnostic on the caller's writer and the caller
// proceeds (REQ-HN-007). Success stays quiet — the seed is not a user-facing
// event.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// sessionWorktreeSeedCodexHooks is the seam over the real seeding call so
// tests can run the real body against temp trees or observe the fail-open
// posture, without touching a real worktree.
var sessionWorktreeSeedCodexHooks = seedCodexHooksReal

// seedCodexHooksReal seeds tree's .codex/hooks.json (MoAI-owned entries only,
// existing file untouched) and is fail-open on every error path.
func seedCodexHooksReal(tree string, out io.Writer) {
	if tree == "" {
		return
	}
	if _, err := codexwiring.SeedHooksIfMissing(tree); err != nil {
		_, _ = fmt.Fprintf(out, "moai: codex hooks seeding skipped (%v); continuing without it\n", err)
	}
}

// seedWorktreeEntryHooks applies the launcher-entry backfill (design D2.5
// A-보완): only a tree that already exists on disk is seeded — one stat per
// entry. A not-yet-created tree is Claude Code's to create, and its next
// entry seeds it.
//
// Call this ONLY after the concurrent-writer admission succeeds: the backfill
// writes into the target tree, so a refused launch must never reach it (audit
// F0, sync-audit-opus.md — a refused entry left `?? .codex/hooks.json` behind).
func seedWorktreeEntryHooks(tree string, warn io.Writer) {
	if tree == "" {
		return
	}
	if info, err := os.Stat(tree); err != nil || !info.IsDir() {
		return
	}
	sessionWorktreeSeedCodexHooks(tree, warn)
}

// seedAdmittedWorktreeHooks resolves the -w/--worktree value of an already
// validated launcher argv the same way ccWorktreeWriterPrecheck and
// resolveWorktreeL2Path do, and applies the launcher-entry backfill. The entry
// flows call it at their post-admission point so the backfill only runs on an
// admitted launch. Resolution mirrors resolveWorktreeL2Path's: an absolute
// value is used directly (prefix validation already happened), a short name
// resolves against <root>/.claude/worktrees/<name>.
func seedAdmittedWorktreeHooks(args []string, warn io.Writer) {
	value, ok := worktreeFlagValue(args)
	if !ok || value == "" {
		return
	}
	tree := value
	if !filepath.IsAbs(tree) {
		root, err := findProjectRootFn()
		if err != nil || root == "" {
			return
		}
		tree = filepath.Join(root, ".claude", "worktrees", value)
	}
	seedWorktreeEntryHooks(tree, warn)
}

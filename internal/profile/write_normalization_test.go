package profile

// Write-side subtree normalization (card t297; SPEC-STATUSLINE-PROFILE-RESPECT-001
// REQ-009 / AC-009 — the deferred half of the t293 read-path work).
//
// The launch ledger's projects map is keyed by project root. Before t297 the
// recorder wrote whatever path it was handed, so a `moai cc -p X` from inside a
// worktree added one row per worktree — a row that dies with the worktree and,
// having no collector, accumulated monotonically (the growth lane-2 recorded in
// t293's verdict.md Residual-risk).
//
// These tests pin both halves of the fix:
//
//   - a subtree launch (no entry of its own) UPDATES the registered ancestor's
//     row instead of adding one (REQ-009), and resolution from a fresh sibling
//     worktree sees the update — read/write coherence with the t293 ancestor
//     walk, which stays the resolver for subtree directories;
//   - the behaviors that must not change stay pinned: a cold-start subtree
//     (no registered ancestor) still records its own live row, a subtree with
//     an entry of its own is updated in place (nested independent projects
//     never fold into a parent), and a legacy duplicate is updated rather than
//     deleted — deletion belongs to the reclamation path, not the recorder.

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// wnProjectsMap reads the ledger's projects map. A missing or malformed
// projects key yields a nil map — callers count len() and index it.
func wnProjectsMap(t *testing.T, base string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, "launch.yaml"))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	var ledger map[string]any
	if err := yaml.Unmarshal(data, &ledger); err != nil {
		t.Fatalf("unmarshal ledger: %v", err)
	}
	projects, _ := ledger["projects"].(map[string]any)
	return projects
}

// wnWorktree creates a real worktree-shaped directory under root so the write
// side's normalization (and the read side's walk) can Stat every segment.
func wnWorktree(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, ".claude", "worktrees", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create worktree %q: %v", dir, err)
	}
	return dir
}

// TestRecordFromSubtreeFoldsIntoRegisteredRoot (AC-009, REQ-009): a launch
// recorded from a subtree of a REGISTERED project updates the registered
// root's row — one registration keys all its subtrees, so creating and
// launching from worktrees never grows the ledger.
func TestRecordFromSubtreeFoldsIntoRegisteredRoot(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta")

	root := filepath.Join(t.TempDir(), "proj")
	wt1 := wnWorktree(t, root, "wt1")
	wt2 := wnWorktree(t, root, "wt2")
	wt3 := wnWorktree(t, root, "wt3")
	subtreeLedger(t, base, [2]string{root, "alpha"})

	// Each worktree launch records the subtree path. The fold must send every
	// one of them to the root's row: no new row per worktree, monotonic growth
	// gone.
	for _, wt := range []string{wt1, wt2, wt3} {
		if err := RecordLastUsedProfileForProject(wt, "beta"); err != nil {
			t.Fatalf("RecordLastUsedProfileForProject(%q): %v", wt, err)
		}
	}

	projects := wnProjectsMap(t, base)
	if len(projects) != 1 {
		t.Fatalf("projects rows = %d (%v), want 1 — worktree launches must fold into the registered root, not add rows",
			len(projects), projects)
	}
	if got := projects[normalizeProjectKey(root)]; got != "beta" {
		t.Errorf("projects[root] = %v, want beta (the last explicit choice)", got)
	}
}

// TestRecordFromSubtreeThenResolveFromFreshSiblingWorktree pins read/write
// coherence: after a subtree launch is folded into the root, a FRESH sibling
// worktree (no entry of its own) resolves through the t293 ancestor walk to
// the recorded profile. The write-side normalization must leave the walk as
// the resolver for subtrees — not strand fresh worktrees on the old value.
func TestRecordFromSubtreeThenResolveFromFreshSiblingWorktree(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta")

	root := filepath.Join(t.TempDir(), "proj")
	wt1 := wnWorktree(t, root, "wt1")
	wt2 := wnWorktree(t, root, "wt2")
	subtreeLedger(t, base, [2]string{root, "alpha"})

	if err := RecordLastUsedProfileForProject(wt1, "beta"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject: %v", err)
	}

	if got := ResolveLaunchProfileForProject(wt2, ""); got != "beta" {
		t.Errorf("fresh sibling worktree resolved to %q, want beta — the folded write is invisible to the ancestor walk", got)
	}
}

// TestRecordFromSubtreeWithNoRegisteredAncestorKeepsOwnRow pins the
// cold-start behavior: with no registered ancestor the recorder has nowhere to
// fold, so the row lives at the subtree path. The row is live while the
// worktree exists and is reclaimed with it (PruneStaleProjectEntries after
// disposal) — it is the bounded residue, not the growth mechanism.
func TestRecordFromSubtreeWithNoRegisteredAncestorKeepsOwnRow(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	root := filepath.Join(t.TempDir(), "fresh-proj")
	wt1 := wnWorktree(t, root, "wt1")

	if err := RecordLastUsedProfileForProject(wt1, "alpha"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject: %v", err)
	}

	projects := wnProjectsMap(t, base)
	key := normalizeProjectKey(wt1)
	if got := projects[key]; got != "alpha" {
		t.Errorf("projects[%q] = %v, want alpha (cold-start keeps its own row)", key, got)
	}
	if len(projects) != 1 {
		t.Errorf("projects rows = %d, want 1", len(projects))
	}
	if got := ResolveLaunchProfileForProject(wt1, ""); got != "alpha" {
		t.Errorf("cold-start worktree resolved to %q, want alpha", got)
	}
}

// TestRecordFromRegisteredNestedProjectUpdatesOwnRowNotParent pins the
// nested-project safety: two REGISTERED projects nested inside one another
// (/mono and /mono/lib) are independent projects. A launch from the inner one
// updates the inner row; the outer registration is untouched. This is why the
// recorder resolves the write key by exact/alias hit FIRST and only falls to
// the ancestor when the path has no entry of its own — the same precedence the
// read side applies (exact match beats subtree).
func TestRecordFromRegisteredNestedProjectUpdatesOwnRowNotParent(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta", "gamma")

	mono := filepath.Join(t.TempDir(), "mono")
	lib := filepath.Join(mono, "lib")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatalf("create %q: %v", lib, err)
	}
	subtreeLedger(t, base, [2]string{mono, "alpha"}, [2]string{lib, "beta"})

	if err := RecordLastUsedProfileForProject(lib, "gamma"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject: %v", err)
	}

	projects := wnProjectsMap(t, base)
	if len(projects) != 2 {
		t.Fatalf("projects rows = %d (%v), want 2 — a registered nested project is its own row, never folded into its parent",
			len(projects), projects)
	}
	if got := projects[normalizeProjectKey(lib)]; got != "gamma" {
		t.Errorf("projects[lib] = %v, want gamma (own row updated in place)", got)
	}
	if got := projects[normalizeProjectKey(mono)]; got != "alpha" {
		t.Errorf("projects[mono] = %v, want alpha (parent registration untouched)", got)
	}
}

// TestRecordFromSubtreeWithLegacyOwnRowUpdatesOwnRow documents the residual a
// pre-t297 binary leaves: a subtree that already carries its own row (written
// by the old recorder) keeps it — the write updates it in place, adds nothing,
// and deletes nothing. The duplicate ends when the worktree is disposed and
// PruneStaleProjectEntries reclaims the dead key; the recorder itself is not
// in the deletion business.
func TestRecordFromSubtreeWithLegacyOwnRowUpdatesOwnRow(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta", "gamma")

	root := filepath.Join(t.TempDir(), "proj")
	wt1 := wnWorktree(t, root, "wt1")
	subtreeLedger(t, base, [2]string{root, "alpha"}, [2]string{wt1, "beta"})

	if err := RecordLastUsedProfileForProject(wt1, "gamma"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject: %v", err)
	}

	projects := wnProjectsMap(t, base)
	if len(projects) != 2 {
		t.Fatalf("projects rows = %d (%v), want 2 — a legacy duplicate is updated, not folded away and not multiplied",
			len(projects), projects)
	}
	if got := projects[normalizeProjectKey(wt1)]; got != "gamma" {
		t.Errorf("projects[wt1] = %v, want gamma (legacy own row updated in place)", got)
	}
	if got := projects[normalizeProjectKey(root)]; got != "alpha" {
		t.Errorf("projects[root] = %v, want alpha (untouched)", got)
	}
	// Exact match beats subtree on the read side, unchanged.
	if got := ResolveLaunchProfileForProject(wt1, ""); got != "gamma" {
		t.Errorf("wt1 resolved to %q, want gamma (own row shadows the ancestor)", got)
	}
}

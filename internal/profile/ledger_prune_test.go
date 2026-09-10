package profile

// Dead-row reclamation for the launch ledger (card t297, REQ-009's lifecycle
// half). PruneStaleProjectEntries removes projects[] entries whose directory
// is gone; these tests pin the predicate (dead removed, live kept),
// idempotence, the non-projects keys' survival, the no-ledger and corrupt
// ledgers, and the two bounded-disposal regressions the card demands:
// N worktrees created and disposed leave the ledger bounded, and the
// cold-start residue is fully reclaimed.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// wnLedgerRaw reads the ledger file's raw text (for corrupt-fixture writes).
func wnLedgerRaw(t *testing.T, base string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, "launch.yaml"))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	return string(data)
}

func TestPruneStaleProjectEntries_RemovesDeadKeepsLive(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	liveRoot := filepath.Join(t.TempDir(), "live-proj")
	deadRoot := filepath.Join(t.TempDir(), "dead-proj")
	if err := os.MkdirAll(liveRoot, 0o755); err != nil {
		t.Fatalf("create live root: %v", err)
	}
	// deadRoot is deliberately NOT created: its row is the stale class.
	subtreeLedger(t, base, [2]string{liveRoot, "alpha"}, [2]string{deadRoot, "alpha"})

	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("PruneStaleProjectEntries: %v", err)
	}

	wantKey := normalizeProjectKey(deadRoot)
	if len(pruned) != 1 || pruned[0] != wantKey {
		t.Fatalf("pruned = %v, want [%s]", pruned, wantKey)
	}
	projects := wnProjectsMap(t, base)
	if _, still := projects[wantKey]; still {
		t.Errorf("dead entry %q survived the prune", wantKey)
	}
	if got := projects[normalizeProjectKey(liveRoot)]; got != "alpha" {
		t.Errorf("live entry = %v, want alpha (live rows are never touched)", got)
	}
	if !strings.Contains(wnLedgerRaw(t, base), "last_profile") && !strings.Contains(wnLedgerRaw(t, base), "projects") {
		t.Errorf("ledger lost its structure after prune: %q", wnLedgerRaw(t, base))
	}
}

func TestPruneStaleProjectEntries_Idempotent(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	deadRoot := filepath.Join(t.TempDir(), "dead-proj") // never created
	liveRoot := filepath.Join(t.TempDir(), "live-proj")
	if err := os.MkdirAll(liveRoot, 0o755); err != nil {
		t.Fatalf("create live root: %v", err)
	}
	subtreeLedger(t, base, [2]string{liveRoot, "alpha"}, [2]string{deadRoot, "alpha"})

	if _, err := PruneStaleProjectEntries(); err != nil {
		t.Fatalf("first prune: %v", err)
	}
	first := wnLedgerRaw(t, base)

	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("second prune: %v", err)
	}
	if len(pruned) != 0 {
		t.Errorf("second prune removed %v, want nothing (idempotence)", pruned)
	}
	if second := wnLedgerRaw(t, base); second != first {
		t.Errorf("second prune rewrote the ledger:\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestPruneStaleProjectEntries_NoLedgerIsNoOp(t *testing.T) {
	subtreeSandbox(t) // sandboxed base, no launch.yaml written

	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("PruneStaleProjectEntries on a missing ledger: %v", err)
	}
	if len(pruned) != 0 {
		t.Errorf("pruned = %v, want empty", pruned)
	}
}

func TestPruneStaleProjectEntries_LedgerWithoutProjectsIsNoOp(t *testing.T) {
	base := subtreeSandbox(t, "alpha")
	pmWriteLedger(t, base, "model: claude-opus-4-6\nlast_profile: alpha\n")

	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("PruneStaleProjectEntries: %v", err)
	}
	if len(pruned) != 0 {
		t.Errorf("pruned = %v, want empty", pruned)
	}
	if got := wnLedgerRaw(t, base); !strings.Contains(got, "model: claude-opus-4-6") {
		t.Errorf("ledger without projects was rewritten: %q", got)
	}
}

func TestPruneStaleProjectEntries_CorruptLedgerErrors(t *testing.T) {
	base := subtreeSandbox(t, "alpha")
	pmWriteLedger(t, base, "\t::not yaml [{{{")

	if _, err := PruneStaleProjectEntries(); err == nil {
		t.Fatal("PruneStaleProjectEntries on a corrupt ledger returned nil, want an error")
	}
}

// TestPruneStaleProjectEntries_KeepsEntryWithDeadProfile pins the predicate's
// boundary: only the KEY's directory is judged. An entry whose recorded
// PROFILE directory vanished is stale for resolution, but the binding may be
// wanted again (the profile can be recreated) and the read side already falls
// through it — the prune must not erase wanted bindings.
func TestPruneStaleProjectEntries_KeepsEntryWithDeadProfile(t *testing.T) {
	base := subtreeSandbox(t) // NO profile directories created at all

	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create root: %v", err)
	}
	subtreeLedger(t, base, [2]string{root, "ghost"})

	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("PruneStaleProjectEntries: %v", err)
	}
	if len(pruned) != 0 {
		t.Fatalf("pruned = %v, want nothing — a dead PROFILE is not a dead PROJECT", pruned)
	}
	if got := wnProjectsMap(t, base)[normalizeProjectKey(root)]; got != "ghost" {
		t.Errorf("entry = %v, want ghost (kept)", got)
	}
}

// --- Card regressions (execution-based) ---

// TestRegression_DisposedWorktreesLeaveLedgerBounded is the card's regression
// (a): N worktrees created, launched from, and disposed leave the ledger at a
// bound that does not depend on N. With the project root registered, the
// launches fold into the root row (N launches, zero new rows); after disposal
// + prune the ledger still holds exactly the root row.
func TestRegression_DisposedWorktreesLeaveLedgerBounded(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create root: %v", err)
	}
	subtreeLedger(t, base, [2]string{root, "alpha"})

	const n = 6
	wtPaths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		wt := wnWorktree(t, root, fmt.Sprintf("card-t%d", i))
		wtPaths = append(wtPaths, wt)
		if err := RecordLastUsedProfileForProject(wt, "alpha"); err != nil {
			t.Fatalf("launch %d: %v", i, err)
		}
		if got := len(wnProjectsMap(t, base)); got != 1 {
			t.Fatalf("after launch %d: rows = %d, want 1 (writes fold into the root)", i, got)
		}
	}

	// Dispose every worktree, then run the reclamation the disposal path
	// invokes.
	for _, wt := range wtPaths {
		if err := os.RemoveAll(wt); err != nil {
			t.Fatalf("dispose %q: %v", wt, err)
		}
	}
	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("prune after disposal: %v", err)
	}
	if len(pruned) != 0 {
		t.Fatalf("prune removed %v, want nothing — folded writes create no dead rows", pruned)
	}
	projects := wnProjectsMap(t, base)
	if len(projects) != 1 {
		t.Fatalf("final rows = %d (%v), want 1 — ledger must stay bounded as worktrees come and go",
			len(projects), projects)
	}
	if got := ResolveLaunchProfileForProject(root, ""); got != "alpha" {
		t.Errorf("root resolved to %q, want alpha", got)
	}
}

// TestRegression_ColdStartResidueIsReclaimed is regression (a)'s cold-start
// half: with NO registered ancestor, each launched worktree holds a live row
// (bounded by live trees, not history); disposal + prune removes every one of
// them, returning the ledger to empty.
func TestRegression_ColdStartResidueIsReclaimed(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	root := t.TempDir()
	const n = 4
	wtPaths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		wt := wnWorktree(t, root, fmt.Sprintf("wt-%d", i))
		wtPaths = append(wtPaths, wt)
		if err := RecordLastUsedProfileForProject(wt, "alpha"); err != nil {
			t.Fatalf("launch %d: %v", i, err)
		}
	}
	if got := len(wnProjectsMap(t, base)); got != n {
		t.Fatalf("live cold-start rows = %d, want %d (bounded by live trees)", got, n)
	}

	for _, wt := range wtPaths {
		if err := os.RemoveAll(wt); err != nil {
			t.Fatalf("dispose %q: %v", wt, err)
		}
	}
	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("prune after disposal: %v", err)
	}
	if len(pruned) != n {
		t.Fatalf("prune removed %d rows, want %d", len(pruned), n)
	}
	if got := len(wnProjectsMap(t, base)); got != 0 {
		t.Fatalf("final rows = %d, want 0 — cold-start residue must be fully reclaimed", got)
	}
}

// TestRegression_LegacyDuplicatesFoldDeadOnesFirst is regression (b): a
// ledger seeded the way the pre-t297 binary left it (root + worktree rows)
// converges — the live duplicate stays (no deletion on write), and once the
// worktree is disposed the prune folds the residue away, leaving exactly the
// distinct projects.
func TestRegression_LegacyDuplicatesFoldDeadOnesFirst(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta")

	root := t.TempDir()
	wt := wnWorktree(t, root, "t289")
	// Seed exactly what the machine's real ledger carried: a root row and a
	// per-worktree row from the old recorder.
	subtreeLedger(t, base, [2]string{root, "alpha"}, [2]string{wt, "beta"})

	// A launch from the worktree under the fixed recorder updates the
	// worktree's own row (exact hit precedes the ancestor) and adds nothing.
	if err := RecordLastUsedProfileForProject(wt, "beta"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if got := len(wnProjectsMap(t, base)); got != 2 {
		t.Fatalf("rows after legacy-duplicate launch = %d, want 2 (no growth)", got)
	}

	// Dispose the worktree: the duplicate's residue is now dead and folds away.
	if err := os.RemoveAll(wt); err != nil {
		t.Fatalf("dispose: %v", err)
	}
	pruned, err := PruneStaleProjectEntries()
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if len(pruned) != 1 {
		t.Fatalf("pruned = %v, want exactly the dead worktree row", pruned)
	}
	projects := wnProjectsMap(t, base)
	if len(projects) != 1 {
		t.Fatalf("final rows = %d (%v), want 1 == distinct projects", len(projects), projects)
	}
	if got := projects[normalizeProjectKey(root)]; got != "alpha" {
		t.Errorf("root row = %v, want alpha", got)
	}
}

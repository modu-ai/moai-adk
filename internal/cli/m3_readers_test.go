package cli

// m3_readers_test.go — SPEC-WORKTREE-HOME-ROOT-001 M3 RED (card t1640):
// the cli readers of AC-WHR-007 and the transition resolver of AC-WHR-022.
//
// Written before the readers learn the home root: today every reader here
// derives its candidates from the two in-repo roots only, so a tree at
// <MoaiHome>/worktrees/projects/<key>/<name> is invisible. The failures are
// the RED (behavioral; the tests name M1 and existing symbols per research
// §R13's M3 row). Legacy shapes are pinned as characterization.
//
// AC-WHR-022 table: t8 (home only) and t5 (home, one directory) resolve to
// their single tree; t7 (home + .moai) and t6 (.moai + .claude) are reported
// ambiguous with a message naming both candidates and instructing an
// absolute path; no root is preferred and neither lookup guesses.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// m3World is the AC-WHR-007 fixture repository: a scratch primary with
// trees at all three roots and the recorded shapes the readers must find.
type m3World struct {
	repo, home             string
	homeT7, homeT5, homeT8 string // home-root trees
	legacyMoaiT7b, moaiT6  string // legacy-moai trees
	claudeT7c, claudeT6    string // legacy-claude trees
}

func m3NewWorld(t *testing.T) *m3World {
	t.Helper()
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	key := filepath.Base(repo)
	w := &m3World{
		repo:          repo,
		home:          home,
		homeT7:        filepath.Join(home, "worktrees", "projects", key, "t7"),
		homeT5:        filepath.Join(home, "worktrees", "projects", key, "t5"),
		homeT8:        filepath.Join(home, "worktrees", "projects", key, "t8"),
		legacyMoaiT7b: filepath.Join(repo, ".moai", "worktrees", "t7b"),
		moaiT6:        filepath.Join(repo, ".moai", "worktrees", "t6"),
		claudeT7c:     filepath.Join(repo, ".claude", "worktrees", "t7c"),
		claudeT6:      filepath.Join(repo, ".claude", "worktrees", "t6"),
	}
	for _, d := range []string{w.homeT7, w.homeT5, w.homeT8, w.legacyMoaiT7b, w.moaiT6, w.claudeT7c, w.claudeT6} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = origFn })
	return w
}

// ------------------------------------------------- AC-WHR-007 factory trio

// TestM3FactoryCardWorktreeDir_HomeRoot pins the landing directory: the
// resolver's home-root path, not the in-repo .moai/worktrees leaf.
func TestM3FactoryCardWorktreeDir_HomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	got := factoryCardWorktreeDir(w.repo, "t7")
	if got != w.homeT7 {
		t.Errorf("factoryCardWorktreeDir = %q, want the home-root landing path %q", got, w.homeT7)
	}
}

// TestM3FactoryRefuseForeignWorktree_HomeRoot pins the REQ-SD-011 refusal at
// the new depth: a foreign directory at the card's home-root landing path is
// refused before any lease edge.
func TestM3FactoryRefuseForeignWorktree_HomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	if err := os.MkdirAll(w.homeT7, 0o755); err != nil {
		t.Fatalf("mkdir foreign dir: %v", err)
	}
	err := factoryRefuseForeignWorktree(withFactoryLandingBase(context.Background(), w.repo), w.repo, "t7")
	if err == nil {
		t.Fatal("a foreign directory at the home-root landing path was not refused")
	}
	if !strings.Contains(err.Error(), w.homeT7) {
		t.Errorf("refusal %q does not name the foreign directory", err)
	}
}

// TestM3FactoryCarryPreviousWorktree_HomeRoot pins the t1521 carry-over at
// the new depth: a previous run of the same card recorded the home-root
// tree, the directory still exists, so the replacement run carries it
// instead of refusing it as foreign.
func TestM3FactoryCarryPreviousWorktree_HomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	if err := os.MkdirAll(w.homeT7, 0o755); err != nil {
		t.Fatalf("mkdir carried tree: %v", err)
	}
	db, err := homestate.OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatalf("open factory db: %v", err)
	}
	defer func() { _ = db.Close() }()
	wt := w.homeT7
	if _, err := db.RecordPicked(context.Background(), "run-prev", "t7", homestate.CardFields{WorktreePath: &wt}, "m3-carry", time.Now().UTC()); err != nil {
		t.Fatalf("record picked: %v", err)
	}
	got, err := factoryCarryPreviousWorktree(withFactoryLandingBase(context.Background(), w.repo), db, w.repo, "run-new", "t7")
	if err != nil {
		t.Fatalf("factoryCarryPreviousWorktree: %v", err)
	}
	if got != w.homeT7 {
		t.Errorf("carry = %q, want the recorded home-root tree %q", got, w.homeT7)
	}
}

// ------------------------------------------------------- AC-WHR-007 readers

// TestM3ResolveLaunchWorktreePath_FindsActualTree pins the chain-ledger
// resolver on the actual locations: the home-root tree resolves to its
// home path, the legacy-moai tree to its own path, and a name with no tree
// keeps the legacy .claude spelling (branch Y's pass-through).
func TestM3ResolveLaunchWorktreePath_FindsActualTree(t *testing.T) {
	w := m3NewWorld(t)
	cases := []struct{ name, want string }{
		{"t7", w.homeT7},
		{"t7b", w.legacyMoaiT7b},
		{"nosuch", filepath.Join(w.repo, ".claude", "worktrees", "nosuch")},
	}
	for _, c := range cases {
		got, ok := resolveLaunchWorktreePath(c.name)
		if !ok {
			t.Errorf("resolveLaunchWorktreePath(%q) resolved nothing, want %q", c.name, c.want)
			continue
		}
		if got != c.want {
			t.Errorf("resolveLaunchWorktreePath(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestM3ResolveWorktreeLaunchTarget_FindsHomeTree pins the anchor resolver:
// a short name naming a home-root tree maps to that tree's path.
func TestM3ResolveWorktreeLaunchTarget_FindsHomeTree(t *testing.T) {
	w := m3NewWorld(t)
	if got := resolveWorktreeLaunchTarget(w.repo, "t7"); got != w.homeT7 {
		t.Errorf("resolveWorktreeLaunchTarget(t7) = %q, want the home-root tree %q", got, w.homeT7)
	}
	if got := resolveWorktreeLaunchTarget(w.repo, "nosuch"); got != filepath.Join(shortWorktreeRoot(w.repo), ".claude", "worktrees", "nosuch") {
		t.Errorf("resolveWorktreeLaunchTarget(nosuch) = %q, want the legacy default spelling", got)
	}
}

// TestM3SeedAdmittedWorktreeHooksFor_HomeRoot pins the entry backfill: the
// short name resolves to the home-root tree and the seed lands there.
func TestM3SeedAdmittedWorktreeHooksFor_HomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	seedAdmittedWorktreeHooksFor(worktreeRead{Value: "t7", Has: true}, nil)
	if _, err := os.Stat(filepath.Join(w.homeT7, ".codex", "hooks.json")); err != nil {
		t.Errorf("the home-root tree was not seeded (hooks.json: %v); the reader resolved the wrong location", err)
	}
}

// TestM3CCWorktreeWriterPrecheck_HomeRoot pins the concurrent-writer
// precheck on the new depth: the short name resolves to the home-root tree
// and the precheck reads THAT tree's lock state — observable because a
// real registered worktree with no anchored session answers nil, while a
// mis-resolved location would fail its git lock read.
func TestM3CCWorktreeWriterPrecheck_HomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	// Replace the plain fixture directory with a real registered worktree:
	// the precheck's lock-state read runs git inside the resolved tree.
	if err := os.RemoveAll(w.homeT7); err != nil {
		t.Fatalf("clear fixture dir: %v", err)
	}
	m3Gitf(t, w.repo, "worktree", "add", "-q", "-b", "WT-t7", "--", w.homeT7)
	// No anchored session: the precheck answers nil for an unanchored tree.
	if err := ccWorktreeWriterPrecheck(worktreeRead{Value: "t7", Has: true}); err != nil {
		t.Errorf("ccWorktreeWriterPrecheck on an unanchored home-root tree = %v, want nil", err)
	}
}

// TestM3AutoOwnerWorktrees_IncludesHomeRoot pins the owner probe's
// candidates: the home-root tree of the card is among them, so a card whose
// recorded tree is outside the legacy roots is not judged owner-not-alive
// (REQ-WHR-008).
func TestM3AutoOwnerWorktrees_IncludesHomeRoot(t *testing.T) {
	w := m3NewWorld(t)
	cands := autoOwnerWorktrees(w.repo, "t7")
	found := false
	for _, c := range cands {
		if c == w.homeT7 {
			found = true
		}
	}
	if !found {
		t.Errorf("autoOwnerWorktrees = %v; the home-root candidate %q is missing", cands, w.homeT7)
	}
}

// ---------------------------------------------------- AC-WHR-022 ambiguity

// TestM3ResolveCodexWorktreeDir_ThreeRoots pins the transition resolver on
// AC-WHR-022's own fixture shape: t7 at the home root AND at
// .moai/worktrees (two different trees, one name), t6 at both legacy roots
// (the GH #1813 shape), t5 at the home root under one directory, t8 at one
// root. Single-tree names resolve; two-tree names are ambiguous with both
// candidates named and an absolute-path instruction; no root is preferred;
// a name at no root is refused rather than guessed.
func TestM3ResolveCodexWorktreeDir_ThreeRoots(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	key := filepath.Base(repo)
	homeT7 := filepath.Join(home, "worktrees", "projects", key, "t7")
	homeT5 := filepath.Join(home, "worktrees", "projects", key, "t5")
	homeT8 := filepath.Join(home, "worktrees", "projects", key, "t8")
	moaiT7 := filepath.Join(repo, ".moai", "worktrees", "t7")
	moaiT6 := filepath.Join(repo, ".moai", "worktrees", "t6")
	claudeT6 := filepath.Join(repo, ".claude", "worktrees", "t6")
	for _, d := range []string{homeT7, homeT5, homeT8, moaiT7, moaiT6, claudeT6} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = origFn })

	// Single-tree names resolve to their one tree.
	for _, c := range []struct{ name, want string }{
		{"t5", homeT5},
		{"t8", homeT8},
	} {
		got, err := resolveCodexWorktreeDir(repo, c.name)
		if err != nil {
			t.Errorf("resolveCodexWorktreeDir(%q): %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveCodexWorktreeDir(%q) = %q, want %q", c.name, got, c.want)
		}
	}

	// Two-tree names are ambiguous, naming both candidates and instructing
	// an absolute path; no root is preferred.
	for _, c := range []struct{ name, first, second string }{
		{"t7", homeT7, moaiT7},
		{"t6", moaiT6, claudeT6},
	} {
		got, err := resolveCodexWorktreeDir(repo, c.name)
		if err == nil {
			t.Errorf("resolveCodexWorktreeDir(%q) = %q; a name at two roots must be reported ambiguous", c.name, got)
			continue
		}
		if !strings.Contains(err.Error(), "absolute path") {
			t.Errorf("ambiguity error for %q does not instruct an absolute path: %v", c.name, err)
		}
		if !strings.Contains(err.Error(), c.first) || !strings.Contains(err.Error(), c.second) {
			t.Errorf("ambiguity error for %q must name both candidates (%s, %s): %v", c.name, c.first, c.second, err)
		}
	}

	// A name at no root is refused, never guessed.
	if _, err := resolveCodexWorktreeDir(repo, "t9"); err == nil {
		t.Error("resolveCodexWorktreeDir(t9) succeeded; a name at no root must be refused")
	}
}

// TestM3CleanupMovedByWitnessSmoke pins that the cleanup guard's reverse
// lookup is reachable from this package with the M1 witness: a moved tree
// standing at its destination reads as moved (the full nine-tree table is
// m3_cleanup_guard_test.go).
func TestM3CleanupMovedByWitnessSmoke(t *testing.T) {
	w := m3NewWorld(t)
	if worktree.MovedBy(w.repo, w.homeT7) {
		t.Error("MovedBy answered true for a tree with no move list; the M1 witness regressed")
	}
}

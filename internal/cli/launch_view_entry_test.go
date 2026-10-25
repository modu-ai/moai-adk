package cli

// Regression tests for card t1791 (GH #1813) and its audit rounds: the -w
// NAME form must enter an existing same-name .moai/worktrees/<name> MoAI
// tree instead of passing the name through to claude, whose short-name
// resolution covers only .claude/worktrees/<name> — the pass-through used to
// create a duplicate branch tree silently (the 2026-10-08 lane-12 incident).
// The entry judges the VIEW's backend-model selection (FinalArgv, option-
// model aware) and the backend argv loses every worktree selection on entry.
// Fixtures are scratch repositories under t.TempDir() with their own
// MOAI_HOME; the project-root seam is stubbed so nothing escapes the fixture.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// entryFixture builds a scratch primary repo with its own MOAI_HOME and
// points the project-root seam at it, returning the primary root. The repo
// must be a REAL git repository: worktreeroot.resolve derives the project
// key from the git common dir.
func entryFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	primary := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}
	if out, err := exec.Command("git", "init", "-q", primary).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return primary, nil }
	t.Cleanup(func() { findProjectRootFn = orig })
	return primary
}

// mustMkdirTree creates a directory subtree, failing the test on error.
func mustMkdirTree(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

// entryDirFor resolves the entry directory for argv through the full view —
// the same shape the cc/glm flows judge.
func entryDirFor(t *testing.T, argv []string) string {
	t.Helper()
	return existingWorktreeEntryDir(resolveLaunchView(argv))
}

// canonical resolves symlinks (the macOS /var alias) so paths derived from
// the resolver's canonicalization agree with the paths asserted here.
func canonical(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// TestEntryDir_LegacyMoaiTreeIsEntered — the #1813 fix: a short name whose
// only existing candidate is .moai/worktrees/<name> answers that directory
// (the launcher enters it by directory and strips the selection), never a
// pass-through that would let claude create .claude/worktrees/<name>.
func TestEntryDir_LegacyMoaiTreeIsEntered(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	moaiTree := filepath.Join(primary, ".moai", "worktrees", name)
	mustMkdirTree(t, moaiTree)
	if got, want := entryDirFor(t, []string{"-w", name}), canonical(t, moaiTree); got != want {
		t.Fatalf("entry dir = %q, want the existing .moai tree %q (a pass-through here creates a duplicate .claude branch tree — GH #1813)", got, want)
	}
}

// TestEntryDir_LegacyMoaiAndClaudeIsAmbiguous — both legacy roots carrying
// the name stay ambiguous: the entry answers "" and the ambiguity report
// belongs to the callers (REQ-WHR-022) — unchanged by the #1813 fix.
func TestEntryDir_LegacyMoaiAndClaudeIsAmbiguous(t *testing.T) {
	primary := entryFixture(t)
	name := "both"
	for _, rel := range []string{
		filepath.Join(".moai", "worktrees", name),
		filepath.Join(".claude", "worktrees", name),
	} {
		mustMkdirTree(t, filepath.Join(primary, rel))
	}
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the two-root ambiguity", got)
	}
}

// TestEntryDir_LegacyClaudeStaysNative — a name whose only candidate is
// .claude/worktrees/<name> keeps answering "": claude re-enters that root
// natively, and the entry must not shadow it.
func TestEntryDir_LegacyClaudeStaysNative(t *testing.T) {
	primary := entryFixture(t)
	name := "native"
	mustMkdirTree(t, filepath.Join(primary, ".claude", "worktrees", name))
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the legacy-claude native re-entry", got)
	}
}

// TestEntrySelection_StrippedOnEntry — the launcher-side half of the
// no-creation guarantee: when the entry resolves to a directory, the BACKEND
// argv handed to the launch carries no worktree selection at all, so claude
// can never resolve a short name against its own .claude/worktrees and
// create a duplicate tree.
func TestEntrySelection_StrippedOnEntry(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	mustMkdirTree(t, filepath.Join(primary, ".moai", "worktrees", name))
	view := resolveLaunchView([]string{"-w", name})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve for the strip to be reachable")
	}
	rest := stripBackendWorktreeSelections(view.FinalArgv)
	for _, tok := range rest {
		if strings.HasPrefix(tok, "-w") || strings.HasPrefix(tok, "--worktree") || tok == name {
			t.Fatalf("worktree selection survived the strip: %v", rest)
		}
	}
}

// TestEntryDir_DuplicateSelectionEntersByFinal — a duplicate selection pair
// enters by the FINAL selection (the backend's own last-wins convention the
// view's Selection encodes; card t1791 audit round 2 finding 1): the first
// strip used to leave the second selection alive in the backend argv.
func TestEntryDir_DuplicateSelectionEntersByFinal(t *testing.T) {
	primary := entryFixture(t)
	second := filepath.Join(primary, ".moai", "worktrees", "second")
	mustMkdirTree(t, second)
	got := entryDirFor(t, []string{"-w", "first", "-w", "second"})
	if got != canonical(t, second) {
		t.Fatalf("entry dir = %q, want the final selection's tree %q", got, canonical(t, second))
	}
}

// TestEntryRead_TargetsTheEnteredTree — the writer-precheck read of the
// entry flow names the ENTERED tree, never the view's first-match StageB
// read (card t1791 audit round 2 finding 1): under a duplicate pair the
// lock check must probe the tree that is about to be entered, or the
// concurrent-writer protection is bypassed.
func TestEntryRead_TargetsTheEnteredTree(t *testing.T) {
	primary := entryFixture(t)
	for _, name := range []string{"safe", "locked"} {
		mustMkdirTree(t, filepath.Join(primary, ".moai", "worktrees", name))
	}
	view := resolveLaunchView([]string{"-w", "safe", "-w", "locked"})
	entryDir := existingWorktreeEntryDir(view)
	if entryDir == "" {
		t.Fatal("precondition: the entry must resolve to the final selection's tree")
	}
	read := entryWorktreeRead(entryDir)
	if !read.Has || !strings.Contains(read.Value, "locked") {
		t.Fatalf("precheck read = %+v, want the entered tree (locked), not StageB's %q", read, view.StageB.Value)
	}
}

// TestEntryDir_OptionValueIsNotASelection — a --worktree token that is the
// VALUE of another backend option is never a selection (card t1791 audit
// round 2 finding 2): the backend option model consumes it, the entry does
// not fire on it, and the backend argv passes through untouched.
func TestEntryDir_OptionValueIsNotASelection(t *testing.T) {
	primary := entryFixture(t)
	second := filepath.Join(primary, ".moai", "worktrees", "second")
	mustMkdirTree(t, second)
	argv := []string{"-w", "first", "--append-system-prompt", "--worktree", "second"}
	view := resolveLaunchView(argv)
	if got := existingWorktreeEntryDir(view); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — --worktree here is the prompt option's value, not a selection", got)
	}
	// And the non-entry flow hands the backend the untouched LaunchArgs (the
	// strip runs only on the entry path — cc.go/glm.go wire it there), so
	// the prompt option and its value survive the launch intact.
	joined := strings.Join(view.LaunchArgs, " ")
	if !strings.Contains(joined, "--append-system-prompt --worktree second") {
		t.Fatalf("the backend argv lost the prompt option or its value: %v", view.LaunchArgs)
	}
}

// TestStrip_FinalArgvLosesEverySelection — the backend argv after the
// launcher consumption loses every worktree selection on the entry path,
// including one the launcher's own dropped separator had sheltered (card
// t1791 audit round 2 finding 3): a surviving selection re-opens the
// nested-creation path after the directory entry.
func TestStrip_FinalArgvLosesEverySelection(t *testing.T) {
	primary := entryFixture(t)
	first := filepath.Join(primary, ".moai", "worktrees", "first")
	mustMkdirTree(t, first)
	view := resolveLaunchView([]string{"-w", "first", "--", "--worktree=first"})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve")
	}
	rest := stripBackendWorktreeSelections(view.FinalArgv)
	for _, tok := range rest {
		if strings.Contains(tok, "worktree") || tok == "first" {
			t.Fatalf("a selection survived the backend-argv strip: %v", rest)
		}
	}
	// A non-selection tail token the backend owns survives the strip.
	view2 := resolveLaunchView([]string{"-w", "first", "--", "--keep", "prompt"})
	rest2 := stripBackendWorktreeSelections(view2.FinalArgv)
	joined := strings.Join(rest2, " ")
	if !strings.Contains(joined, "--keep") || !strings.Contains(joined, "prompt") {
		t.Fatalf("backend-owned tail tokens lost: %v", rest2)
	}
}

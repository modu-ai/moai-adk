package cli

// Regression tests for card t1791 (GH #1813): the -w NAME form must enter an
// existing same-name .moai/worktrees/<name> MoAI tree instead of passing the
// name through to claude, whose short-name resolution covers only
// .claude/worktrees/<name> — the pass-through used to create a duplicate
// branch tree silently (the 2026-10-08 lane-12 incident). Fixtures are
// scratch repositories under t.TempDir() with their own MOAI_HOME; the
// project-root seam is stubbed so nothing escapes the fixture.

import (
	"os"
	"os/exec"
	"path/filepath"
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

// TestEntryDir_LegacyMoaiTreeIsEntered — the #1813 fix: a short name whose
// only existing candidate is .moai/worktrees/<name> answers that directory
// (the launcher enters it by directory and strips the selection), never a
// pass-through that would let claude create .claude/worktrees/<name>.
func TestEntryDir_LegacyMoaiTreeIsEntered(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	moaiTree := filepath.Join(primary, ".moai", "worktrees", name)
	if err := os.MkdirAll(moaiTree, 0o755); err != nil {
		t.Fatalf("mkdir moai tree: %v", err)
	}
	got := existingWorktreeEntryDir([]string{"-w", name})
	// The resolver canonicalizes the primary root (EvalSymlinks) — on macOS
	// /var is an alias of /private/var, so compare canonical forms.
	want, err := filepath.EvalSymlinks(moaiTree)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", moaiTree, err)
	}
	if got != want {
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
		if err := os.MkdirAll(filepath.Join(primary, rel), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
	}
	if got := existingWorktreeEntryDir([]string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the two-root ambiguity", got)
	}
}

// TestEntryDir_LegacyClaudeStaysNative — a name whose only candidate is
// .claude/worktrees/<name> keeps answering "": claude re-enters that root
// natively, and the entry must not shadow it.
func TestEntryDir_LegacyClaudeStaysNative(t *testing.T) {
	primary := entryFixture(t)
	name := "native"
	if err := os.MkdirAll(filepath.Join(primary, ".claude", "worktrees", name), 0o755); err != nil {
		t.Fatalf("mkdir claude tree: %v", err)
	}
	if got := existingWorktreeEntryDir([]string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the legacy-claude native re-entry", got)
	}
}

// TestEntrySelection_StrippedOnEntry — the launcher-side half of the
// no-creation guarantee: when the entry resolves to a directory, the argv
// handed to the backend carries no worktree selection at all, so claude can
// never resolve a short name against its own .claude/worktrees and create a
// duplicate tree (the #1813 creation path is the pass-through this strips).
func TestEntrySelection_StrippedOnEntry(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	moaiTree := filepath.Join(primary, ".moai", "worktrees", name)
	if err := os.MkdirAll(moaiTree, 0o755); err != nil {
		t.Fatalf("mkdir moai tree: %v", err)
	}
	if existingWorktreeEntryDir([]string{"-w", name}) == "" {
		t.Fatal("precondition: the entry must resolve for the strip to be reachable")
	}
	rest := stripWorktreeSelection([]string{"-w", name, "-c", "extra"})
	for _, tok := range rest {
		if tok == name || tok == "-w" {
			t.Fatalf("worktree selection survived the strip: %v", rest)
		}
	}
}

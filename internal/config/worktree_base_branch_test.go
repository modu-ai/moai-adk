package config

// worktree_base_branch_test.go — SPEC-WORKTREE-BASEREF-001 M1 (card t313).
// AC-WBR-001: the schema carries git_strategy.worktree_base_branch, its default
// is the empty string, and an omitted key raises no error.

import (
	"os"
	"path/filepath"
	"testing"
)

// writeGitStrategySection writes a git-strategy.yaml carrying the given body
// under <root>/.moai/config/sections/ and returns the project root.
func writeGitStrategySection(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	return root
}

// TestGitStrategyWorktreeBaseBranchOmittedIsEmpty pins AC-WBR-001: a project
// whose git-strategy.yaml omits worktree_base_branch loads with the empty
// string and no error — the REQ-WBR-002 neutral default.
func TestGitStrategyWorktreeBaseBranchOmittedIsEmpty(t *testing.T) {
	root := writeGitStrategySection(t, "git_strategy:\n    mode: manual\n    provider: github\n")
	if got := LoadWorktreeBaseBranch(root); got != "" {
		t.Errorf("omitted key: got %q, want \"\" (REQ-WBR-002 neutral default)", got)
	}
}

// TestGitStrategyWorktreeBaseBranchLoadsConfiguredValue pins that a set value
// reaches the typed struct through the same section loader every other
// git_strategy key uses.
func TestGitStrategyWorktreeBaseBranchLoadsConfiguredValue(t *testing.T) {
	root := writeGitStrategySection(t, "git_strategy:\n    mode: manual\n    worktree_base_branch: develop\n")
	if got := LoadWorktreeBaseBranch(root); got != "develop" {
		t.Errorf("configured value: got %q, want \"develop\"", got)
	}
}

// TestGitStrategyWorktreeBaseBranchTrimsWhitespace pins acceptance.md §D.2's
// first edge case: a value with surrounding whitespace is trimmed before use,
// never written through untrimmed.
func TestGitStrategyWorktreeBaseBranchTrimsWhitespace(t *testing.T) {
	root := writeGitStrategySection(t, "git_strategy:\n    worktree_base_branch: \"  develop  \"\n")
	if got := LoadWorktreeBaseBranch(root); got != "develop" {
		t.Errorf("whitespace value: got %q, want \"develop\"", got)
	}
}

// TestGitStrategyWorktreeBaseBranchMissingFileIsEmpty pins the fail-open read:
// no config file at all yields the neutral default rather than an error.
func TestGitStrategyWorktreeBaseBranchMissingFileIsEmpty(t *testing.T) {
	if got := LoadWorktreeBaseBranch(t.TempDir()); got != "" {
		t.Errorf("missing file: got %q, want \"\"", got)
	}
}

// TestGitStrategyConfigCarriesWorktreeBaseBranchField pins the struct member
// itself, so the yaml tag cannot be silently dropped by a later edit.
func TestGitStrategyConfigCarriesWorktreeBaseBranchField(t *testing.T) {
	var gs GitStrategyConfig
	gs.WorktreeBaseBranch = "trunk"
	if gs.WorktreeBaseBranch != "trunk" {
		t.Fatal("GitStrategyConfig.WorktreeBaseBranch is not assignable")
	}
}

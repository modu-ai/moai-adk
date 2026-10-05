package settings

// worktree_base_branch_test.go — SPEC-WORKTREE-BASEREF-001 M5 (card t313).
// AC-WBR-011 schema half: the control is a free-text field that accepts any
// branch name. AC-WBR-014: saving it through the typed section path does not
// drop the keys the file carries but GitStrategyConfig does not model.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findFieldDef returns the registered FieldDef with the given name.
func findFieldDef(t *testing.T, name string) FieldDef {
	t.Helper()
	for _, f := range AllFields() {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("field %q is not registered in settings.AllFields()", name)
	return FieldDef{}
}

// TestWorktreeBaseBranchFieldTypeIsText pins the operator ruling recorded at
// plan.md §A D2.1. A TypeSelect with main / develop as its options would bake
// two repository-specific branch names into the SHIPPED schema, leaving a user
// whose default branch is `trunk` unable to select their own — the free text is
// not a shortcut taken for lack of a better widget, it is the only shape that
// stays neutral across downstream projects.
func TestWorktreeBaseBranchFieldTypeIsText(t *testing.T) {
	f := findFieldDef(t, "git_strategy.worktree_base_branch")

	if f.Type != TypeText {
		t.Errorf("Type = %q, want %q — a closed option set would not be template-neutral", f.Type, TypeText)
	}
	if len(f.Options) != 0 {
		t.Errorf("free-text field carries %d options; it must carry none", len(f.Options))
	}
	// An arbitrary third branch name must be accepted (or no predicate declared).
	if f.Validate != nil && !f.Validate("trunk") {
		t.Error("Validate rejects \"trunk\"; the field must accept any branch name")
	}
	if f.Validate != nil && !f.Validate("") {
		t.Error("Validate rejects the empty value; empty is the neutral take-no-action state")
	}
}

// TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys is AC-WBR-014.
//
// GitStrategyConfig's ModeProfile models no develop_branch /
// release_branch_prefix / rc_version_format field, and the typed save
// re-marshals the struct — so a naive write would silently drop three keys this
// repository depends on. This SPEC does not repair that schema gap, but its own
// write path must not newly expose it.
func TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "git-strategy.yaml")
	original := `git_strategy:
    mode: manual
    provider: github
    github_username: ""
    worktree_base_branch: ""
    manual:
        workflow: git-flow
        environment: local
        develop_branch: develop
        release_branch_prefix: release/
        rc_version_format: vX.Y.Z-rc.N
`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	field := findFieldDef(t, "git_strategy.worktree_base_branch")
	if err := applyTypedEdits(root, []FieldDef{field}, []string{"develop"}); err != nil {
		t.Fatalf("typed save: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	written := string(raw)

	if !strings.Contains(written, "worktree_base_branch: develop") &&
		!strings.Contains(written, `worktree_base_branch: "develop"`) {
		t.Errorf("the edited value did not reach the file:\n%s", written)
	}
	for _, key := range []string{"develop_branch", "release_branch_prefix", "rc_version_format"} {
		if !strings.Contains(written, key) {
			t.Errorf("typed save dropped the unmodelled key %q:\n%s", key, written)
		}
	}
}

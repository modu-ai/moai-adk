package template

// worktree_base_branch_neutrality_test.go — SPEC-WORKTREE-BASEREF-001 M1
// (card t313). AC-WBR-002 / REQ-WBR-003: the shipped git-strategy template
// carries worktree_base_branch with an EMPTY value.
//
// The prohibition binds the VALUE side of the colon only. A house-style
// comment naming main and develop as examples is permitted and is what the
// free-text control's description says in prose (REQ-WBR-014) — a shipped
// VALUE naming a branch is what would make the template repository-specific
// and unusable for a downstream project on `trunk` or `master`.
//
// This is the mechanical form of AC-WBR-002's grep (plan-audit iter-2 debt N5:
// the criterion's first command is human-read).

import (
	"os"
	"strings"
	"testing"
)

const gitStrategyTmplPath = "templates/.moai/config/sections/git-strategy.yaml.tmpl"

func TestWorktreeBaseBranchTemplateValueIsEmpty(t *testing.T) {
	raw, err := os.ReadFile(gitStrategyTmplPath)
	if err != nil {
		t.Fatalf("read %s: %v", gitStrategyTmplPath, err)
	}

	var found int
	for _, line := range strings.Split(string(raw), "\n") {
		// Strip any trailing comment before inspecting the value: the comment
		// side is explicitly allowed to name branches.
		code := line
		if i := strings.Index(code, "#"); i >= 0 {
			code = code[:i]
		}
		key, value, ok := strings.Cut(code, ":")
		if !ok || strings.TrimSpace(key) != "worktree_base_branch" {
			continue
		}
		found++
		got := strings.Trim(strings.TrimSpace(value), `"'`)
		if got != "" {
			t.Errorf("shipped template names a repository-specific branch %q for worktree_base_branch; "+
				"REQ-WBR-003 requires an empty value (the comment may name examples, the value may not)", got)
		}
	}

	if found != 1 {
		t.Fatalf("expected exactly 1 worktree_base_branch key in %s, found %d", gitStrategyTmplPath, found)
	}
}

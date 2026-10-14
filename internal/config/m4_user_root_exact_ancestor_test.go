package config

// m4_user_root_exact_ancestor_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, codex
// sync-audit finding 3 (rcpt-7e8e4000add18035a38e22c9). A user-root entry that names
// one file (user-root:claude-agents/plan-auditor.md) matched only that file: the
// directory holding it did not match, so rm -rf of the directory passed the exact
// arm. The dir and prefix forms already judge their ancestors (Gate round 39);
// the exact form must judge them too.

import "testing"

func TestZoneUserRootExactEntryCoversItsAncestorDirectory(t *testing.T) {
	e, err := parseZoneEntry("user-root:claude-agents/plan-auditor.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("the ancestor directory is covered in both spellings", func(t *testing.T) {
		for _, rel := range []string{"user-root:claude-agents", "user-root:claude-agents/"} {
			if !e.Match(FoldZoneText(rel)) {
				t.Errorf("the exact entry did not cover its ancestor directory %q", rel)
			}
		}
	})
	t.Run("a sibling file is not covered", func(t *testing.T) {
		if e.Match(FoldZoneText("user-root:claude-agents/other.md")) {
			t.Error("the exact entry covered a sibling file")
		}
	})
	t.Run("the file itself is covered", func(t *testing.T) {
		if !e.Match(FoldZoneText("user-root:claude-agents/plan-auditor.md")) {
			t.Error("the exact entry missed its own file")
		}
	})
	t.Run("prefix lookalikes and repository spellings are not covered", func(t *testing.T) {
		for _, rel := range []string{"user-root:claude-agents/plan", "user-root:claude-agentsx", "claude-agents/plan-auditor.md"} {
			if e.Match(FoldZoneText(rel)) {
				t.Errorf("the exact entry covered %q, want not covered", rel)
			}
		}
	})
}

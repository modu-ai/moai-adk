package template

import (
	"io/fs"
	"strings"
	"testing"
)

// userOwnedSkillPrefixes is the set of user-owned skill-name prefixes that
// MUST NOT appear in the embedded template FS: the canonical hns- generation
// plus the legacy harness- generation (dual-pattern per the hns- artifact
// prefix rename). NOTE: my-harness- is a substring superset of neither and is
// covered by the harness- check only when literally present; it is listed
// explicitly for completeness.
var userOwnedSkillPrefixes = []string{"hns-", "harness-", "my-harness-"}

// TestNamespaceLeakHarnessSkills enforces the user-owned namespace boundary
// for skills. The hns-* prefix (canonical) and the harness-* / my-harness-*
// prefixes (legacy generations) are reserved for user-generated artifacts
// emitted by the v4 harness Builder.
// These artifacts MUST NOT appear in the embedded template FS — leak indicates
// either a misclassification or accidental commit of user-area files.
//
// Sentinel: NAMESPACE_LEAK_HARNESS_SKILL
// Policy: CLAUDE.local.md §24.1 + .claude/rules/moai/development/skill-authoring.md § Skills Namespace Policy
// Origin: chore commit 4f1135684 (2026-05-23) — first enforcement after moai-adk-go
// domain specialists were leaked into template. Originally matched the legacy code-side prefix
// (the legacy code-side prefix); renamed to harness-* per the namespace catch-up,
// then extended with the hns-* canonical prefix (dual-pattern rejection).
func TestNamespaceLeakHarnessSkills(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var leaked []string
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == ".claude/skills" || !d.IsDir() {
			return nil
		}
		parts := strings.Split(path, "/")
		if len(parts) != 3 { // only top-level directories
			return nil
		}
		for _, prefix := range userOwnedSkillPrefixes {
			if strings.HasPrefix(parts[2], prefix) {
				leaked = append(leaked, path)
				break
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/skills) error: %v", walkErr)
	}

	if len(leaked) > 0 {
		t.Errorf("NAMESPACE_LEAK_HARNESS_SKILL: hns-* / harness-* / my-harness-* skills MUST NOT be embedded in template (user-owned namespace).\nLeaked entries: %v\nRemediation: remove from internal/template/templates/.claude/skills/ and from catalog.yaml.\nPolicy SSOT: CLAUDE.local.md §24.1, .claude/rules/moai/development/skill-authoring.md § Skills Namespace Policy.", leaked)
	}
}

// TestNamespaceLeakHarnessAgentsDir enforces the user-owned directory boundary
// for agents. The .claude/agents/harness/ directory is reserved for user-generated
// domain specialist agents emitted by moai-meta-harness during /moai project Phase 5+.
// The directory MUST NOT exist in the embedded template FS at all — its presence
// indicates either a misclassification or accidental commit of user-area files.
//
// Sentinel: NAMESPACE_LEAK_HARNESS_AGENTS_DIR
// Policy: CLAUDE.local.md §24.2 + .claude/rules/moai/development/agent-authoring.md § Agent Directory Convention
// Origin: chore commit 4f1135684 (2026-05-23) — first enforcement after 4 moai-adk-go
// domain specialists were leaked into template under .claude/agents/harness/.
func TestNamespaceLeakHarnessAgentsDir(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	entries, err := fs.ReadDir(fsys, ".claude/agents")
	if err != nil {
		t.Fatalf("ReadDir(.claude/agents) error: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == "harness" {
			harnessEntries, _ := fs.ReadDir(fsys, ".claude/agents/harness")
			files := make([]string, 0, len(harnessEntries))
			for _, h := range harnessEntries {
				files = append(files, h.Name())
			}
			t.Errorf("NAMESPACE_LEAK_HARNESS_AGENTS_DIR: .claude/agents/harness/ MUST NOT exist in template (user-owned directory).\nFound %d entries: %v\nRemediation: remove internal/template/templates/.claude/agents/harness/ entirely and remove entries from catalog.yaml.\nPolicy SSOT: CLAUDE.local.md §24.2, .claude/rules/moai/development/agent-authoring.md § Agent Directory Convention.", len(harnessEntries), files)
			return
		}
	}
}

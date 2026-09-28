package template

import (
	"io/fs"
	"testing"
)

// workflow_rule_paths_pinned_test.go: paths-scope preservation guard for the four
// workflow rules reduced under SPEC-INSTRUCTION-BUDGET-SCOPE-001 (card t1180).
//
// The reduction (compression + splits) must not narrow, widen, or otherwise alter
// any of these files' `paths:` loading globs: a narrowed glob silently unloads the
// rule from triggers it was written for; a widened glob silently loads it onto new
// triggers (the issue-#1059 shape). REQ-IBS-010 preserves the globs unless the SPEC
// is amended, so the expected values are RECORDED HERE rather than derived from the
// tree — a tree-derived expectation would tautologically pass and outlive its
// requirement as an unattributable stale guard after the card closes.
//
// Precedent: skill_authoring_paths_glob_test.go (frontmatter parse via
// parseFrontmatterAndBody). This guard reads the EMBEDDED template tree; local ↔
// template byte-identity is enforced separately by rule_template_mirror_test.go, so
// pinning the mirror pins the local copy transitively.

func TestWorkflowRulePathsPinned(t *testing.T) {
	t.Parallel()

	pinned := map[string]string{
		".claude/rules/moai/workflow/spec-workflow.md":            "**/.moai/specs/**,**/.moai/config/sections/quality.yaml",
		".claude/rules/moai/workflow/worktree-integration.md":     "**/.claude/agents/**,**/.claude/worktrees/**,**/.moai/worktrees/**,**/.claude/teams/**",
		".claude/rules/moai/workflow/session-handoff-examples.md": "**/session-handoff.md",
		".claude/rules/moai/workflow/kanban-dispatch-detail.md":   "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md",
	}

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for rulePath, want := range pinned {
		data, readErr := fs.ReadFile(fsys, rulePath)
		if readErr != nil {
			t.Errorf("ReadFile(%q) error: %v — rule missing from the template tree", rulePath, readErr)
			continue
		}
		fm, _, parseErr := parseFrontmatterAndBody(string(data))
		if parseErr != "" {
			t.Errorf("frontmatter parse failed for %s: %s", rulePath, parseErr)
			continue
		}
		got, ok := fm["paths"]
		if !ok || got == "" {
			t.Errorf("%s: missing or empty paths frontmatter field (was %q)", rulePath, got)
			continue
		}
		if got != want {
			t.Errorf("%s: paths scope drifted under reduction:\n got:  %q\n want: %q", rulePath, got, want)
		}
	}
}

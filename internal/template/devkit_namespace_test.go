package template

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// @MX:NOTE: [AUTO] devkit_namespace_test.go — SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001
// REQ-DHC-005 / AC-DHC-005 embedded-tree-absence guard for the devkit dev-only
// maintainer harness.
//
// The devkit harness (entry command, manifest, Runner, specialist sub-agents) is
// a DEV-ONLY maintainer asset in the user-owned harness namespace. It MUST NOT
// leak into internal/template/templates/ (the user-facing distribution tree).
//
// [D1 re-anchor] The actual CI protection for devkit artifacts is
// "absence-from-embedded-tree", modeled on embedded_namespace_test.go
// (TestTemplateAgentsStructure: {moai}-only allowlist on .claude/agents/). The
// dev_only_skill_test.go walker is .claude/skills/-only and therefore cannot
// protect commands/agents/workflows artifacts. This test fills that gap: it
// walks EmbeddedTemplates() and asserts the absence of every devkit artifact
// shape.
//
// Sentinel on failure: DEVKIT_NAMESPACE_LEAK
// Cross-reference: .moai/docs/dev-only-commands-isolation.md, CLAUDE.local.md §21/§2.
//
// RED/GREEN: planting a `.claude/commands/harness/` path (or a `harness-devkit*`
// agent, or a `.claude/workflows/harness-devkit-*` file) under
// internal/template/templates/ and running `make build` re-generates embedded.go
// with the leak compiled in — this test then FAILS (RED). Removing the leak and
// rebuilding restores PASS (GREEN).

// TestDevkitNamespaceNoLeak asserts that the embedded template tree contains
// NONE of the devkit harness artifact shapes:
//
//	(a) .claude/commands/harness/   path absent
//	(b) harness-devkit* agent files absent (anywhere under .claude/agents/)
//	(c) .claude/workflows/harness-devkit-* files absent
func TestDevkitNamespaceNoLeak(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error = %v, want nil", err)
	}

	walkErr := fs.WalkDir(fsys, ".claude", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			// .claude subtree may be partially absent in some build states; skip.
			return nil //nolint:nilerr // tolerate partial trees; absence is the success case
		}

		// (a) commands/harness/ path MUST NOT appear (the devkit entry command +
		//     manifest live there; the whole harness commands subdir is dev-only).
		if strings.Contains(filePath, ".claude/commands/harness") {
			t.Errorf(
				"DEVKIT_NAMESPACE_LEAK: dev-only harness command path %q found in embedded template tree. "+
					"`.claude/commands/harness/` is a dev-only maintainer namespace (SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 REQ-DHC-004) and MUST NOT be distributed.",
				filePath,
			)
		}

		// (c) workflows/harness-devkit-* files MUST NOT appear (the Runner).
		base := path.Base(filePath)
		if strings.Contains(filePath, ".claude/workflows") && strings.HasPrefix(base, "harness-devkit-") {
			t.Errorf(
				"DEVKIT_NAMESPACE_LEAK: dev-only devkit Runner %q found in embedded template tree. "+
					"`.claude/workflows/harness-devkit-*` is a dev-only maintainer asset and MUST NOT be distributed.",
				filePath,
			)
		}

		// (b) harness-devkit* agent files MUST NOT appear anywhere under
		//     .claude/agents/ (the specialist sub-agents).
		if strings.Contains(filePath, ".claude/agents") && strings.HasPrefix(base, "harness-devkit") {
			t.Errorf(
				"DEVKIT_NAMESPACE_LEAK: dev-only devkit specialist %q found in embedded template tree. "+
					"`harness-devkit*` agents are dev-only maintainer assets and MUST NOT be distributed.",
				filePath,
			)
		}

		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude) error = %v, want nil", walkErr)
	}
}

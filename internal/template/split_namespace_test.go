package template

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// @MX:NOTE: [AUTO] split_namespace_test.go — SPEC-V3R6-DEV-HARNESS-SPLIT-001
// REQ-DHS-004 / REQ-DHS-005 embedded-tree-absence guard for the three split
// dev-only maintainer harnesses (release-update / github / release).
//
// The split harnesses (thin command entries, the release-update manifest +
// Runner, the three specialist sub-agents) are DEV-ONLY maintainer assets in the
// user-owned harness namespace. They MUST NOT leak into
// internal/template/templates/ (the user-facing distribution tree).
//
// History: this guard replaces the SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001
// unified-harness namespace guard. That consolidation's single unified harness
// was split into three independent harnesses by SPEC-V3R6-DEV-HARNESS-SPLIT-001,
// so the guard now asserts the three split artifact prefixes are absent instead
// of the single legacy unified-harness prefix.
//
// The actual CI protection for these artifacts is "absence-from-embedded-tree",
// modeled on embedded_namespace_test.go (TestTemplateAgentsStructure:
// {moai}-only allowlist on .claude/agents/). The dev_only_skill_test.go walker is
// .claude/skills/-only and therefore cannot protect commands/agents/workflows
// artifacts. This test fills that gap: it walks EmbeddedTemplates() and asserts
// the absence of every split-harness artifact shape.
//
// Sentinel on failure: SPLIT_HARNESS_NAMESPACE_LEAK
// Cross-reference: .moai/docs/dev-only-commands-isolation.md, CLAUDE.local.md §21/§2.
//
// RED/GREEN: planting a `.claude/commands/harness/` path (or a
// `harness-{release-update,github,release}*` agent, or a
// `.claude/workflows/harness-{release-update,github,release}-*` file) under
// internal/template/templates/ and running `make build` re-generates embedded.go
// with the leak compiled in — this test then FAILS (RED). Removing the leak and
// rebuilding restores PASS (GREEN).

// splitHarnessAgentPrefixes is the canonical set of dev-only split-harness
// artifact-name prefixes that MUST NOT appear under .claude/agents/ or
// .claude/workflows/ in the embedded template tree.
var splitHarnessAgentPrefixes = []string{
	"harness-release-update",
	"harness-github",
	"harness-release",
}

// TestSplitHarnessNamespaceNoLeak asserts that the embedded template tree
// contains NONE of the split-harness artifact shapes:
//
//	(a) .claude/commands/harness/   path absent (thin commands + release-update manifest)
//	(b) harness-{release-update,github,release}* agent files absent under .claude/agents/
//	(c) .claude/workflows/harness-{release-update,github,release}-* files absent (the Runner)
func TestSplitHarnessNamespaceNoLeak(t *testing.T) {
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

		// (a) commands/harness/ path MUST NOT appear (the split thin commands +
		//     the release-update manifest live there; the whole harness commands
		//     subdir is dev-only).
		if strings.Contains(filePath, ".claude/commands/harness") {
			t.Errorf(
				"SPLIT_HARNESS_NAMESPACE_LEAK: dev-only harness command path %q found in embedded template tree. "+
					"`.claude/commands/harness/` is a dev-only maintainer namespace (SPEC-V3R6-DEV-HARNESS-SPLIT-001 REQ-DHS-005) and MUST NOT be distributed.",
				filePath,
			)
		}

		base := path.Base(filePath)

		// (c) workflows/harness-{release-update,github,release}-* files MUST NOT
		//     appear (the Runner; only release-update has one, but all three
		//     prefixes are guarded for completeness).
		if strings.Contains(filePath, ".claude/workflows") {
			for _, prefix := range splitHarnessAgentPrefixes {
				if strings.HasPrefix(base, prefix+"-") {
					t.Errorf(
						"SPLIT_HARNESS_NAMESPACE_LEAK: dev-only split-harness Runner %q found in embedded template tree. "+
							"`.claude/workflows/%s-*` is a dev-only maintainer asset and MUST NOT be distributed.",
						filePath, prefix,
					)
					break
				}
			}
		}

		// (b) harness-{release-update,github,release}* agent files MUST NOT appear
		//     anywhere under .claude/agents/ (the specialist sub-agents).
		if strings.Contains(filePath, ".claude/agents") {
			for _, prefix := range splitHarnessAgentPrefixes {
				if strings.HasPrefix(base, prefix) {
					t.Errorf(
						"SPLIT_HARNESS_NAMESPACE_LEAK: dev-only split-harness specialist %q found in embedded template tree. "+
							"`%s*` agents are dev-only maintainer assets and MUST NOT be distributed.",
						filePath, prefix,
					)
					break
				}
			}
		}

		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude) error = %v, want nil", walkErr)
	}
}

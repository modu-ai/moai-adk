package template

import (
	"io/fs"
	"strings"
	"testing"
)

// @MX:NOTE: [AUTO] embedded_namespace_test.go — SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001 REQ-HNC-005 regression guard
// @MX:SPEC: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001 REQ-HNC-001/005
//
// These tests statically verify the CLAUDE.local.md §24 "Harness Namespace 분리 정책"
// invariants on the embedded template filesystem:
//
//  1. internal/template/templates/.claude/agents/ MUST contain ONLY the
//     canonical FLAT system subdirectory {moai} — NO user-domain directories
//     such as `harness/` (which belongs to USER projects post-/moai project
//     interview, not to the moai-adk template baseline).
//
//  2. internal/template/templates/.claude/skills/ MUST contain at most the
//     canonical `moai-harness-*` allowlist {moai-harness-learner}. Any other
//     `moai-harness-*` prefixed directory (e.g. moai-harness-cli-template,
//     moai-harness-patterns) signals a §24 contract violation — those names
//     must use the `harness-*` prefix and live in the user project, not
//     in the template baseline.
//
// Sentinel on failure: HARNESS_NAMESPACE_LEAK
// Cross-reference: CLAUDE.local.md §24, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
//
// These tests prevent regression: if a future `make build` regenerates
// embedded.go with leaked artifacts re-introduced under templates/.claude/,
// the build will fail at `go test ./internal/template/...` time.

// TestTemplateAgentsStructure verifies REQ-HNC-005 invariant 1:
// internal/template/templates/.claude/agents/ contains exactly the canonical
// FLAT system subdirectory {moai} and no others.
func TestTemplateAgentsStructure(t *testing.T) {
	t.Parallel()

	tmplFS, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error = %v, want nil", err)
	}

	entries, err := fs.ReadDir(tmplFS, ".claude/agents")
	if err != nil {
		t.Fatalf("fs.ReadDir(.claude/agents) error = %v, want nil", err)
	}

	// Post SPEC-V3R6-AGENT-TEAM-REBUILD-001 (+ V2-V3-CLEAN-REINSTALL-001): the
	// agent catalog consolidated to 7 retained MoAI-custom agents organized in a
	// single FLAT subdir {moai}. The earlier {core, expert, meta} split was
	// superseded (FLAT layout restored as canonical). REQ-MRR-001: enumeration
	// updated to current retained catalog reality.
	expected := map[string]bool{
		"moai": true,
	}

	actual := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			actual[e.Name()] = true
		}
	}

	// Detect missing expected subdirs.
	for name := range expected {
		if !actual[name] {
			t.Errorf("expected .claude/agents/%s/ subdir missing from embedded template", name)
		}
	}

	// Detect unexpected (leaked) subdirs.
	for name := range actual {
		if !expected[name] {
			t.Errorf("unexpected .claude/agents/%s/ subdir in embedded template — only {moai} allowed post SPEC-V3R6-AGENT-TEAM-REBUILD-001 FLAT layout (HARNESS_NAMESPACE_LEAK)", name)
		}
	}
}

// TestTemplateMoaiHarnessSkillsAllowlist verifies REQ-HNC-005 invariant 2:
// internal/template/templates/.claude/skills/ may contain `moai-harness-*`
// prefixed directories ONLY if they match the canonical allowlist
// {moai-harness-learner}.
//
// Note: `moai-meta-harness` uses the `moai-meta-*` prefix (NOT `moai-harness-*`)
// so it is NOT subject to this test. The strict `moai-harness-*` allowlist
// is intentionally minimal — `moai-harness-learner` is the only template-shipped
// skill carrying the `moai-harness-*` prefix.
func TestTemplateMoaiHarnessSkillsAllowlist(t *testing.T) {
	t.Parallel()

	tmplFS, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error = %v, want nil", err)
	}

	entries, err := fs.ReadDir(tmplFS, ".claude/skills")
	if err != nil {
		t.Fatalf("fs.ReadDir(.claude/skills) error = %v, want nil", err)
	}

	allowed := map[string]bool{
		"moai-harness-learner": true,
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "moai-harness-") {
			continue
		}
		if !allowed[name] {
			t.Errorf("unexpected moai-harness-* skill in embedded template: %q — only {moai-harness-learner} allowed per CLAUDE.local.md §24.1 (HARNESS_NAMESPACE_LEAK)", name)
		}
	}
}

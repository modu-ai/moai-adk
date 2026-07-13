// SPEC-HNS-PREFIX-RENAME-001 M2 — RED-phase specification tests.
//
// These tests pin the NEW canonical user-owned prefix `hns-` introduced by the
// harness user-artifact prefix rename:
//   - .claude/skills/hns-*        (canonical user harness skill directories)
//   - .claude/workflows/hns-*.js  (canonical user Runner Workflows)
//
// Both surfaces MUST be recognized as user-owned by plan.IsUserAreaPath AND
// plan.IsUserOwnedNamespace, IN ADDITION TO the two legacy generations
// (harness-*, my-harness-*), so `moai update` preserves all three generations
// (tri-generation recognition, REQ-HPR-005..008).
//
// Written BEFORE the M2 GREEN-phase extension — the hns- positive cases FAIL
// initially because the classifiers do not yet recognize the hns- prefix.
//
// @MX:NOTE: [AUTO] AC-HPR-003/AC-HPR-004 specification — pins skills/hns-* +
// workflows/hns-*.js user-owned with exact HasPrefix matching (REQ-HPR-008).
// @MX:SPEC: SPEC-HNS-PREFIX-RENAME-001 acceptance.md AC-HPR-003, AC-HPR-004
package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
)

// TestUserOwnedNamespace_HNS pins REQ-HPR-005/REQ-HPR-008 on the authoritative
// plan.IsUserOwnedNamespace classifier.
func TestUserOwnedNamespace_HNS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// Canonical hns- generation (NEW)
		{"hns skill SKILL.md", ".claude/skills/hns-x/SKILL.md", true},
		{"hns skill nested file", ".claude/skills/hns-acme-verify/references/notes.md", true},
		{"hns workflow run.js", ".claude/workflows/hns-x-run.js", true},
		{"hns workflow windows separator", ".claude\\workflows\\hns-x-run.js", true},
		{"hns skill windows separator", ".claude\\skills\\hns-x\\SKILL.md", true},

		// Legacy generations continue to be recognized (no regression)
		{"legacy harness skill", ".claude/skills/harness-x/SKILL.md", true},
		{"legacy my-harness skill", ".claude/skills/my-harness-x/SKILL.md", true},
		{"legacy harness workflow", ".claude/workflows/harness-x-run.js", true},

		// Template-managed moai-harness-* MUST NOT be user-owned (REQ-HPR-008)
		{"moai-harness-learner template-managed", ".claude/skills/moai-harness-learner/SKILL.md", false},

		// Exact-prefix discipline for workflows: hnsx-*.js is NOT a recognized
		// Runner prefix (byte-exact `hns-` HasPrefix, edge case AC §E.1).
		{"hnsx workflow not recognized", ".claude/workflows/hnsx-foo.js", false},

		// Case sensitivity: upper/mixed-case variants are NOT recognized (§E.6).
		{"uppercase prefix workflow not recognized", ".claude/workflows/HNS-x-run.js", false},

		// NOTE: .claude/skills/hnsx-foo/ IS user-owned via the pre-existing
		// REQ-UNP-009 general custom-skill rule (any non-moai skill dir), NOT via
		// the hns- prefix branch. The hns--specific exact-prefix negative is
		// asserted on plan.IsUserAreaPath below (which has no general custom rule).
		{"hnsx skill user-owned via REQ-UNP-009 general rule", ".claude/skills/hnsx-foo/SKILL.md", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := plan.IsUserOwnedNamespace(tt.rel)
			if got != tt.want {
				t.Errorf("plan.IsUserOwnedNamespace(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

// TestUserAreaPath_HNS pins the same hns- surfaces on the plan.IsUserAreaPath guard
// (used by cleanMoaiManagedPaths + the overlay write loop). Unlike
// plan.IsUserOwnedNamespace, this guard has NO general custom-skill rule, so it is
// the surface where the exact `hns-` HasPrefix negative (hnsx-foo) is provable.
func TestUserAreaPath_HNS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// Canonical hns- generation (NEW)
		{"hns skill SKILL.md", ".claude/skills/hns-x/SKILL.md", true},
		{"hns workflow run.js", ".claude/workflows/hns-x-run.js", true},
		{"hns skill windows separator", ".claude\\skills\\hns-x\\SKILL.md", true},

		// Legacy generations (no regression)
		{"legacy harness skill", ".claude/skills/harness-x/SKILL.md", true},
		{"legacy my-harness skill", ".claude/skills/my-harness-x/SKILL.md", true},
		{"legacy harness workflow", ".claude/workflows/harness-x-run.js", true},

		// Exact-prefix negatives (AC-HPR-003 edge cases §E.1)
		{"hnsx skill not hns- prefixed", ".claude/skills/hnsx-foo/SKILL.md", false},
		{"hnsfoo skill not hns- prefixed", ".claude/skills/hnsfoo/SKILL.md", false},
		{"hnsx workflow not hns- prefixed", ".claude/workflows/hnsx-foo.js", false},

		// Template-managed moai-harness-* never user-area (REQ-HPR-008)
		{"moai-harness-learner template-managed", ".claude/skills/moai-harness-learner/SKILL.md", false},

		// Case sensitivity (§E.6): byte-exact matching only.
		{"uppercase prefix skill not recognized", ".claude/skills/HNS-x/SKILL.md", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := plan.IsUserAreaPath(tt.rel)
			if got != tt.want {
				t.Errorf("plan.IsUserAreaPath(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

// TestUpdateNamespaceHNS_TriGenerationPreservation is the AC-HPR-004 E2E
// sandbox: plant artifacts of ALL THREE prefix generations (hns-*, harness-*,
// my-harness-*) plus agents under .claude/agents/harness/, run the update
// stale-file removal flow (cleanMoaiManagedPaths), and assert every planted
// artifact survives byte-identical (zero deletions, zero modifications).
//
// Pattern precedent: TestPreserveMyHarnessOnUpdate (update_preserve_my_harness_test.go).
func TestUpdateNamespaceHNS_TriGenerationPreservation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	type planted struct {
		rel     string
		content string
	}
	artifacts := []planted{
		// hns- canonical generation
		{".claude/skills/hns-acme-verify/SKILL.md", "hns skill content"},
		{".claude/workflows/hns-acme-run.js", "// hns runner\n"},
		// harness- legacy gen-2
		{".claude/skills/harness-acme-verify/SKILL.md", "harness skill content"},
		{".claude/workflows/harness-acme-run.js", "// harness runner\n"},
		// my-harness- legacy gen-3
		{".claude/skills/my-harness-old/SKILL.md", "my-harness skill content"},
		// agents directory (generation-agnostic, directory-level preservation)
		{".claude/agents/harness/hns-acme-core-specialist.md", "hns specialist"},
		{".claude/agents/harness/harness-acme-core-specialist.md", "harness specialist"},
	}

	pre := map[string]string{}
	for _, a := range artifacts {
		abs := filepath.Join(tmpDir, filepath.FromSlash(a.rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte(a.content), 0o644); err != nil {
			t.Fatalf("write %s: %v", abs, err)
		}
		pre[a.rel] = fileSHA256(t, abs)
	}

	// Exercise the stale-file removal step of moai update.
	if err := cleanMoaiManagedPaths(tmpDir, io.Discard); err != nil {
		t.Fatalf("cleanMoaiManagedPaths: %v", err)
	}

	// Every planted artifact survives byte-identical.
	for _, a := range artifacts {
		abs := filepath.Join(tmpDir, filepath.FromSlash(a.rel))
		if _, err := os.Stat(abs); err != nil {
			t.Errorf("%s removed by cleanMoaiManagedPaths (must be preserved): %v", a.rel, err)
			continue
		}
		if post := fileSHA256(t, abs); post != pre[a.rel] {
			t.Errorf("%s content changed across update flow (must be byte-identical)", a.rel)
		}
	}
}

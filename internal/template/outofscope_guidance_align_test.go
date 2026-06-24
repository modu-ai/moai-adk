// outofscope_guidance_align_test.go: re-drift guard for the SPEC-authoring
// exclusions guidance. The manager-spec agent template MUST instruct authors to
// use the `### Out of Scope — <topic>` H3 form (which satisfies the
// `OutOfScopeRule` lint in internal/spec/lint.go), and MUST NOT regress to the
// bare `## Exclusions (What NOT to Build)` H2 mandate that the lint rule rejects.
//
// Origin: SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001 M5 (optional CI re-drift guard).
package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestOutOfScopeGuidanceAligned verifies that the manager-spec template guidance
// directs SPEC authors to the `### Out of Scope` convention and has not regressed
// to the bare H2 exclusions mandate that fails OutOfScopeRule.
func TestOutOfScopeGuidanceAligned(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerSpecPath = ".claude/agents/moai/manager-spec.md"

	data, readErr := fs.ReadFile(fsys, managerSpecPath)
	if readErr != nil {
		t.Fatalf("failed to read manager-spec.md from embedded FS: %v", readErr)
	}
	body := string(data)

	// Positive: the `### Out of Scope` H3 authoring-guidance form must be present.
	if !strings.Contains(body, "### Out of Scope") {
		t.Error("manager-spec.md guidance must instruct authors to use the " +
			"`### Out of Scope — <topic>` H3 form (satisfies OutOfScopeRule); " +
			"the `### Out of Scope` string is absent")
	}

	// Negative (re-drift guard): the bare H2 exclusions mandate must NOT reappear.
	// This is the exact shape that fails OutOfScopeRule (no `out of scope` text,
	// no `###` heading). It previously sat in a `- [HARD] Every spec.md MUST
	// include` bullet at manager-spec.md line 84.
	if strings.Contains(body, "MUST include `## Exclusions (What NOT to Build)`") {
		t.Error("manager-spec.md guidance regressed to the bare " +
			"`## Exclusions (What NOT to Build)` H2 mandate, which fails " +
			"OutOfScopeRule (MissingExclusions). Use the `### Out of Scope — " +
			"<topic>` H3 form instead.")
	}
}

// manager_develop_present_test.go: verifies that manager-develop.md exists in the embedded FS.
// REQ-RA-003 mapping.
//
// Originally (SPEC-V3R2-RT-005 M1-M2): verified manager-develop.md is a full active agent.
// Updated (SPEC-V3R2-ORC-001 M5): manager-develop was a transitional name; ORC-001
// canonizes manager-cycle as the unified DDD/TDD agent name and retires manager-develop.
// Tests now verify the retirement stub pattern (consistent with TestRetirementCompletenessAssertion).
package template

import (
	"io/fs"
	"testing"
)

// TestManagerDevelopRetiredStubPresent verifies that manager-develop.md exists in the
// embedded FS as a properly formed retirement stub.
//
// SPEC-V3R2-ORC-001: manager-develop (transitional name from RT-005) was retired in favor
// of the canonical manager-cycle name. The stub must be present for backward compatibility.
func TestManagerDevelopRetiredStubPresent(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	// Verify file presence
	_, statErr := fs.Stat(fsys, managerDevelopPath)
	if statErr != nil {
		t.Fatalf("manager-develop.md not found in embedded FS. "+
			"Retirement stub must be present for backward compatibility (SPEC-V3R2-ORC-001): %v", statErr)
	}
}

// TestManagerDevelopIsRetiredStub verifies that manager-develop.md's frontmatter
// correctly identifies it as a retired stub pointing to manager-cycle.
//
// SPEC-V3R2-ORC-001: manager-develop must have retired: true and retired_replacement: manager-cycle.
func TestManagerDevelopIsRetiredStub(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	data, readErr := fs.ReadFile(fsys, managerDevelopPath)
	if readErr != nil {
		t.Fatalf("failed to read manager-develop.md: %v", readErr)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("manager-develop.md frontmatter parse error: %s", parseErr)
	}

	// retired: field must be present with value "true"
	retiredVal, hasRetired := fm["retired"]
	if !hasRetired {
		t.Error("manager-develop.md must have 'retired: true' frontmatter field (SPEC-V3R2-ORC-001)")
	} else if retiredVal != "true" {
		t.Errorf("manager-develop.md retired field must be 'true', got: %q", retiredVal)
	}

	// retired_replacement must point to manager-cycle
	replacement, hasReplacement := fm["retired_replacement"]
	if !hasReplacement {
		t.Error("manager-develop.md must have 'retired_replacement' frontmatter field")
	} else if replacement != "manager-cycle" {
		t.Errorf("manager-develop.md retired_replacement must be 'manager-cycle', got: %q", replacement)
	}

	// tools must be empty
	tools, hasTools := fm["tools"]
	if hasTools && tools != "" && tools != "[]" {
		t.Errorf("manager-develop.md retired stub must have empty tools list, got: %q", tools)
	}

	// name field must match
	if name, ok := fm["name"]; ok && name != "manager-develop" {
		t.Errorf("manager-develop.md name field must be 'manager-develop', got: %q", name)
	}
}

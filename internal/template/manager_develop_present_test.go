// manager_develop_present_test.go: verifies that manager-develop.md exists in the embedded FS
// as the active unified DDD/TDD implementation agent.
// REQ-RA-003 mapping.
//
// History:
//
//	SPEC-V3R2-RT-005 M1-M2: manager-develop.md was a full active agent (transitional name).
//	SPEC-V3R2-ORC-001 M5: renamed to manager-cycle; manager-develop became a retired stub.
//	ORC-001 follow-up rename: manager-develop restored as the canonical active agent name;
//	manager-cycle is now the retired stub.
package template

import (
	"io/fs"
	"testing"
)

// TestManagerDevelopActiveAgentPresent verifies that manager-develop.md exists in the
// embedded FS as a properly formed active agent definition.
//
// ORC-001 follow-up rename: manager-develop is the canonical active unified DDD/TDD agent.
// manager-cycle is now a retired stub pointing to manager-develop.
func TestManagerDevelopActiveAgentPresent(t *testing.T) {
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
			"Active agent must be present (ORC-001 follow-up rename): %v", statErr)
	}
}

// TestManagerDevelopIsActiveAgent verifies that manager-develop.md's frontmatter
// correctly identifies it as the active agent (not a retired stub).
//
// ORC-001 follow-up rename: manager-develop must NOT have retired: true and must
// have the correct name field, tools, and hooks.
func TestManagerDevelopIsActiveAgent(t *testing.T) {
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

	// retired: field must NOT be present (or must be false)
	if retiredVal, hasRetired := fm["retired"]; hasRetired && retiredVal == "true" {
		t.Error("manager-develop.md must NOT have 'retired: true' — it is the active agent (ORC-001 follow-up rename)")
	}

	// name field must match
	if name, ok := fm["name"]; !ok {
		t.Error("manager-develop.md must have a 'name' frontmatter field")
	} else if name != "manager-develop" {
		t.Errorf("manager-develop.md name field must be 'manager-develop', got: %q", name)
	}

	// tools must be non-empty (active agent has tools)
	tools, hasTools := fm["tools"]
	if !hasTools || tools == "" || tools == "[]" {
		t.Error("manager-develop.md active agent must have a non-empty tools list")
	}
}

// TestManagerCycleIsRetiredStub verifies that manager-cycle.md's frontmatter
// correctly identifies it as a retired stub pointing to manager-develop.
//
// ORC-001 follow-up rename: manager-cycle must have retired: true and
// retired_replacement: manager-develop.
func TestManagerCycleIsRetiredStub(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerCyclePath = ".claude/agents/moai/manager-cycle.md"

	data, readErr := fs.ReadFile(fsys, managerCyclePath)
	if readErr != nil {
		t.Fatalf("failed to read manager-cycle.md: %v", readErr)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("manager-cycle.md frontmatter parse error: %s", parseErr)
	}

	// retired: field must be present with value "true"
	retiredVal, hasRetired := fm["retired"]
	if !hasRetired {
		t.Error("manager-cycle.md must have 'retired: true' frontmatter field (ORC-001 follow-up rename)")
	} else if retiredVal != "true" {
		t.Errorf("manager-cycle.md retired field must be 'true', got: %q", retiredVal)
	}

	// retired_replacement must point to manager-develop
	replacement, hasReplacement := fm["retired_replacement"]
	if !hasReplacement {
		t.Error("manager-cycle.md must have 'retired_replacement' frontmatter field")
	} else if replacement != "manager-develop" {
		t.Errorf("manager-cycle.md retired_replacement must be 'manager-develop', got: %q", replacement)
	}

	// tools must be empty
	tools, hasTools := fm["tools"]
	if hasTools && tools != "" && tools != "[]" {
		t.Errorf("manager-cycle.md retired stub must have empty tools list, got: %q", tools)
	}

	// name field must match
	if name, ok := fm["name"]; ok && name != "manager-cycle" {
		t.Errorf("manager-cycle.md name field must be 'manager-cycle', got: %q", name)
	}
}

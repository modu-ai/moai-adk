// manager_develop_present_test.go: verifies that manager-develop.md exists in the embedded FS.
// REQ-RA-003 mapping.
//
// M1 RED phase: manager-develop.md is not yet in the embedded FS, so this intentionally FAILs.
// It becomes GREEN after manager-develop.md is added in M2.
package template

import (
	"io/fs"
	"testing"
)

// TestManagerDevelopPresentInEmbeddedFS verifies that manager-develop.md
// exists in the embedded FS and meets the minimum size requirement.
//
// REQ-RA-003: manager-develop.md must be included in the embedded FS after make build.
// Expected RED: file absent, FAIL (GREEN at M2)
func TestManagerDevelopPresentInEmbeddedFS(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	// Verify file presence
	info, statErr := fs.Stat(fsys, managerDevelopPath)
	if statErr != nil {
		t.Fatalf("RETIREMENT_INCOMPLETE_manager-tdd: manager-develop.md not found in embedded FS. "+
			"Must be added in SPEC-V3R3-RETIRED-AGENT-001 M2 (REQ-RA-003): %v", statErr)
	}

	// Size sanity check: minimum 5000 bytes (validates full agent definition)
	const minSize = 5000
	if info.Size() < minSize {
		t.Errorf("manager-develop.md is too small: %d bytes (minimum %d bytes required). "+
			"May not be a complete agent definition file (REQ-RA-003)", info.Size(), minSize)
	}
}

// TestManagerDevelopFrontmatterValid verifies that manager-develop.md's frontmatter
// contains all required fields.
//
// REQ-RA-001: manager-develop.md must have complete frontmatter.
// Expected RED: file itself absent, FAIL (file read fails — coupled with TestManagerDevelopPresentInEmbeddedFS)
func TestManagerDevelopFrontmatterValid(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	data, readErr := fs.ReadFile(fsys, managerDevelopPath)
	if readErr != nil {
		t.Fatalf("failed to read manager-develop.md (file absent — must be added in M2): %v", readErr)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("manager-develop.md frontmatter parse error: %s", parseErr)
	}

	// Validate required frontmatter fields (REQ-RA-001: full frontmatter)
	requiredFields := []string{
		"name",
		"description",
		"tools",
		"model",
	}
	for _, field := range requiredFields {
		val, ok := fm[field]
		if !ok || val == "" {
			t.Errorf("manager-develop.md frontmatter missing or empty required field '%s' (REQ-RA-001)", field)
		}
	}

	// retired: field must not be present (active agent)
	if _, hasRetired := fm["retired"]; hasRetired {
		t.Errorf("manager-develop.md must be an active agent and must not have a 'retired:' field (REQ-RA-001)")
	}

	// Verify the name field equals manager-develop
	if name, ok := fm["name"]; ok && name != "manager-develop" {
		t.Errorf("manager-develop.md name field must be 'manager-develop', got: %q", name)
	}
}

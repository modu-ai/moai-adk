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

	const managerDevelopPath = ".claude/agents/core/manager-develop.md"

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

	const managerDevelopPath = ".claude/agents/core/manager-develop.md"

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

// TestPurgedZombieAgentsAbsent verifies that the 8 retired/placeholder zombie
// agents have been fully purged from the embedded FS (no longer retained as
// stubs).
//
// Workflow audit 2026-05-16 Bundle C + audit recommendation finding F-003:
// the zombie agents below were absorbed into active replacements
// (manager-develop / manager-quality / expert-performance / builder-harness)
// and the placeholder stubs were removed. This test locks in the post-purge
// contract: if any zombie reappears in templates, regression is caught.
func TestPurgedZombieAgentsAbsent(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	purged := []string{
		"builder-agent",
		"builder-plugin",
		"builder-skill",
		"expert-debug",
		"expert-testing",
		"manager-cycle",
		"manager-ddd",
		"manager-tdd",
	}

	// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: scan all 4 domain subfolders for zombie regression.
	domains := []string{"core", "expert", "meta", "harness"}
	for _, name := range purged {
		for _, domain := range domains {
			path := ".claude/agents/" + domain + "/" + name + ".md"
			if _, statErr := fs.Stat(fsys, path); statErr == nil {
				t.Errorf("ZOMBIE_AGENT_REGRESSION: %s reappeared in embedded FS — must remain purged (Bundle C / F-003)", path)
			}
		}
	}
}

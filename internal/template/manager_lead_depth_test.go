package template_test

// TestManagerLeadDepthSeal verifies the depth-2 hierarchical-team seal:
//
//  1. `.claude/agents/moai/manager-lead.md` is the SOLE retained agent whose
//     `tools:` field contains the literal token `Agent` — it is the depth-1
//     carrier that opens the fan-out seam exactly one layer deep.
//  2. Every OTHER retained agent under `.claude/agents/moai/*.md` MUST omit
//     `Agent` from `tools:` — the flat-hierarchy invariant preserved by tool
//     omission (every retained MoAI agent other than manager-lead).
//  3. Any future leaf-worker agent file that declares itself a manager-lead
//     spawn target (frontmatter `leaf_of: manager-lead` OR body marker
//     `<!-- manager-lead leaf-worker -->`) MUST omit `Agent` from `tools:`.
//     Depth-2, no further recursion — the seal binds at lint time.
//
// On violation, sentinel: DEPTH_SEAL_VIOLATION.
//
// Rationale: the flat-hierarchy guarantee historically rested on every retained
// agent omitting `Agent` from `tools:`. manager-lead is the sole carve-out
// (depth-1 carrier). The CI test catches (a) a future agent file that adds
// `Agent` to `tools:` without being manager-lead, (b) a future leaf-worker
// file that nests further by carrying `Agent`, and (c) a regression that
// removes `Agent` from manager-lead's own `tools:` list.
//
// Cross-references:
//   - .claude/agents/moai/manager-lead.md § Depth-2 Seal
//   - .claude/rules/moai/core/agent-common-protocol.md § Background Agent Execution
//   - CLAUDE.md §4 Watch note (subagent nesting enabled by default as of recent
//     Claude Code versions; the seal is a MoAI policy invariant, not a runtime one).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestManagerLeadIsSoleAgentCarrier asserts that manager-lead.md is the only
// retained agent under .claude/agents/moai/ whose tools: field contains Agent.
func TestManagerLeadIsSoleAgentCarrier(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootDepth(t)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "moai")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Skipf(".claude/agents/moai/ not found at %s; skipping depth-seal audit", agentsDir)
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", agentsDir, err)
	}

	carriers := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(agentsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Read %s: %v", path, err)
		}
		fm := extractFrontmatter(data)
		toolsLine := frontmatterFieldValue(fm, "tools")
		if toolsLine == "" {
			continue
		}
		if hasAgentToken(toolsLine) {
			carriers = append(carriers, entry.Name())
		}
	}

	// Verify exactly one carrier and that it is manager-lead.md.
	if len(carriers) != 1 {
		t.Errorf(
			"DEPTH_SEAL_VIOLATION: expected exactly 1 agent under .claude/agents/moai/ "+
				"carrying Agent in tools: (manager-lead.md); found %d carriers: %v. "+
				"manager-lead is the sole depth-1 carrier; every other retained agent "+
				"preserves the flat hierarchy by tool omission.",
			len(carriers), carriers,
		)
		return
	}
	if carriers[0] != "manager-lead.md" {
		t.Errorf(
			"DEPTH_SEAL_VIOLATION: the sole Agent-carrier under .claude/agents/moai/ "+
				"must be manager-lead.md; found %s. Either rename the carrier or remove "+
				"Agent from its tools: list.",
			carriers[0],
		)
	}
}

// TestManagerLeadCarriesAgent asserts that manager-lead.md itself carries the
// Agent token in its tools: list — the depth-1 carrier invariant. A regression
// that removes Agent from manager-lead's tools would silently close the fan-out
// seam and defeat the agent's purpose.
func TestManagerLeadCarriesAgent(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootDepth(t)
	leadPath := filepath.Join(projectRoot, ".claude", "agents", "moai", "manager-lead.md")

	data, err := os.ReadFile(leadPath)
	if err != nil {
		t.Skipf("manager-lead.md not found at %s; skipping (test ships alongside the agent file)", leadPath)
	}

	fm := extractFrontmatter(data)
	toolsLine := frontmatterFieldValue(fm, "tools")
	if toolsLine == "" {
		t.Fatalf(
			"DEPTH_SEAL_VIOLATION: manager-lead.md has no tools: field in frontmatter; "+
				"the depth-1 carrier MUST declare tools: with Agent present.",
		)
	}
	if !hasAgentToken(toolsLine) {
		t.Errorf(
			"DEPTH_SEAL_VIOLATION: manager-lead.md tools: field is missing the Agent "+
				"token (found: %q). manager-lead is the sole Agent-carrier; removing "+
				"it closes the fan-out seam. tools: MUST include Agent.",
			toolsLine,
		)
	}
}

// TestNoNestedLeafWorkerCarrier scans .claude/agents/**/*.md (including any
// future leaf-worker files under harness/ or other subdirs) for files that
// declare themselves manager-lead leaf-worker spawn targets, and asserts they
// do NOT carry Agent in their tools: list — the depth-2 seal.
func TestNoNestedLeafWorkerCarrier(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootDepth(t)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Skipf(".claude/agents/ not found at %s; skipping depth-seal audit", agentsDir)
	}

	violations := []string{}
	err := filepath.WalkDir(agentsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		// manager-lead.md itself is the depth-1 carrier; skip.
		if strings.HasSuffix(path, filepath.Join("moai", "manager-lead.md")) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !isLeafWorkerOf(data) {
			return nil
		}
		fm := extractFrontmatter(data)
		toolsLine := frontmatterFieldValue(fm, "tools")
		if toolsLine == "" {
			return nil // no tools: declared; nothing to check
		}
		if hasAgentToken(toolsLine) {
			rel, _ := filepath.Rel(projectRoot, path)
			violations = append(violations, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf(
			"DEPTH_SEAL_VIOLATION: the following files declare themselves manager-lead "+
				"leaf-worker spawn targets AND carry Agent in tools: (depth-2 nesting "+
				"forbidden — leaf workers MUST omit Agent): %v",
			violations,
		)
	}
}

// isLeafWorkerOf reports whether the file content declares itself a manager-lead
// leaf-worker spawn target. The declaration is opt-in via either:
//   - frontmatter field `leaf_of: manager-lead`
//   - body marker `<!-- manager-lead leaf-worker -->`
func isLeafWorkerOf(data []byte) bool {
	content := string(data)
	if strings.Contains(content, "leaf_of: manager-lead") {
		return true
	}
	if strings.Contains(content, "<!-- manager-lead leaf-worker -->") {
		return true
	}
	return false
}

// extractFrontmatter returns the YAML frontmatter block (the text between the
// first `---` and the second `---`). If no frontmatter is present, returns "".
func extractFrontmatter(data []byte) string {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return ""
	}
	return rest[:idx]
}

// frontmatterFieldValue extracts the value of a top-level `key:` line from the
// frontmatter block. Returns "" if not present.
func frontmatterFieldValue(fm, key string) string {
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimSpace(line)
		prefix := key + ":"
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(trimmed[len(prefix):])
		}
	}
	return ""
}

// hasAgentToken reports whether a CSV tools: value contains the literal token
// `Agent`. Tokens are split on commas; whitespace is trimmed. A substring
// match (e.g. "TaskManager" containing "Agent") is rejected — only a
// standalone Agent token counts.
func hasAgentToken(toolsCSV string) bool {
	for _, tok := range strings.Split(toolsCSV, ",") {
		if strings.TrimSpace(tok) == "Agent" {
			return true
		}
	}
	return false
}

// findProjectRootDepth walks up to find the directory containing go.mod.
func findProjectRootDepth(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

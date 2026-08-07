// mcp_template_neutrality_test.go: SPEC-TREND-MCP-001 M1 AC-TMC-002 / AC-TMC-003
// regression guard for the new template-managed `.mcp.json` surface.
//
// The distributed `.mcp.json` is the FIRST template file whose primary content
// is JSON entries referencing external package names. This test asserts the
// load-bearing invariants that the broader neutrality audits (C1/C2/C4/C5/C6 +
// SPEC-ID/commit-SHA/date leak) do not specifically frame for an MCP package
// list:
//
//   - the active mcpServers map carries EXACTLY 3 default-on entries
//     (context7, chrome-devtools, playwright) — ast-grep and moai stay opt-in
//     via `moai mcp add`, NEVER in the distributed default;
//   - no `$comment` JSONC form appears (standard JSON only);
//   - no entry carries a resolved secret (every env value is a ${VAR} literal);
//   - no SPEC-ID, commit SHA, macOS-bias absolute path, CLAUDE.local.md
//     reference, or `PR #N` reference leaks into the distributed file.
//
// Verified in ISOLATION via:
//
//	go test ./internal/template/... -run TestMCPNeutrality
package template_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// mcpAllowedActiveKeys is the exact set of active mcpServers keys permitted in
// the distributed default. ast-grep + moai stay opt-in (omitted from the active
// map, activated via `moai mcp add ...`).
var mcpAllowedActiveKeys = map[string]struct{}{
	"context7":        {},
	"chrome-devtools": {},
	"playwright":      {},
}

// mcpForbiddenTokenRes are the leak-class regexes the broader neutrality / leak
// audits own at the project level; this test re-checks them on the new file as
// a defense-in-depth regression guard.
var mcpForbiddenTokenRes = []*regexp.Regexp{
	regexp.MustCompile(`SPEC-[A-Z][A-Z0-9]+-[0-9]{3}`), // SPEC-ID leak
	regexp.MustCompile(`\b[0-9a-f]{40}\b`),             // 40-char commit SHA
	regexp.MustCompile(`\b[0-9a-f]{7,8}\b`),            // short SHA (word-bounded)
	regexp.MustCompile(`/Users/`),                      // macOS-bias absolute path
	regexp.MustCompile(`CLAUDE\.local\.md`),            // maintainer-only local file
	regexp.MustCompile(`PR #[0-9]+`),                   // pull-request number
	regexp.MustCompile(`\$comment`),                    // JSONC $comment form (REQ-TMC-004)
}

// TestMCPNeutralityTemplateShape asserts AC-TMC-001 (3-active-entries shape),
// AC-TMC-002 (no SPEC IDs / SHAs / macOS paths / PR refs), and AC-TMC-003 (SPEC-
// ID / SHA / date leak detector passes on the new file).
func TestMCPNeutralityTemplateShape(t *testing.T) {
	t.Parallel()
	root := findNeutralityRoot(t) // .../internal/template/templates
	path := filepath.Join(root, ".mcp.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template .mcp.json: %v", err)
	}

	// AC-TMC-002 / AC-TMC-003: forbidden-token scan over the raw file bytes.
	for _, re := range mcpForbiddenTokenRes {
		if re.Match(data) {
			t.Errorf("forbidden token %q in template .mcp.json:\n%s", re.String(), string(data))
		}
	}

	// AC-TMC-001: structural shape — exactly the 3 default-on active entries.
	var doc struct {
		Schema         string         `json:"$schema"`
		McpServers     map[string]any `json:"mcpServers"`
		StaggeredStart map[string]any `json:"staggeredStartup"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("template .mcp.json is not valid JSON: %v", err)
	}
	if len(doc.McpServers) != 3 {
		t.Errorf("template .mcp.json mcpServers count = %d, want exactly 3 (context7 + chrome-devtools + playwright)", len(doc.McpServers))
	}
	for k := range doc.McpServers {
		if _, ok := mcpAllowedActiveKeys[k]; !ok {
			t.Errorf("template .mcp.json carries non-default-on entry %q (only context7/chrome-devtools/playwright are permitted in the distributed default)", k)
		}
	}
	for _, required := range []string{"context7", "chrome-devtools", "playwright"} {
		if _, ok := doc.McpServers[required]; !ok {
			t.Errorf("template .mcp.json is missing required default-on entry %q", required)
		}
	}

	// Secret hygiene (REQ-TMC-002): no resolved secrets; any env value MUST be a
	// ${VAR} literal. Walk every entry's env block.
	for name, entry := range doc.McpServers {
		m, ok := entry.(map[string]any)
		if !ok {
			t.Errorf("entry %q is not a JSON object", name)
			continue
		}
		if env, ok := m["env"].(map[string]any); ok {
			for k, v := range env {
				s, ok := v.(string)
				if !ok {
					t.Errorf("entry %q env[%q] is not a string literal", name, k)
					continue
				}
				if !strings.HasPrefix(s, "${") || !strings.HasSuffix(s, "}") {
					t.Errorf("entry %q env[%q] = %q must be a ${VAR} literal (no resolved secrets)", name, k, s)
				}
			}
		}
	}
}

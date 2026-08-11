// mcp_doctrine_parity_test.go: CI guard for the MCP Configuration section of
// `.claude/rules/moai/core/settings-management.md` and its template mirror.
//
// The doctrine has been silently reverted twice by working-tree sweeps that
// restored a stale copy of the file: the shipped rule went back to claiming
// MoAI-ADK "no longer ships or provisions MCP servers", contradicting both the
// template-managed `.mcp.json` and the `moai mcp-server` entry the CLI
// actually writes. A user reading the reverted rule would conclude the MCP
// surface does not exist.
//
// SPEC-MCP-DEFAULT-ON-001 (2026-08-12) inverted the provisioning contract from
// opt-in default-off to default-on: the `moai` entry is now provisioned by
// `moai init` unless explicitly declined. This guard was updated in the same
// change to pin the new default-on contract instead of the old opt-in wording.
//
// The two copies are intentionally NOT byte-identical (the template is the
// neutralized variant), so this guard asserts the load-bearing claims rather
// than parity: the retired sentence is absent, and the provisioning behaviour
// plus the default-on local server are described, in BOTH trees.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mcpDoctrineRelPath = ".claude/rules/moai/core/settings-management.md"

// retiredMCPSentence is the pre-provisioning claim. Its reappearance is the
// exact regression signature this guard exists to catch.
const retiredMCPSentence = "no longer ships or provisions MCP servers"

func TestMCPConfigurationDoctrine(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	trees := map[string]string{
		"source":   filepath.Join(projectRoot, mcpDoctrineRelPath),
		"template": filepath.Join(projectRoot, "internal", "template", "templates", mcpDoctrineRelPath),
	}

	for name, path := range trees {
		name, path := name, path
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			body := string(data)

			if strings.Contains(body, retiredMCPSentence) {
				t.Errorf("MCP_DOCTRINE_REGRESSION: %s still carries the retired claim %q; "+
					"MoAI-ADK ships a template-managed .mcp.json and provisions the "+
					"`moai mcp-server` entry, so this sentence is false", path, retiredMCPSentence)
			}
			for _, want := range []string{
				"mcp-server",   // the local stdio server is named
				"moai mcp add", // the activation path for the catalogued entries
				"default-on",   // the SPEC-MCP-DEFAULT-ON-001 provisioning contract
			} {
				if !strings.Contains(body, want) {
					t.Errorf("MCP_DOCTRINE_REGRESSION: %s does not mention %q; the MCP Configuration "+
						"section must describe the provisioned surface and its default-on contract", path, want)
				}
			}
		})
	}
}

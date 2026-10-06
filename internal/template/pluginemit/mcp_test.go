// mcp_test.go — SPEC-PLUGIN-MARKETPLACE-001 M1 MCP derivation guard (AC-007 (c)).
package pluginemit_test

import (
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
)

// TestMCPEntryDerivedFromTemplate changes the moai entry inside a synthetic
// template tree and asserts the Codex manifest entry follows, that no other
// template server is carried, and that a tree without a usable moai entry is
// refused rather than papered over with a retyped literal.
func TestMCPEntryDerivedFromTemplate(t *testing.T) {
	cases := []struct {
		name    string
		mcp     string
		command string
		args    []string
	}{
		{
			name:    "default entry",
			mcp:     defaultMCPJSON,
			command: "moai",
			args:    []string{"mcp-server"},
		},
		{
			name:    "changed command and args",
			mcp:     `{"mcpServers": {"moai": {"command": "alt-launcher", "args": ["serve", "--flag"]}, "other": {"command": "x", "args": []}}}`,
			command: "alt-launcher",
			args:    []string{"serve", "--flag"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pub := emitSynthetic(t, tc.mcp)
			cp := decodeFile(t, pub, codexPluginPath)
			servers, _ := cp["mcpServers"].(map[string]any)
			if len(servers) != 1 {
				t.Fatalf("mcpServers = %v, want exactly the moai entry", servers)
			}
			entry, _ := servers["moai"].(map[string]any)
			if entry == nil {
				t.Fatalf("mcpServers has no moai key: %v", servers)
			}
			if entry["command"] != tc.command {
				t.Errorf("command = %v, want %q (copied from the template)", entry["command"], tc.command)
			}
			var gotArgs []string
			for _, a := range entry["args"].([]any) {
				gotArgs = append(gotArgs, a.(string))
			}
			if !reflect.DeepEqual(gotArgs, tc.args) {
				t.Errorf("args = %v, want %v (copied from the template)", gotArgs, tc.args)
			}
		})
	}

	t.Run("missing moai entry is refused", func(t *testing.T) {
		_, err := pluginemit.Emit(syntheticTemplate(`{"mcpServers": {"context7": {"command": "npx", "args": []}}}`), pluginemit.DefaultOptions())
		if err == nil {
			t.Fatal("Emit succeeded without a moai entry in the template .mcp.json")
		}
	})
	t.Run("malformed template is refused", func(t *testing.T) {
		_, err := pluginemit.Emit(syntheticTemplate(`{not json`), pluginemit.DefaultOptions())
		if err == nil {
			t.Fatal("Emit succeeded over a malformed template .mcp.json")
		}
	})
}

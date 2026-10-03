// mcp.go — the MCP entry derivation unit (REQ-007).
//
// The plugin's moai MCP server entry is copied from the template .mcp.json,
// never retyped, so a change to the template entry reaches the plugin on the
// next `make plugin-emit` and Claude's duplicate-command suppression keeps
// matching the project's own entry. M1 uses the unit for the Codex plugin
// manifest; M2 reuses it for the payload file plugins/moai/.mcp.json.
package pluginemit

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// MCPEntry is the moai MCP server entry: its command and args, as in the
// template. No other field is carried.
type MCPEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// DeriveMCPEntry reads the moai entry of the template .mcp.json at path. A
// missing file, malformed JSON, a missing moai entry or an empty command is an
// error: the entry is never defaulted.
func DeriveMCPEntry(fsys fs.FS, path string) (MCPEntry, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return MCPEntry{}, fmt.Errorf("pluginemit: read template MCP source %s: %w", path, err)
	}
	var doc struct {
		MCPServers map[string]MCPEntry `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return MCPEntry{}, fmt.Errorf("pluginemit: parse %s: %w", path, err)
	}
	entry, ok := doc.MCPServers[PluginName]
	if !ok || entry.Command == "" {
		return MCPEntry{}, fmt.Errorf("pluginemit: %s has no usable %q entry under mcpServers", path, PluginName)
	}
	if entry.Args == nil {
		entry.Args = []string{}
	}
	return entry, nil
}

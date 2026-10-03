package pluginemit

import "io/fs"

// MCPEntry is the moai MCP server entry copied from the template .mcp.json.
type MCPEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// DeriveMCPEntry reads the moai entry of the template .mcp.json at path.
func DeriveMCPEntry(fsys fs.FS, path string) (MCPEntry, error) {
	return MCPEntry{}, nil
}

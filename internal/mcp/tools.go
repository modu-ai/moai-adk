package mcp

// defineLSPTools returns the full set of LSP-backed MCP tool definitions.
func defineLSPTools() []Tool {
	return []Tool{
		{
			Name:        "goto_definition",
			Description: "Jump to the definition of a symbol at a given file and position",
			InputSchema: filePositionSchema(),
		},
		{
			Name:        "find_references",
			Description: "Find all references to a symbol at a given file and position",
			InputSchema: filePositionSchema(),
		},
		{
			Name:        "hover",
			Description: "Get type and documentation info for a symbol at a given position",
			InputSchema: filePositionSchema(),
		},
		{
			Name:        "document_symbols",
			Description: "List all symbols (functions, types, variables) in a file",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"file": {Type: "string", Description: "Relative file path"},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "diagnostics",
			Description: "Get LSP diagnostics (errors, warnings) for a file",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"file": {Type: "string", Description: "Relative file path"},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "rename",
			Description: "Rename a symbol across the project",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"file":     {Type: "string", Description: "Relative file path"},
					"line":     {Type: "integer", Description: "1-based line number"},
					"column":   {Type: "integer", Description: "1-based column number"},
					"new_name": {Type: "string", Description: "New name for the symbol"},
				},
				Required: []string{"file", "line", "column", "new_name"},
			},
		},
	}
}

// filePositionSchema returns the common schema used by tools that require a
// file path and a cursor position.
func filePositionSchema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"file":   {Type: "string", Description: "Relative file path"},
			"line":   {Type: "integer", Description: "1-based line number"},
			"column": {Type: "integer", Description: "1-based column number"},
		},
		Required: []string{"file", "line", "column"},
	}
}

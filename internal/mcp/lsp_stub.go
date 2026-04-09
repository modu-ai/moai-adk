package mcp

import (
	"context"
	"fmt"
)

// notConnectedMsg is the standard message returned by all stub LSP operations.
// Phase 2 will replace these stubs with real language server connections.
const notConnectedMsg = "LSP bridge not yet connected to a language server. " +
	"Install gopls (Go), typescript-language-server (TS), or pyright (Python) and restart."

func (h *LSPHandler) gotoDefinition(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, line, col := extractPosition(args)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("goto_definition: file=%s line=%d col=%d\n%s", file, line, col, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

func (h *LSPHandler) findReferences(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, line, col := extractPosition(args)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("find_references: file=%s line=%d col=%d\n%s", file, line, col, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

func (h *LSPHandler) hover(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, line, col := extractPosition(args)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("hover: file=%s line=%d col=%d\n%s", file, line, col, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

func (h *LSPHandler) documentSymbols(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, _ := args["file"].(string)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("document_symbols: file=%s\n%s", file, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

func (h *LSPHandler) diagnostics(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, _ := args["file"].(string)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("diagnostics: file=%s\n%s", file, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

func (h *LSPHandler) rename(_ context.Context, args map[string]any) (*ToolResult, error) {
	file, line, col := extractPosition(args)
	newName, _ := args["new_name"].(string)
	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("rename: file=%s line=%d col=%d new_name=%s\n%s", file, line, col, newName, notConnectedMsg),
		}},
		IsError: true,
	}, nil
}

// extractPosition pulls the file, line, and column arguments from a tool call
// arguments map, applying safe defaults when values are absent or malformed.
func extractPosition(args map[string]any) (file string, line, col int) {
	file, _ = args["file"].(string)
	line = intFromAny(args["line"])
	col = intFromAny(args["column"])
	return file, line, col
}

// intFromAny converts a value that may be float64 (JSON default) or int to int.
func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

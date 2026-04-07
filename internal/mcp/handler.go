package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// LSPHandler handles MCP tool calls and dispatches them to LSP-backed operations.
// In Phase 1 all LSP operations return a "not connected" stub response.
type LSPHandler struct {
	tools      []Tool
	projectDir string
}

// NewLSPHandler creates an MCP handler with the full set of LSP tools pre-registered.
func NewLSPHandler(projectDir string) *LSPHandler {
	return &LSPHandler{
		projectDir: projectDir,
		tools:      defineLSPTools(),
	}
}

// HandleMethod dispatches an MCP method call to the appropriate handler.
func (h *LSPHandler) HandleMethod(ctx context.Context, method string, params json.RawMessage) (any, error) {
	switch method {
	case "initialize":
		return h.handleInitialize()
	case "tools/list":
		return h.handleToolsList()
	case "tools/call":
		return h.handleToolsCall(ctx, params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (h *LSPHandler) handleInitialize() (*InitializeResult, error) {
	return &InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "moai-lsp",
			Version: "1.0.0",
		},
	}, nil
}

func (h *LSPHandler) handleToolsList() (map[string]any, error) {
	return map[string]any{"tools": h.tools}, nil
}

func (h *LSPHandler) handleToolsCall(ctx context.Context, params json.RawMessage) (*ToolResult, error) {
	var call ToolCallParams
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	switch call.Name {
	case "goto_definition":
		return h.gotoDefinition(ctx, call.Arguments)
	case "find_references":
		return h.findReferences(ctx, call.Arguments)
	case "hover":
		return h.hover(ctx, call.Arguments)
	case "document_symbols":
		return h.documentSymbols(ctx, call.Arguments)
	case "diagnostics":
		return h.diagnostics(ctx, call.Arguments)
	case "rename":
		return h.rename(ctx, call.Arguments)
	default:
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("unknown tool: %s", call.Name),
			}},
			IsError: true,
		}, nil
	}
}

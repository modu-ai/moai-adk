package cli

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// mcp_code_tools.go — the M5 code-query tools (REQ-GF-017..019): signature-
// level answers from the code-derived layer, every response carrying the
// tree+commit provenance. Read-only hint annotations match the audit-family
// precedent (#1619).
//
// CR round-2 (3855001953): all three handlers resolve their tree via
// resolveToolProjectRoot(req) — a supplied project_root is honored and a bad
// one REJECTED — so a worktree session gets its OWN tree's answer (the t246
// wrong-tree defect family), never a silent fallback to the primary checkout.
//
// CR round-2 (3855001978): results are shaped by the package's toolJSON /
// toolErr like every other moai tool — data rides in StructuredContent with
// a "<tool>: ok" text fallback; errors carry "<tool>: <message>" as an
// IsError text result. The file's former private marshal/error-shaping
// near-duplicates of those helpers are gone.

func handleGraphFileAPI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rel, err := req.RequireString("file")
	if err != nil {
		return toolErr("graph_file_api", err), nil
	}
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("graph_file_api", err), nil
	}
	res, err := graph.FileAPI(root, rel)
	if err != nil {
		return toolErr("graph_file_api", err), nil
	}
	return toolJSON("graph_file_api", res), nil
}

func handleGraphFindCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return toolErr("graph_find_code", err), nil
	}
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("graph_find_code", err), nil
	}
	matches, prov, err := graph.FindCode(root, query)
	if err != nil {
		return toolErr("graph_find_code", err), nil
	}
	return toolJSON("graph_find_code", map[string]interface{}{
		"matches":    matches,
		"provenance": prov,
	}), nil
}

func handleGraphTraceCalls(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol, err := req.RequireString("symbol")
	if err != nil {
		return toolErr("graph_trace_calls", err), nil
	}
	depth := req.GetInt("depth", 1)
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("graph_trace_calls", err), nil
	}
	callers, callees, err := graph.TraceCalls(root, symbol, depth)
	if err != nil {
		return toolErr("graph_trace_calls", err), nil
	}
	return toolJSON("graph_trace_calls", map[string]interface{}{
		"symbol":     symbol,
		"callers":    callers,
		"callees":    callees,
		"provenance": graph.AnswerProvenance(root),
	}), nil
}

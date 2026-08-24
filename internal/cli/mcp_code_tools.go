package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// mcp_code_tools.go — the M5 code-query tools (REQ-GF-017..019): signature-
// level answers from the code-derived layer, every response carrying the
// tree+commit provenance. Read-only hint annotations match the audit-family
// precedent (#1619).

// resolveCodeQueryRoot resolves the tree the code queries answer from. A
// worktree session MUST get its own tree's answer (the t246 wrong-tree
// defect family) — the project root resolver, never a shared cache path.
var resolveCodeQueryRoot = resolveProjectDir

func handleGraphFileAPI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rel, err := req.RequireString("file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	root := resolveCodeQueryRoot()
	res, err := graph.FileAPI(root, rel)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonToolResult(res)
}

func handleGraphFindCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	root := resolveCodeQueryRoot()
	matches, prov, err := graph.FindCode(root, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonToolResult(map[string]interface{}{
		"matches":   matches,
		"provenance": prov,
	})
}

func handleGraphTraceCalls(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol, err := req.RequireString("symbol")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	depth, _ := req.RequireInt("depth")
	root := resolveCodeQueryRoot()
	callers, callees, err := graph.TraceCalls(root, symbol, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonToolResult(map[string]interface{}{
		"symbol":    symbol,
		"callers":   callers,
		"callees":   callees,
		"provenance": graph.AnswerProvenance(root),
	})
}

// jsonToolResult marshals v with a stable error wrap.
func jsonToolResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("mcp: marshal code-query result: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

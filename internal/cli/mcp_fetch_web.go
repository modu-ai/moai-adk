// mcp_fetch_web.go — SPEC-FETCH-WEB-001 M4 (REQ-FW-009). The fetch_web MCP
// tool: a thin wrapper that reads the workflow.jev.enabled gate and, once
// past it, calls the ONE shared pipeline entry point (RunFetchPipeline) the
// CLI command calls — no second scoring copy, no transport code here. The
// tool's result shape is the CLI --json shape, so both surfaces observe
// identical status, block, and exit-code semantics.
package cli

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
)

// fetchWebToolName is the registered MCP tool name (operator-fixed naming,
// spec.md §D5). It MUST match the catalogue entry in internal/mcp and the
// catalogue documentation in .claude/rules/moai/core/moai-mcp-tools*.md.
const fetchWebToolName = "fetch_web"

// registerFetchWebTool registers the tool. Split out of
// registerMoaiMCPTools to keep the registration table scannable; the closure
// argument is the same `add` helper the table uses, so the per-tool
// enablement map and the catalogue-equality guard see this tool identically.
func registerFetchWebTool(add func(name string, tool mcp.Tool, handler server.ToolHandlerFunc)) {
	add(fetchWebToolName, mcp.NewTool(
		fetchWebToolName,
		mcp.WithDescription("Fetch a web page and return VERBATIM excerpts (paragraph/list-item/table-row/code blocks with heading paths) that answer a question, scored in two Jev stages. Same pipeline and output contract as `moai fetch web`: statuses filtered/passthrough/no_match/empty/fetch_error with the exit-status mapping carried as exit_code. Gated behind workflow.jev.enabled (shipped default false): with the gate off this tool makes zero network requests — no page fetch, no Jev call — and reports fetch_error (exit_code 2) pointing at WebFetch."),
		mcp.WithString("url", mcp.Required(), mcp.Description("The page URL to fetch.")),
		mcp.WithString("question", mcp.Required(), mcp.Description("The question the returned excerpts must answer.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleFetchWeb)
}

// @MX:WARN: [AUTO] gate-read path — handleFetchWeb reads workflow.jev.enabled on every call and must keep it first and short-circuiting
// @MX:REASON: [AUTO] a reordered or degraded gate read would let a network-fetching tool run from a project that never opted in while the zero-network gate-off contract (REQ-FW-008) depends on this ordering; the seams are consulted only after the gate passes.
func handleFetchWeb(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, rootErr := resolveToolProjectRootWithSource(req)
	if rootErr != nil {
		return toolErr(fetchWebToolName, rootErr), nil
	}
	// The gate is read FIRST and short-circuits: while the capability is
	// disabled no page client and no asker are constructed, so no request
	// can exist (the CLI command shares this ordering under the same
	// constraint family).
	enabled, readErr := jevEnabled(root)
	if readErr != nil || !enabled {
		reason := "workflow.jev.enabled is false (shipped default); no page fetch and no Jev call — use WebFetch (or enable Jev)"
		if readErr != nil {
			reason = "workflow.jev.enabled gate unreadable; treated as disabled; no page fetch and no Jev call — use WebFetch"
		}
		return toolJSONWithRoot(fetchWebToolName, map[string]any{
			"status":    string(fetchweb.StatusFetchError),
			"exit_code": fetchweb.StatusFetchError.ExitCode(),
			"condition": reason,
		}, root, source), nil
	}

	url, err := req.RequireString("url")
	if err != nil {
		return toolErr(fetchWebToolName, errors.New("url is required: the page URL to fetch")), nil
	}
	question, err := req.RequireString("question")
	if err != nil {
		return toolErr(fetchWebToolName, errors.New("question is required: what the excerpts must answer")), nil
	}

	// The ONE shared pipeline entry point (AC-FW-006): the same function the
	// CLI command calls, through the same seams its tests inject.
	res := RunFetchPipeline(ctx, url, question, fetchWebPageClientFn(), fetchWebAskerFn())

	out := map[string]any{
		"status":       string(res.Status),
		"exit_code":    res.Status.ExitCode(),
		"url":          url,
		"question":     question,
		"total_tokens": res.TotalTokens, // approximate (fetchweb package note)
		"kept_tokens":  res.KeptTokens,  // approximate (fetchweb package note)
	}
	if res.Condition != "" {
		out["condition"] = res.Condition
	}
	if len(res.Blocks) > 0 {
		blocks := make([]map[string]any, len(res.Blocks))
		for i, b := range res.Blocks {
			blocks[i] = map[string]any{
				"kind":    string(b.Kind),
				"heading": b.HeadingPath,
				"text":    b.Text,
				"aside":   b.Aside,
			}
		}
		out["blocks"] = blocks
	}
	return toolJSONWithRoot(fetchWebToolName, out, root, source), nil
}

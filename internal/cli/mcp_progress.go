// Package cli — MCP progress notification helper.
//
// mcp_progress.go provides notifyMCPProgress, the helper long-running tool
// handlers call to surface human-readable progress to the MCP client (Claude
// Code) DURING a tool call — before the final CallToolResult is returned.
//
// Without this, a long audit (codex_audit, glm_audit, codex_task, audit_multi)
// sends NOTHING to the client for its entire duration, so the screen appears
// frozen and Claude Code's idle watchdog (stdio default 30m) may abort the
// call. The progress notification both surfaces the step AND resets the idle
// watchdog.
//
// Two notification channels are emitted (best-effort, fail-open):
//   - notifications/message (logging) — token-independent; surfaces the step
//     as a log line whenever the client renders server logging.
//   - notifications/progress — emitted only when the client supplied a
//     progressToken on the request (the standard MCP progress opt-in); this
//     is the channel Claude Code's idle watchdog listens to.
//
// A send error is advisory-only and MUST NOT fail the tool (fail-open).
//
// @MX:NOTE: progress UX for long-running MCP tools — addresses the "no
// on-screen progress while an MCP tool runs" gap; Claude Code aborts a tool
// that sends no response or progress for the idle window (stdio default 30m)
// per the MCP doc, and these notifications reset that watchdog.
package cli

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCP notification methods consumed by notifyMCPProgress.
const (
	mcpProgressMethod = "notifications/progress"
	mcpLoggingMethod  = "notifications/message"
)

// extractProgressToken pulls the client-supplied progressToken from a tool
// request's _meta, if present. Returns nil when the client did not opt into
// progress reporting (no _meta, or no progressToken field).
func extractProgressToken(req mcp.CallToolRequest) mcp.ProgressToken {
	if req.Params.Meta == nil {
		return nil
	}
	return req.Params.Meta.ProgressToken
}

// notifyMCPProgress surfaces a progress step to the MCP client during a
// long-running tool call. It is a no-op (fail-open) when no MCPServer is
// reachable from ctx, and emits notifications best-effort — a send error never
// propagates to the tool result.
//
// token is the progressToken extracted from the request (nil when the client
// did not opt in); progress is an advisory 0..1-ish fraction; message is the
// human-readable step (e.g. "codex 응답 대기 중...").
func notifyMCPProgress(ctx context.Context, token mcp.ProgressToken, progress float64, message string) {
	s := server.ServerFromContext(ctx)
	if s == nil {
		return // not inside an MCP server context (e.g. direct unit test)
	}
	// Logging notification — token-independent; surfaces the step as a log
	// line whenever the client renders server logging.
	_ = s.SendNotificationToClient(ctx, mcpLoggingMethod, map[string]any{
		"level": "info",
		"data": map[string]any{
			"progress": progress,
			"message":  message,
		},
	})
	// Progress notification — only when the client opted in via progressToken.
	// This is the channel Claude Code's idle watchdog listens to.
	if token == nil {
		return
	}
	_ = s.SendNotificationToClient(ctx, mcpProgressMethod, map[string]any{
		"progressToken": token,
		"progress":      progress,
		"message":       message,
	})
}

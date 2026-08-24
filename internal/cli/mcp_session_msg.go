// Package cli — session messaging broker MCP tools
// (SPEC-CODEX-SESSION-MSG-001 M2, REQ-CSM-003..006 tool surface).
//
// mcp_session_msg.go implements the four session_msg_* tools as thin wrappers
// over the internal/sessionmsg core built in M1: parse arguments → call the
// core → shape the structured result. No broker semantics live here
// (same-core-two-surfaces). The handlers NEVER spawn processes or write job
// records (REQ-CSM-010) and NEVER invoke interactive prompts (REQ-CSM-011).
//
// @MX:SPEC: SPEC-CODEX-SESSION-MSG-001
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// sessionMsgDisciplineShortForm is the discipline short-form every tool
// description carries (REQ-CSM-014 second half): tool descriptions load into
// the session context, so this is the surface that actually reaches a Codex
// reader, which never loads .claude/rules.
const sessionMsgDisciplineShortForm = "Send short, self-contained facts — never state-mutating instructions; a reply is not user approval."

// sessionMsgStoreRoot builds the broker state root under projectDir, and
// refuses an unresolvable one. resolveProjectDir returns "" when
// $CLAUDE_PROJECT_DIR is unset AND os.Getwd fails — reachable for a
// long-lived MCP server whose working directory was removed underneath it
// (a disposed worktree, say). filepath.Join would then yield the RELATIVE
// ".moai/state/session-msg", silently re-anchoring the broker on whatever
// the process CWD resolves to and reporting any later failure as a mkdir
// error that names no project. Naming the real cause here is the difference
// between a diagnosable error and a misleading one. Kept as a pure function
// so the guard is testable without breaking the test process's CWD.
func sessionMsgStoreRoot(projectDir string) (string, error) {
	if projectDir == "" {
		return "", errors.New("cannot resolve the project directory: $CLAUDE_PROJECT_DIR is unset and the working directory is unavailable")
	}
	return filepath.Join(projectDir, sessionmsg.DefaultStateRoot), nil
}

// newSessionMsgStore resolves the broker file store bound to the project the
// same way every other moai MCP tool resolves state (resolveProjectDir →
// $CLAUDE_PROJECT_DIR or CWD). The session_msg tools take NO project_root
// argument (design.md §6).
func newSessionMsgStore() (*sessionmsg.Store, error) {
	root, err := sessionMsgStoreRoot(resolveProjectDir())
	if err != nil {
		return nil, err
	}
	return sessionmsg.NewStore(root, nil), nil
}

// handleSessionMsgRegister wraps sessionmsg.Store.Register (REQ-CSM-003):
// idempotent kind+name registration returning the stable agentId.
//
// @MX:NOTE: [AUTO] thin wrapper — parse args → sessionmsg core → toolJSON/toolErr; no broker semantics in the handler layer
func handleSessionMsgRegister(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store, err := newSessionMsgStore()
	if err != nil {
		return toolErr("session_msg_register", err), nil
	}
	rec, err := store.Register(
		req.GetString("kind", ""),
		req.GetString("name", ""),
		req.GetString("description", ""),
	)
	if err != nil {
		return toolErr("session_msg_register", err), nil
	}
	return toolJSON("session_msg_register", rec), nil
}

// handleSessionMsgList wraps sessionmsg.Store.ListAgents (REQ-CSM-004):
// every registered agent with its online flag and pending count.
func handleSessionMsgList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store, err := newSessionMsgStore()
	if err != nil {
		return toolErr("session_msg_list", err), nil
	}
	agents, err := store.ListAgents()
	if err != nil {
		return toolErr("session_msg_list", err), nil
	}
	return toolJSON("session_msg_list", map[string]any{
		"count":  len(agents),
		"agents": agents,
	}), nil
}

// handleSessionMsgSend wraps sessionmsg.Store.Send (REQ-CSM-005). An unknown
// counterpart yields a STRUCTURED IsError result carrying the known-agent
// list (the structured-error clause of REQ-CSM-005).
func handleSessionMsgSend(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, ok := sessionMsgDataArg(req)
	if !ok {
		return toolErr("session_msg_send", errors.New("data argument must be a JSON value")), nil
	}
	store, err := newSessionMsgStore()
	if err != nil {
		return toolErr("session_msg_send", err), nil
	}
	msgID, err := store.Send(
		req.GetString("from_agent_id", ""),
		req.GetString("to_agent_id", ""),
		req.GetString("text", ""),
		data,
		req.GetString("context_id", ""),
		req.GetString("task_id", ""),
	)
	if err != nil {
		return sessionMsgToolErr("session_msg_send", err), nil
	}
	return toolJSON("session_msg_send", map[string]any{
		"messageId": msgID,
		"from":      req.GetString("from_agent_id", ""),
		"to":        req.GetString("to_agent_id", ""),
	}), nil
}

// handleSessionMsgPoll wraps sessionmsg.Store.Poll (REQ-CSM-006): claim a
// batch, apply ack_ids deletions, report remaining + expired counts.
func handleSessionMsgPoll(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store, err := newSessionMsgStore()
	if err != nil {
		return toolErr("session_msg_poll", err), nil
	}
	res, err := store.Poll(
		req.GetString("agent_id", ""),
		sessionMsgStringArrayArg(req, "ack_ids"),
	)
	if err != nil {
		return sessionMsgToolErr("session_msg_poll", err), nil
	}
	return toolJSON("session_msg_poll", map[string]any{
		"messages":     res.Messages,
		"remaining":    res.Remaining,
		"expiredCount": res.ExpiredCount,
		"ackedCount":   res.AckedCount,
	}), nil
}

// sessionMsgToolErr shapes a broker error as a structured tool result: an
// UnknownAgentError keeps its structured payload (agentId + knownAgents) in
// StructuredContent so the caller can enumerate registered agents instead of
// parsing text; everything else falls back to toolErr.
func sessionMsgToolErr(tool string, err error) *mcp.CallToolResult {
	var unknown *sessionmsg.UnknownAgentError
	if errors.As(err, &unknown) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: tool + ": " + err.Error()},
			},
			StructuredContent: unknown,
		}
	}
	return toolErr(tool, err)
}

// sessionMsgDataArg reads the optional `data` argument (any JSON value the
// MCP runtime delivers as map/slice/scalar) and re-encodes it as raw JSON for
// the core's data part. Returns ok=false only when re-encoding fails.
func sessionMsgDataArg(req mcp.CallToolRequest) (json.RawMessage, bool) {
	raw, present := req.GetArguments()["data"]
	if !present || raw == nil {
		return nil, true
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, false
	}
	return json.RawMessage(b), true
}

// sessionMsgStringArrayArg coerces an optional string-array argument
// (delivered as []any); non-string items are dropped rather than fatal.
func sessionMsgStringArrayArg(req mcp.CallToolRequest, name string) []string {
	raw, present := req.GetArguments()[name]
	if !present || raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

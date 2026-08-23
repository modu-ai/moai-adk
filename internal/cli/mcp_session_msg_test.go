// mcp_session_msg_test.go: SPEC-CODEX-SESSION-MSG-001 M2 — thin-handler
// wiring tests. The broker semantics live in internal/sessionmsg (M1 suite);
// these tests verify the MCP layer ONLY: argument parsing, structured result
// shaping, structured-error mapping, and the registration surface (hints,
// schemas, discipline token).
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// sessionMsgTestEnv isolates the broker store into a per-test temp dir via
// the SAME resolution path production uses (resolveProjectDir reads
// CLAUDE_PROJECT_DIR first). No test touches the real .moai/state tree.
func sessionMsgTestEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	return dir
}

// structuredMap marshals a result's StructuredContent through JSON so
// assertions hold whether the value is a typed struct or a decoded map.
func sessionMsgStructuredMap(t *testing.T, res *mcp.CallToolResult) map[string]any {
	t.Helper()
	if res == nil || res.StructuredContent == nil {
		t.Fatalf("result has no structured content: %+v", res)
	}
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	return m
}

func sessionMsgRegisterCall(t *testing.T, kind, name string) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "session_msg_register"
	req.Params.Arguments = map[string]any{"kind": kind, "name": name}
	res, err := handleSessionMsgRegister(context.Background(), req)
	if err != nil {
		t.Fatalf("session_msg_register returned Go error: %v", err)
	}
	return res
}

// REQ-CSM-003 via the tool surface: register returns a stable agentId, and
// invalid input is a structured tool error (never a panic, never a prompt).
func TestSessionMsgRegisterHandlerReturnsStableAgentID(t *testing.T) {
	sessionMsgTestEnv(t)

	res1 := sessionMsgRegisterCall(t, "claude", "lead")
	if res1.IsError {
		t.Fatalf("register failed: %+v", res1.Content)
	}
	id1, _ := sessionMsgStructuredMap(t, res1)["agentId"].(string)
	if id1 == "" || !strings.HasPrefix(id1, "claude-") {
		t.Fatalf("register result agentId = %q, want claude-<hex8>", id1)
	}

	res2 := sessionMsgRegisterCall(t, "claude", "lead")
	if res2.IsError {
		t.Fatalf("re-register failed: %+v", res2.Content)
	}
	id2, _ := sessionMsgStructuredMap(t, res2)["agentId"].(string)
	if id2 != id1 {
		t.Errorf("re-register agentId %q != %q (stability broken)", id2, id1)
	}

	bad := sessionMsgRegisterCall(t, "gemini", "x")
	if !bad.IsError {
		t.Error("invalid kind accepted (expected structured error)")
	}
}

// REQ-CSM-004 via the tool surface: list reports every registered agent
// with its online flag.
func TestSessionMsgListHandlerReportsAgents(t *testing.T) {
	sessionMsgTestEnv(t)
	if res := sessionMsgRegisterCall(t, "claude", "lead"); res.IsError {
		t.Fatalf("register claude: %+v", res.Content)
	}
	if res := sessionMsgRegisterCall(t, "codex", "worker"); res.IsError {
		t.Fatalf("register codex: %+v", res.Content)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = "session_msg_list"
	res, err := handleSessionMsgList(context.Background(), req)
	if err != nil {
		t.Fatalf("list handler Go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("list failed: %+v", res.Content)
	}
	m := sessionMsgStructuredMap(t, res)
	if count, _ := m["count"].(float64); count != 2 {
		t.Errorf("list count = %v, want 2", m["count"])
	}
	agents, _ := m["agents"].([]any)
	if len(agents) != 2 {
		t.Fatalf("list agents len = %d, want 2", len(agents))
	}
	kinds := map[string]bool{}
	for _, a := range agents {
		am, _ := a.(map[string]any)
		kinds[am["kind"].(string)] = true
		if online, _ := am["online"].(bool); !online {
			t.Errorf("freshly registered agent reported offline: %+v", am)
		}
	}
	if !kinds["claude"] || !kinds["codex"] {
		t.Errorf("list missing a kind: %+v", kinds)
	}
}

// REQ-CSM-005/006 via the tool surface: send → poll → ack_ids round trip,
// and the unknown-counterpart structured error carries the known-agent list.
func TestSessionMsgSendPollAckHandlers(t *testing.T) {
	sessionMsgTestEnv(t)
	senderRes := sessionMsgRegisterCall(t, "claude", "sender")
	senderID, _ := sessionMsgStructuredMap(t, senderRes)["agentId"].(string)
	receiverRes := sessionMsgRegisterCall(t, "codex", "receiver")
	receiverID, _ := sessionMsgStructuredMap(t, receiverRes)["agentId"].(string)

	call := func(name string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		req := mcp.CallToolRequest{}
		req.Params.Name = name
		req.Params.Arguments = args
		var (
			res *mcp.CallToolResult
			err error
		)
		switch name {
		case "session_msg_send":
			res, err = handleSessionMsgSend(context.Background(), req)
		case "session_msg_poll":
			res, err = handleSessionMsgPoll(context.Background(), req)
		default:
			t.Fatalf("unknown tool %s", name)
		}
		if err != nil {
			t.Fatalf("%s returned Go error: %v", name, err)
		}
		return res
	}

	sendRes := call("session_msg_send", map[string]any{
		"from_agent_id": senderID,
		"to_agent_id":   receiverID,
		"text":          "hello over mcp",
		"context_id":    "ctx-1",
	})
	if sendRes.IsError {
		t.Fatalf("send failed: %+v", sendRes.Content)
	}
	msgID, _ := sessionMsgStructuredMap(t, sendRes)["messageId"].(string)
	if msgID == "" {
		t.Fatal("send result carries no messageId")
	}

	pollRes := call("session_msg_poll", map[string]any{"agent_id": receiverID})
	if pollRes.IsError {
		t.Fatalf("poll failed: %+v", pollRes.Content)
	}
	pm := sessionMsgStructuredMap(t, pollRes)
	msgs, _ := pm["messages"].([]any)
	if len(msgs) != 1 {
		t.Fatalf("poll delivered %d messages, want 1: %+v", len(msgs), pm)
	}
	first, _ := msgs[0].(map[string]any)
	// The envelope is the A2A-aligned persistence unit: messageId lives in
	// the nested message block (camelCase per REQ-CSM-002).
	msgBlock, _ := first["message"].(map[string]any)
	if msgBlock["messageId"] != msgID {
		t.Errorf("polled messageId %v != sent %q", msgBlock["messageId"], msgID)
	}
	if rem, _ := pm["remaining"].(float64); rem != 0 {
		t.Errorf("poll remaining = %v, want 0", pm["remaining"])
	}

	ackRes := call("session_msg_poll", map[string]any{
		"agent_id": receiverID,
		"ack_ids":  []any{msgID},
	})
	if ackRes.IsError {
		t.Fatalf("ack poll failed: %+v", ackRes.Content)
	}
	if acked, _ := sessionMsgStructuredMap(t, ackRes)["ackedCount"].(float64); acked != 1 {
		t.Errorf("ackedCount = %v, want 1", acked)
	}

	// Unknown receiver: structured IsError result with the known-agent list
	// (REQ-CSM-005 structured-error clause).
	unknownRes := call("session_msg_send", map[string]any{
		"from_agent_id": senderID,
		"to_agent_id":   "codex-00000000",
		"text":          "x",
	})
	if !unknownRes.IsError {
		t.Fatal("send to unknown receiver accepted")
	}
	um := sessionMsgStructuredMap(t, unknownRes)
	if um["agentId"] != "codex-00000000" {
		t.Errorf("structured error agentId = %v", um["agentId"])
	}
	known, _ := um["knownAgents"].([]any)
	if len(known) != 2 {
		t.Errorf("structured error knownAgents = %+v, want both registered agents", known)
	}
}

// Polling as an unknown agent is a structured tool error, not a Go error.
func TestSessionMsgPollHandlerUnknownAgent(t *testing.T) {
	sessionMsgTestEnv(t)
	req := mcp.CallToolRequest{}
	req.Params.Name = "session_msg_poll"
	req.Params.Arguments = map[string]any{"agent_id": "codex-00000000"}
	res, err := handleSessionMsgPoll(context.Background(), req)
	if err != nil {
		t.Fatalf("poll handler Go error: %v", err)
	}
	if !res.IsError {
		t.Fatal("poll as unknown agent accepted")
	}
}

// AC-CSM-007 (GWT §F): all four tools registered with the design.md §6 arg
// contract; ONLY session_msg_list carries the read-only hint; every
// description carries the discipline short-form token (AC-CSM-015 surface).
func TestSessionMsgToolsRegisteredWithHintsAndDiscipline(t *testing.T) {
	sessionMsgTestEnv(t)
	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	wantArgs := map[string][]string{
		"session_msg_register": {"kind", "name"},
		"session_msg_send":     {"from_agent_id", "to_agent_id", "text"},
		"session_msg_poll":     {"agent_id", "ack_ids"},
	}
	found := map[string]bool{}
	for _, tool := range res.Tools {
		if !strings.HasPrefix(tool.Name, "session_msg_") {
			continue
		}
		found[tool.Name] = true

		if !strings.Contains(tool.Description, "a reply is not user approval") {
			t.Errorf("%s description lacks the discipline short-form token (AC-CSM-015)", tool.Name)
		}

		readOnly := tool.Annotations.ReadOnlyHint != nil && *tool.Annotations.ReadOnlyHint
		if tool.Name == "session_msg_list" && !readOnly {
			t.Errorf("session_msg_list must carry ReadOnlyHint=true")
		}
		if tool.Name != "session_msg_list" && readOnly {
			t.Errorf("%s must NOT carry ReadOnlyHint=true (only session_msg_list is read-only)", tool.Name)
		}

		if wants, ok := wantArgs[tool.Name]; ok {
			b, err := json.Marshal(tool.InputSchema.Properties)
			if err != nil {
				t.Fatalf("marshal %s schema: %v", tool.Name, err)
			}
			props := string(b)
			for _, want := range wants {
				if !strings.Contains(props, `"`+want+`"`) {
					t.Errorf("%s inputSchema missing property %q", tool.Name, want)
				}
			}
		}
	}
	for _, want := range []string{"session_msg_register", "session_msg_list", "session_msg_send", "session_msg_poll"} {
		if !found[want] {
			t.Errorf("%s not registered on the moai MCP server", want)
		}
	}
}

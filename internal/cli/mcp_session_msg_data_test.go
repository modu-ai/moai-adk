// mcp_session_msg_data_test.go: handler-layer coverage for the optional
// `data` argument (SPEC-CODEX-SESSION-MSG-001 M2). sessionMsgDataArg is the
// one argument-parsing branch in the handler layer with no other test, and it
// is the branch a structured payload has to survive intact.
package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// An unresolvable project directory is refused by name rather than silently
// re-anchoring the broker on a relative path.
//
// Mutant this test catches that a shallower one would not: an implementation
// that returns filepath.Join("", DefaultStateRoot) with a nil error still
// "works" for every caller whose CWD happens to be the project root, so a
// test asserting only the happy path passes. The empty-input case is
// asserted to be an ERROR, and the happy path is asserted to still join, so
// a guard that refuses everything fails too.
func TestSessionMsgStoreRootRefusesUnresolvableProjectDir(t *testing.T) {
	if _, err := sessionMsgStoreRoot(""); err == nil {
		t.Fatal("empty project dir accepted — broker would anchor on a relative path")
	} else if !strings.Contains(err.Error(), "project directory") {
		t.Errorf("error does not name the cause: %v", err)
	}

	root, err := sessionMsgStoreRoot(filepath.FromSlash("/tmp/project"))
	if err != nil {
		t.Fatalf("resolvable project dir rejected: %v", err)
	}
	if want := filepath.Join(filepath.FromSlash("/tmp/project"), sessionmsg.DefaultStateRoot); root != want {
		t.Errorf("store root = %q, want %q", root, want)
	}
}

// A structured `data` argument survives the handler round trip: it is
// re-encoded into the envelope's data part on send and comes back out of poll
// with its shape and values intact.
//
// Mutant this test catches that a shallower one would not: an implementation
// that drops the data part entirely still returns a messageId and still
// delivers one message, so a test asserting only "send succeeded" or
// "poll returned 1 message" passes. The assertion walks into the polled
// envelope's parts and compares the payload's fields, so a silently discarded
// or flattened (e.g. fmt.Sprint'ed) payload fails. A nested object and a
// non-string scalar are both included, because a stringifying implementation
// survives a flat all-strings payload.
func TestSessionMsgSendHandlerCarriesDataArgument(t *testing.T) {
	sessionMsgTestEnv(t)
	senderID, _ := sessionMsgStructuredMap(t, sessionMsgRegisterCall(t, "claude", "sender"))["agentId"].(string)
	receiverID, _ := sessionMsgStructuredMap(t, sessionMsgRegisterCall(t, "codex", "receiver"))["agentId"].(string)
	if senderID == "" || receiverID == "" {
		t.Fatal("registration produced no agentId")
	}

	sendReq := mcp.CallToolRequest{}
	sendReq.Params.Name = "session_msg_send"
	sendReq.Params.Arguments = map[string]any{
		"from_agent_id": senderID,
		"to_agent_id":   receiverID,
		"text":          "payload attached",
		"data": map[string]any{
			"kind":   "verdict",
			"passed": true,
			"scores": map[string]any{"craft": float64(4)},
		},
	}
	sendRes, err := handleSessionMsgSend(context.Background(), sendReq)
	if err != nil {
		t.Fatalf("send handler Go error: %v", err)
	}
	if sendRes.IsError {
		t.Fatalf("send with data failed: %+v", sendRes.Content)
	}

	pollReq := mcp.CallToolRequest{}
	pollReq.Params.Name = "session_msg_poll"
	pollReq.Params.Arguments = map[string]any{"agent_id": receiverID}
	pollRes, err := handleSessionMsgPoll(context.Background(), pollReq)
	if err != nil {
		t.Fatalf("poll handler Go error: %v", err)
	}
	if pollRes.IsError {
		t.Fatalf("poll failed: %+v", pollRes.Content)
	}

	pm := sessionMsgStructuredMap(t, pollRes)
	msgs, ok := pm["messages"].([]any)
	if !ok || len(msgs) != 1 {
		t.Fatalf("poll delivered %v messages, want 1: %+v", len(msgs), pm)
	}
	env, ok := msgs[0].(map[string]any)
	if !ok {
		t.Fatalf("envelope is not an object: %T", msgs[0])
	}
	msgBlock, ok := env["message"].(map[string]any)
	if !ok {
		t.Fatalf("message block is not an object: %T", env["message"])
	}
	parts, ok := msgBlock["parts"].([]any)
	if !ok {
		t.Fatalf("parts is not an array: %T", msgBlock["parts"])
	}

	var data map[string]any
	for _, p := range parts {
		part, ok := p.(map[string]any)
		if !ok {
			t.Fatalf("part is not an object: %T", p)
		}
		if kind, _ := part["kind"].(string); kind != "data" {
			continue
		}
		data, ok = part["data"].(map[string]any)
		if !ok {
			t.Fatalf("data part payload is not an object: %T", part["data"])
		}
	}
	if data == nil {
		t.Fatalf("no data part in the polled envelope: %+v", parts)
	}

	if kind, _ := data["kind"].(string); kind != "verdict" {
		t.Errorf("data.kind = %v, want %q", data["kind"], "verdict")
	}
	if passed, ok := data["passed"].(bool); !ok || !passed {
		t.Errorf("data.passed = %#v, want the boolean true (not a stringified copy)", data["passed"])
	}
	scores, ok := data["scores"].(map[string]any)
	if !ok {
		t.Fatalf("data.scores is not a nested object: %T", data["scores"])
	}
	if craft, _ := scores["craft"].(float64); craft != 4 {
		t.Errorf("data.scores.craft = %v, want 4", scores["craft"])
	}
}

// A `data` argument that cannot be re-encoded as JSON is a structured tool
// error, not a Go error and not a silent drop.
func TestSessionMsgSendHandlerRejectsUnencodableData(t *testing.T) {
	sessionMsgTestEnv(t)
	senderID, _ := sessionMsgStructuredMap(t, sessionMsgRegisterCall(t, "claude", "sender"))["agentId"].(string)
	receiverID, _ := sessionMsgStructuredMap(t, sessionMsgRegisterCall(t, "codex", "receiver"))["agentId"].(string)

	req := mcp.CallToolRequest{}
	req.Params.Name = "session_msg_send"
	req.Params.Arguments = map[string]any{
		"from_agent_id": senderID,
		"to_agent_id":   receiverID,
		"text":          "x",
		"data":          make(chan int), // json.Marshal cannot encode a channel
	}
	res, err := handleSessionMsgSend(context.Background(), req)
	if err != nil {
		t.Fatalf("send handler Go error: %v", err)
	}
	if !res.IsError {
		t.Fatal("unencodable data argument accepted")
	}
}

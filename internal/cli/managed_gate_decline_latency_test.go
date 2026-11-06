// managed_gate_decline_latency_test.go — SPEC-MCP-NEUTRALITY-001 M3 (card
// t1632, debt N5 / AC-MN-012): the "without stalling" clause is QUANTIFIED —
// an elicitation request arriving on the gate client path receives its
// automatic decline within a bounded wall-clock budget, and the reply echoes
// the request id verbatim (no stdin interleaving: one reply, on the session's
// own write channel).
package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestManagedGateElicitationDeclineBounded pins the N5 quantification: the
// automatic decline returns within 50 ms of the request (a stall would hang
// the session past any bound) and the reply is the schema-valid
// {"action":"decline"} result with the request's own id, verbatim.
func TestManagedGateElicitationDeclineBounded(t *testing.T) {
	client, server := newHardenPair(t)

	rawParams, err := json.Marshal(map[string]any{
		"threadId": "th-n5", "turnId": "T1", "itemId": "item-n5",
		"serverName": "context7", "message": "need input",
	})
	if err != nil {
		t.Fatal(err)
	}
	reqID := json.RawMessage(`501`)
	req := managedCodexAppReply{
		ID:     reqID,
		Method: "mcpServer/elicitation/request",
		Params: rawParams,
	}

	start := time.Now()
	if err := client.answerServerRequest(req); err != nil {
		t.Fatalf("answerServerRequest: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("the automatic decline took %s — the without-stalling bound is 50ms", elapsed)
	}

	select {
	case frame := <-server.frames:
		if !frame.isReply() {
			t.Fatalf("the client wrote a %q request, want the elicitation reply", frame.Method)
		}
		if string(frame.ID) != string(reqID) {
			t.Errorf("reply id = %s, want the request's own %s (verbatim echo)", frame.ID, reqID)
		}
		if !strings.Contains(string(frame.Result), `"action":"decline"`) {
			t.Fatalf("reply result %s, want the decline action", frame.Result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no reply was written — the elicitation request stalled the session")
	}
}

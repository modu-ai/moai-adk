package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// SPEC codex-gate-protocol — request-shape pin for the corrected codex
// app-server JSON-RPC session. These tests assert the FOUR protocol gaps are
// closed in the request the client actually sends, independent of the live
// binary:
//
//  1. initialize is sent FIRST with a clientInfo {name,version} object.
//  2. thread/start is sent SECOND and its returned thread.id is threaded into
//     the review request as threadId.
//  3. review/start's target is the INTERNALLY TAGGED OBJECT
//     {"type":"uncommittedChanges"}, NOT a bare string.
//  4. the exchange is session-oriented: review/start is the THIRD request and
//     carries threadId matching the thread/start response.
//
// The live binary round-trip (real BLOCK on an injection+AWS-key fixture) is
// pinned separately in codex_review_gate_live_test.go.

// sentRequest parses a recorded sent NDJSON line into a generic map.
func sentRequest(t *testing.T, line string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("parse sent request %q: %v", line, err)
	}
	return m
}

// TestCodexRPC_InitializeFirstWithClientInfo pins gap #1: the FIRST sent request
// is initialize carrying a clientInfo {name,version} object.
func TestCodexRPC_InitializeFirstWithClientInfo(t *testing.T) {
	sess := withCodexSession(t, codexSessionScript("clean"))
	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart, map[string]any{"target": codexTargetUncommitted}); err != nil {
		t.Fatalf("rpc: %v", err)
	}
	if len(sess.sent) == 0 {
		t.Fatal("no requests were sent")
	}
	req := sentRequest(t, sess.sent[0])
	if m, _ := req["method"].(string); m != codexMethodInitialize {
		t.Errorf("first request method = %q, want %q", m, codexMethodInitialize)
	}
	params, _ := req["params"].(map[string]any)
	clientInfo, ok := params["clientInfo"].(map[string]any)
	if !ok {
		t.Fatalf("initialize params must carry clientInfo object; params=%v", params)
	}
	if clientInfo["name"] == "" || clientInfo["version"] == "" {
		t.Errorf("clientInfo must carry name+version; got %v", clientInfo)
	}
}

// TestCodexRPC_TargetIsTaggedObject pins gap #3: review/start's target is the
// object {"type":"uncommittedChanges"}, even when the caller passes a bare
// string (the legacy shape codex-cli 0.146.1 rejects with -32600).
func TestCodexRPC_TargetIsTaggedObject_BareStringLifted(t *testing.T) {
	// thread/start result carries thread.id = "tid-fake" (from codexSessionScript).
	sess := withCodexSession(t, codexSessionScript("clean"))
	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted, // bare string — must be lifted
	}); err != nil {
		t.Fatalf("rpc: %v", err)
	}
	if len(sess.sent) < 3 {
		t.Fatalf("expected ≥3 sent requests; got %d", len(sess.sent))
	}
	reviewReq := sentRequest(t, sess.sent[2]) // 3rd request = review/start
	if m, _ := reviewReq["method"].(string); m != codexMethodReviewStart {
		t.Fatalf("3rd request method = %q, want %q", m, codexMethodReviewStart)
	}
	params, _ := reviewReq["params"].(map[string]any)
	target, ok := params["target"].(map[string]any)
	if !ok {
		t.Fatalf("review/start target must be a JSON object; got %T (%v)", params["target"], params["target"])
	}
	if t2, _ := target["type"].(string); t2 != codexTargetUncommitted {
		t.Errorf("target.type = %q, want %q", t2, codexTargetUncommitted)
	}
	// gap #2: threadId present and matches thread/start's result.thread.id.
	threadID, _ := params["threadId"].(string)
	if threadID != "tid-fake" {
		t.Errorf("review/start threadId = %q, want %q (must match thread/start result.thread.id)", threadID, "tid-fake")
	}
}

// TestCodexRPC_BareStringTargetNotSentInReviewRequest asserts the bare string
// shape codex rejects ("target":"uncommittedChanges") is NEVER serialized into
// the review/start request — proving gap #3 is closed at the wire level.
func TestCodexRPC_BareStringTargetNotSentInReviewRequest(t *testing.T) {
	sess := withCodexSession(t, codexSessionScript("clean"))
	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart, map[string]any{"target": codexTargetUncommitted}); err != nil {
		t.Fatalf("rpc: %v", err)
	}
	if len(sess.sent) < 3 {
		t.Fatalf("expected ≥3 sent requests; got %d", len(sess.sent))
	}
	// The review/start line must NOT carry the bare-string shape codex rejects.
	bad := `"target":"` + codexTargetUncommitted + `"`
	if strings.Contains(sess.sent[2], bad) {
		t.Errorf("review/start must NOT serialize target as a bare string (codex rejects it); line:\n%s", sess.sent[2])
	}
}

// TestSynthesizeReviewOutput_FindingBulletsMapToFail pins the verdict synthesis:
// codex's severity-tagged finding bullets ("- [P1] ...") ⇒ fail; a clean review
// ⇒ pass. Grounded in codex-cli 0.146.1 live output on both directions.
func TestSynthesizeReviewOutput_FindingBulletsMapToFail(t *testing.T) {
	cases := map[string]string{
		"- [P1] injection at vuln.go:5\n- [P1] AWS key at vuln.go:7": "fail",
		"- [P2] minor style issue":                                   "fail",
		"The change introduces no blocking issues.":                  "pass",
		"": "pass",
	}
	for text, want := range cases {
		if got := synthesizeReviewOutput(text).Verdict; got != want {
			t.Errorf("synthesize(%q).Verdict = %q, want %q", text, got, want)
		}
	}
}

package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// SPEC-CODEX-PHASE2-001 M1 — reusable session handle (REQ-CX2-001) + model/effort
// SSOT wiring (REQ-CX2-002). Verifies AC-CX2-001, AC-CX2-003, AC-CX2-004.
//
// The transport, the id matching, the error arm, and the turn loop already exist
// and are pinned by codex_review_rpc_test.go / codex_rpc_error_test.go (AP-1);
// these tests exercise the SPLIT (a handshake reachable as a reusable session)
// and the destination the resolved model/effort actually lands in.

// codexTwoTurnScript builds a canned NDJSON transcript that answers the
// handshake (ids 1-2) and TWO successive turns (ids 3 and 4) on the same thread.
// The second turn's transcript is what a reused session must be able to drive
// without repeating initialize / thread/start.
func codexTwoTurnScript(firstReview, secondReview string) []string {
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn-1","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn-1","completedAtMs":1,"item":{"type":"exitedReviewMode","id":"e1","review":` + jsonString(firstReview) + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn-1","status":"completed"}}}`,
		`{"id":4,"result":{"turn":{"id":"trn-2","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn-2","completedAtMs":2,"item":{"type":"exitedReviewMode","id":"e2","review":` + jsonString(secondReview) + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn-2","status":"completed"}}}`,
	}
}

// countSentMethod counts how many recorded request lines carry the given method.
func countSentMethod(t *testing.T, sent []string, method string) int {
	t.Helper()
	n := 0
	for _, line := range sent {
		if m, _ := sentRequest(t, line)["method"].(string); m == method {
			n++
		}
	}
	return n
}

// sentParams returns the params object of the i-th recorded request line.
func sentParams(t *testing.T, sent []string, i int) map[string]any {
	t.Helper()
	if i >= len(sent) {
		t.Fatalf("request index %d out of range (%d sent)", i, len(sent))
	}
	params, _ := sentRequest(t, sent[i])["params"].(map[string]any)
	return params
}

// AC-CX2-001 — a second turn is issuable on the returned session handle without
// repeating the handshake: exactly one initialize, exactly one thread/start,
// both turns carrying the same threadId, and both returning their result.
func TestCodexSession_SecondTurnReusesThread(t *testing.T) {
	sess := withCodexSession(t, codexTwoTurnScript("clean, no findings", "- [P1] second turn found an issue"))

	handle, err := openCodexSession(context.Background(), "/fake/codex", map[string]any{"cwd": t.TempDir()})
	if err != nil {
		t.Fatalf("openCodexSession: %v", err)
	}
	defer func() { _ = handle.close() }()

	if handle.threadID != "tid-fake" {
		t.Fatalf("session threadID = %q, want %q (retained from thread/start)", handle.threadID, "tid-fake")
	}

	first, err := handle.runTurn(context.Background(), codexMethodReviewStart, map[string]any{"target": codexTargetUncommitted})
	if err != nil {
		t.Fatalf("first turn: %v", err)
	}
	second, err := handle.runTurn(context.Background(), codexMethodReviewStart, map[string]any{"target": codexTargetUncommitted})
	if err != nil {
		t.Fatalf("second turn: %v", err)
	}

	if first.Verdict != "pass" {
		t.Errorf("first turn verdict = %q, want pass", first.Verdict)
	}
	if second.Verdict != "fail" {
		t.Errorf("second turn verdict = %q, want fail", second.Verdict)
	}

	if n := countSentMethod(t, sess.sent, codexMethodInitialize); n != 1 {
		t.Errorf("initialize sent %d times, want exactly 1 (the handshake is not repeated)", n)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 1 {
		t.Errorf("thread/start sent %d times, want exactly 1 (the thread is reused)", n)
	}
	if n := countSentMethod(t, sess.sent, codexMethodReviewStart); n != 2 {
		t.Errorf("review/start sent %d times, want 2 (one per turn)", n)
	}

	// Both turns address the SAME thread.
	for _, i := range []int{2, 3} {
		if got, _ := sentParams(t, sess.sent, i)["threadId"].(string); got != "tid-fake" {
			t.Errorf("request %d threadId = %q, want tid-fake", i, got)
		}
	}
}

// writeCodexLLMFixture writes a minimal llm.yaml carrying a per-agent override
// for the codex audit agent key, and points the project-dir seam at it.
func writeCodexLLMFixture(t *testing.T, model, effort string) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	yaml := "llm:\n  agent_overrides:\n    " + codexAuditAgentKey + ":\n      model: " + model + "\n      effort: " + effort + "\n"
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	prev := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = prev })
	return root
}

// AC-CX2-003 / AC-CX2-004 — the SSOT-resolved model reaches the params ACTUALLY
// transmitted to codex. thread/start is the destination for the session-level
// model (ReviewStartParams carries no model field at all); turn/start carries
// both model and effort. This is the positive counterpart to the vacuous
// TestMCPAudit_NoDirectFrontmatterRead guard (plan.md §B B3).
func TestCodexSession_ResolvedModelReachesTransmittedParams(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "high")
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("rpc: %v", err)
	}

	// thread/start (2nd request) carries the resolved model.
	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got != "gpt-5-codex" {
		t.Errorf("thread/start model = %q, want %q (resolved via ResolveAgentModelEffort)", got, "gpt-5-codex")
	}

	// turn/start (3rd request) carries BOTH model and effort.
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "gpt-5-codex" {
		t.Errorf("turn/start model = %q, want %q", got, "gpt-5-codex")
	}
	if got, _ := turn["effort"].(string); got != "high" {
		t.Errorf("turn/start effort = %q, want %q", got, "high")
	}
}

// AC-CX2-003 second arm — an explicit `model` argument is transmitted verbatim,
// closing the silent drop at buildCodexReviewParams (spec.md §A.3 G3).
func TestCodexSession_ExplicitModelOverridesResolved(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "high")
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"model":  "o4-mini",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("rpc: %v", err)
	}

	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != "o4-mini" {
		t.Errorf("thread/start model = %q, want the explicit override %q", got, "o4-mini")
	}
	if got, _ := sentParams(t, sess.sent, 2)["model"].(string); got != "o4-mini" {
		t.Errorf("turn/start model = %q, want the explicit override %q", got, "o4-mini")
	}
}

// C7 non-regression — the default profile matrix resolves the audit agent to a
// Claude model ("opus"), which the codex app-server cannot serve. Transmitting
// it would break the review gate for every project that has not opted in, so a
// non-codex-servable resolved model is dropped and the request stays
// byte-identical to the pre-M1 shape (no model, no effort).
func TestCodexSession_NonCodexModelNotTransmitted(t *testing.T) {
	root := writeCodexLLMFixture(t, "opus", "high")
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("rpc: %v", err)
	}

	if _, ok := sentParams(t, sess.sent, 1)["model"]; ok {
		t.Error("thread/start must omit a non-codex-servable model (codex cannot serve a Claude id)")
	}
	turn := sentParams(t, sess.sent, 2)
	if _, ok := turn["model"]; ok {
		t.Error("turn/start must omit a non-codex-servable model")
	}
	if _, ok := turn["effort"]; ok {
		t.Error("turn/start must omit the paired effort when the model is dropped (Claude effort vocabulary)")
	}
}

// review/start carries neither model nor effort: ReviewStartParams declares only
// {delivery, target, threadId} (codex-cli 0.146.1 generate-json-schema), so
// injecting either would be an unknown field on the gate's own request path.
func TestCodexSession_ReviewStartCarriesNoModelOrEffort(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "high")
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted,
		"cwd":    root,
	}); err != nil {
		t.Fatalf("rpc: %v", err)
	}

	review := sentParams(t, sess.sent, 2)
	for _, field := range []string{"model", "effort"} {
		if _, ok := review[field]; ok {
			t.Errorf("review/start must not carry %q — ReviewStartParams declares only delivery/target/threadId", field)
		}
	}
	// The session-level model still reaches codex via thread/start.
	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != "gpt-5-codex" {
		t.Errorf("thread/start model = %q, want %q (the only reachable destination on the review path)", got, "gpt-5-codex")
	}
}

// codexServableModel is the guard that keeps a non-codex model id off the wire.
func TestCodexServableModel(t *testing.T) {
	servable := []string{"gpt-5-codex", "gpt-4.1", "o3", "o4-mini", "codex-mini-latest"}
	notServable := []string{"", "opus", "sonnet", "haiku", "claude-opus-5", "glm-4.6"}
	for _, m := range servable {
		if !codexServableModel(m) {
			t.Errorf("codexServableModel(%q) = false, want true", m)
		}
	}
	for _, m := range notServable {
		if codexServableModel(m) {
			t.Errorf("codexServableModel(%q) = true, want false", m)
		}
	}
}

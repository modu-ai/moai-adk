package codexadapter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// decode is a small helper: map the adapter output back to a generic map so
// assertions read against the wire shape rather than the Go struct.
func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", b, err)
	}
	return m
}

// TestContinueFalseRewritesToDecisionBlock — AC-REQ-2a.
func TestContinueFalseRewritesToDecisionBlock(t *testing.T) {
	t.Parallel()

	in := []byte(`{"continue":false,"stopReason":"X"}`)
	out, _, err := MapOutput(hook.EventStop, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	got := decode(t, out)
	if got["decision"] != "block" {
		t.Errorf("decision = %v, want block", got["decision"])
	}
	if got["reason"] != "X" {
		t.Errorf("reason = %v, want X", got["reason"])
	}
	if _, ok := got["continue"]; ok {
		t.Error("continue survived the rewrite; Codex ignores it")
	}
	if _, ok := got["stopReason"]; ok {
		t.Error("stopReason survived the rewrite; Codex ignores it")
	}
}

// TestContinueFalseWithoutStopReasonGetsDefaultReason — AC-REQ-2b.
//
// Codex rejects a decision:block carrying an empty reason, so an empty reason
// would turn a block into a no-op.
func TestContinueFalseWithoutStopReasonGetsDefaultReason(t *testing.T) {
	t.Parallel()

	out, _, err := MapOutput(hook.EventStop, []byte(`{"continue":false}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	got := decode(t, out)
	if got["decision"] != "block" {
		t.Fatalf("decision = %v, want block", got["decision"])
	}
	reason, _ := got["reason"].(string)
	if strings.TrimSpace(reason) == "" {
		t.Fatal("reason is empty; Codex rejects decision:block without a non-empty reason")
	}
}

// TestContinueTrueIsNotRewritten asserts only the blocking form is rewritten.
func TestContinueTrueIsNotRewritten(t *testing.T) {
	t.Parallel()

	out, _, err := MapOutput(hook.EventStop, []byte(`{"continue":true}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if got := decode(t, out); got["decision"] != nil {
		t.Errorf("continue:true produced decision=%v; want no rewrite", got["decision"])
	}
}

// TestSystemMessageRoutedToAdditionalContext — AC-REQ-2c.
func TestSystemMessageRoutedToAdditionalContext(t *testing.T) {
	t.Parallel()

	out, discards, err := MapOutput(hook.EventUserPromptSubmit, []byte(`{"systemMessage":"hello"}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if len(discards) != 0 {
		t.Errorf("discards = %v, want none (UserPromptSubmit has a working channel)", discards)
	}

	got := decode(t, out)
	hso, ok := got["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput missing in %s", out)
	}
	if hso["additionalContext"] != "hello" {
		t.Errorf("additionalContext = %v, want hello", hso["additionalContext"])
	}
	if hso["hookEventName"] != "UserPromptSubmit" {
		t.Errorf("hookEventName = %v, want UserPromptSubmit", hso["hookEventName"])
	}
	if _, ok := got["systemMessage"]; ok {
		t.Error("systemMessage survived; Codex ignores it")
	}
}

// TestSystemMessageDiscardedWhereNoChannel — AC-REQ-3a.
//
// The discard is a real reduction in advisory capability, not a lossless
// translation. It must be announced.
func TestSystemMessageDiscardedWhereNoChannel(t *testing.T) {
	t.Parallel()

	out, discards, err := MapOutput(hook.EventPostToolUse, []byte(`{"systemMessage":"advisory text"}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if got := decode(t, out); len(got) != 0 {
		t.Errorf("output = %s, want empty object", out)
	}

	if len(discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(discards))
	}
	d := discards[0]
	if d.Event != hook.EventPostToolUse {
		t.Errorf("discard event = %s, want PostToolUse", d.Event)
	}
	if d.Key != "systemMessage" {
		t.Errorf("discard key = %s, want systemMessage", d.Key)
	}
	if d.ContentLength != len("advisory text") {
		t.Errorf("discard length = %d, want %d", d.ContentLength, len("advisory text"))
	}
}

// TestDiscardRecordCarriesNoContent — AC-REQ-3a.
//
// Length rather than content keeps the diagnostic from becoming an
// exfiltration path for whatever the hook was reporting.
func TestDiscardRecordCarriesNoContent(t *testing.T) {
	t.Parallel()

	const secret = "TOKEN-abc123-SHOULD-NOT-APPEAR"
	_, discards, err := MapOutput(hook.EventPostToolUse, []byte(`{"systemMessage":"`+secret+`"}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if len(discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(discards))
	}

	line, err := json.Marshal(discards[0])
	if err != nil {
		t.Fatalf("marshal discard: %v", err)
	}
	if strings.Contains(string(line), secret) {
		t.Fatalf("discard record leaked the content: %s", line)
	}
}

// TestWorkingKeysPassThroughByteIdentical — AC-REQ-2d.
//
// Every unnecessary translation is a drift point between the harnesses.
func TestWorkingKeysPassThroughByteIdentical(t *testing.T) {
	t.Parallel()

	for _, in := range []string{
		`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"nope"}}`,
		`{"decision":"block","reason":"keep going"}`,
		`{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"ctx"}}`,
	} {
		out, discards, err := MapOutput(hook.EventPreToolUse, []byte(in))
		if err != nil {
			t.Fatalf("MapOutput(%s) error = %v", in, err)
		}
		if len(discards) != 0 {
			t.Errorf("MapOutput(%s) discarded %v; want none", in, discards)
		}
		if string(out) != in {
			t.Errorf("MapOutput(%s)\n got %s\nwant byte-identical", in, out)
		}
	}
}

// TestDiscardBranchCountMatchesTested — AC-REQ-3b.
//
// The earlier phrasing claimed a later-added branch would fail the AC, which
// nothing enforced. The shared constant makes adding a branch without a test a
// failure here.
func TestDiscardBranchCountMatchesTested(t *testing.T) {
	t.Parallel()

	tested := []string{"systemMessage", "continue", "stopReason"}
	if len(tested) != DiscardBranchCount {
		t.Fatalf("tested discard branches = %d, DiscardBranchCount = %d — a branch was added without a test",
			len(tested), DiscardBranchCount)
	}

	for _, key := range tested {
		if !isDiscardableKey(key) {
			t.Errorf("%q is not a known discardable key", key)
		}
	}
}

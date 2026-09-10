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
	// The PreToolUse decision branch (card t590): allow/ask/defer dropped,
	// empty-reason deny repaired, dangling reason dropped.
	decisionBranches := []string{"preToolUseDecision"}
	if len(tested)+len(decisionBranches) != DiscardBranchCount {
		t.Fatalf("tested discard branches = %d, DiscardBranchCount = %d — a branch was added without a test",
			len(tested)+len(decisionBranches), DiscardBranchCount)
	}

	for _, key := range tested {
		if !isDiscardableKey(key) {
			t.Errorf("%q is not a known discardable key", key)
		}
	}
}

// TestPreToolUseAllowWithoutUpdatedInputDropped — card t590.
//
// Codex's PreToolUse parser (codex-rs hooks/src/engine/output_parser.rs,
// unsupported_pre_tool_use_hook_specific_output) rejects
// permissionDecision:allow that carries no updatedInput. MoAI's safe-path
// allow has no updatedInput, so under Codex every safe PreToolUse verdict was
// reported as an invalid hook output and the pre-check was skipped. Dropping
// the decision degrades to a no-opinion `{}`, which hands the choice to
// Codex's own approval flow.
func TestPreToolUseAllowWithoutUpdatedInputDropped(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	if got := decode(t, out); len(got) != 0 {
		t.Errorf("output = %s, want empty object", out)
	}
	if len(discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(discards))
	}
	if discards[0].Event != hook.EventPreToolUse {
		t.Errorf("discard event = %s, want PreToolUse", discards[0].Event)
	}
	if discards[0].Key != "hookSpecificOutput" {
		t.Errorf("discard key = %s, want hookSpecificOutput", discards[0].Key)
	}
}

// TestPreToolUseAllowDegradationRecordedNotSilent — card t590.
//
// The no-silence obligation: a dropped allow is a real reduction in control
// capability and must be announced through the discard record.
func TestPreToolUseAllowDegradationRecordedNotSilent(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"},"systemMessage":"scan clean"}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	if got := decode(t, out); len(got) != 0 {
		t.Errorf("output = %s, want empty object (systemMessage has no channel on PreToolUse either)", out)
	}
	if len(discards) != 2 {
		t.Fatalf("discards = %d, want 2 (decision drop + systemMessage)", len(discards))
	}
	keys := map[string]bool{}
	for _, d := range discards {
		keys[d.Key] = true
	}
	if !keys["hookSpecificOutput"] || !keys["systemMessage"] {
		t.Errorf("discard keys = %v, want hookSpecificOutput + systemMessage", keys)
	}
}

// TestPreToolUseAskDropped — card t590.
//
// Codex always rejects permissionDecision:ask on PreToolUse. Dropping it maps
// to no-opinion, which under Codex means the normal approval flow decides —
// the same semantics ask carries on Claude Code.
func TestPreToolUseAskDropped(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"ask","permissionDecisionReason":"confirm?"}}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	if got := decode(t, out); len(got) != 0 {
		t.Errorf("output = %s, want empty object", out)
	}
	if len(discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(discards))
	}
}

// TestPreToolUseDenyWithoutReasonGetsDefaultReason — card t590.
//
// Codex accepts permissionDecision:deny only with a non-empty
// permissionDecisionReason. A blank-reason deny would be rejected outright, so
// the adapter fills a default rather than losing the block.
func TestPreToolUseDenyWithoutReasonGetsDefaultReason(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":""}}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	if len(discards) != 0 {
		t.Errorf("discards = %v, want none (deny is repaired, not dropped)", discards)
	}
	got := decode(t, out)
	hso, ok := got["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput missing in %s", out)
	}
	if hso["permissionDecision"] != "deny" {
		t.Errorf("permissionDecision = %v, want deny (kept)", hso["permissionDecision"])
	}
	reason, _ := hso["permissionDecisionReason"].(string)
	if strings.TrimSpace(reason) == "" {
		t.Fatal("permissionDecisionReason still empty; Codex rejects a blank-reason deny")
	}
}

// TestPreToolUseAllowWithUpdatedInputPassesThrough — card t590.
//
// allow WITH updatedInput is the one valid allow form on Codex. MoAI never
// emits it today; the adapter must not translate it if it ever appears.
func TestPreToolUseAllowWithUpdatedInputPassesThrough(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"command":"ls"}}}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if len(discards) != 0 {
		t.Errorf("discards = %v, want none", discards)
	}
	if string(out) != string(in) {
		t.Errorf("got %s, want byte-identical pass-through", out)
	}
}

// TestPreToolUseDanglingReasonDropped — card t590.
//
// Codex rejects permissionDecisionReason without permissionDecision. MoAI
// factories always pair them, so this is defensive — but an unpaired reason
// must not survive to a guaranteed-invalid output.
func TestPreToolUseDanglingReasonDropped(t *testing.T) {
	t.Parallel()

	in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecisionReason":"orphaned"}}`)
	out, discards, err := MapOutput(hook.EventPreToolUse, in)
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	got := decode(t, out)
	hso, ok := got["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput missing in %s", out)
	}
	if _, ok := hso["permissionDecisionReason"]; ok {
		t.Error("permissionDecisionReason survived without a permissionDecision; Codex rejects it")
	}
	if len(discards) != 0 {
		t.Errorf("discards = %v, want none (repair, not a dropped message)", discards)
	}
}

// TestPreToolUseEmptyObjectPassesThroughByteIdentical — card t590.
//
// The default/plan-mode safe path already emits the no-opinion form; the
// adapter must not re-marshal it.
func TestPreToolUseEmptyObjectPassesThroughByteIdentical(t *testing.T) {
	t.Parallel()

	out, discards, err := MapOutput(hook.EventPreToolUse, []byte(`{}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}
	if len(discards) != 0 {
		t.Errorf("discards = %v, want none", discards)
	}
	if string(out) != `{}` {
		t.Errorf("got %s, want byte-identical {}", out)
	}
}

// TestPreToolUseContinueFalseStillBecomesDecisionBlock — card t590.
//
// The existing continue:false rewrite and the new decision handling must
// compose: stripping the universal keys AND yielding a Codex-valid block.
func TestPreToolUseContinueFalseStillBecomesDecisionBlock(t *testing.T) {
	t.Parallel()

	out, _, err := MapOutput(hook.EventPreToolUse, []byte(`{"continue":false,"stopReason":"dangerous command"}`))
	if err != nil {
		t.Fatalf("MapOutput error = %v", err)
	}

	got := decode(t, out)
	if got["decision"] != "block" {
		t.Errorf("decision = %v, want block", got["decision"])
	}
	if got["reason"] != "dangerous command" {
		t.Errorf("reason = %v, want dangerous command", got["reason"])
	}
	if _, ok := got["hookSpecificOutput"]; ok {
		t.Error("hookSpecificOutput appeared on a universal-key rewrite")
	}
}

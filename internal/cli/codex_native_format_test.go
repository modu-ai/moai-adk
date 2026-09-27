package cli

// SPEC-CODEX-PARSER-SHAPE-001 M4 — candidate (b): the native review request
// pins its output format (REQ-CPS-005 as amended) and the native fall-through
// reports a body carrying no recognized signal as inconclusive instead of a
// silent pass (AC-CPS-004).
//
// What these tests establish is the OFFLINE half only: the native request
// asks for the pinned format, the format it asks for is one the parser reads,
// a body following the pin stays pass (REQ-CPS-009's protected class), and a
// no-signal body is downgraded. Whether live codex follows the request
// against a target project's own instructions is AC-CPS-016, which only a
// recorded live observation can decide — no test in this tree can satisfy it.

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCodexNativeFormatPin_SharesConstantsWithBuilder ties the native pin to
// the (d) family mechanically: the native pin instruction is built FROM the
// adversarial pin constants, so both requests ask for one format and a drift
// between the pin and the recognizers cannot happen on one path alone
// (AC-CPS-004's green path, "sharing constants with the request builder").
func TestCodexNativeFormatPin_SharesConstantsWithBuilder(t *testing.T) {
	for _, want := range []string{
		codexAdversarialVerdictFormat,
		codexAdversarialFindingFormat,
		"Verdict: pass", "Verdict: fail", "Verdict: inconclusive",
	} {
		if !strings.Contains(codexNativeReviewFormatPin, want) {
			t.Errorf("native pin does not carry %q\npin: %s", want, codexNativeReviewFormatPin)
		}
	}
}

// TestCodexNativeRequest_CarriesFormatPin is the carrier half of the pin:
// ReviewStartParams declares only {delivery, target, threadId} (measured,
// codex-cli 0.157.0 generate-json-schema), so — exactly as the session model
// rides thread/start (REQ-CX2-002) — the format instruction must reach codex
// on the thread the native review opens, as
// ThreadStartParams.developerInstructions. The review target is untouched (the
// custom-substitution route is forbidden by REQ-CPS-005), and the adversarial
// path, which pins its own format in the turn prompt, carries no native pin.
func TestCodexNativeRequest_CarriesFormatPin(t *testing.T) {
	threadParamsFor := func(t *testing.T, method string) map[string]any {
		t.Helper()
		sess := withCodexSession(t, codexSessionScript("clean"))
		if _, err := runCodexReviewRPC(t.Context(), "/fake/codex", method, map[string]any{"target": codexTargetUncommitted}); err != nil {
			t.Fatalf("rpc: %v", err)
		}
		for _, line := range sess.sent {
			var req struct {
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				t.Fatalf("unmarshal sent line: %v", err)
			}
			if req.Method == codexMethodThreadStart {
				return req.Params
			}
		}
		t.Fatalf("no %s request among %d sent lines", codexMethodThreadStart, len(sess.sent))
		return nil
	}

	// Native: the pin rides the thread the review runs on.
	params := threadParamsFor(t, codexMethodReviewStart)
	instr, ok := params["developerInstructions"].(string)
	if !ok || instr == "" {
		t.Fatalf("native thread/start params carry no developerInstructions: %v", params)
	}
	for _, want := range []string{"Verdict: pass", "Verdict: fail", "Verdict: inconclusive", "- [P1]"} {
		if !strings.Contains(instr, want) {
			t.Errorf("native developerInstructions does not carry %q\ninstruction: %s", want, instr)
		}
	}

	// The review/start request itself stays schema-clean: the pin does not
	// leak onto the request the gate's schema cannot carry.
	sess := withCodexSession(t, codexSessionScript("clean"))
	if _, err := runCodexReviewRPC(t.Context(), "/fake/codex", codexMethodReviewStart, map[string]any{"target": codexTargetUncommitted}); err != nil {
		t.Fatalf("rpc: %v", err)
	}
	for _, line := range sess.sent {
		var req struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			t.Fatalf("unmarshal sent line: %v", err)
		}
		if req.Method == codexMethodReviewStart {
			if _, leak := req.Params["developerInstructions"]; leak {
				t.Errorf("review/start params carry developerInstructions; ReviewStartParams declares only {delivery, target, threadId}: %v", req.Params)
			}
		}
	}

	// Adversarial: its pin lives in the turn prompt, not on the thread.
	if got := threadParamsFor(t, codexMethodTurnStart); got["developerInstructions"] != nil {
		t.Errorf("adversarial thread/start carries developerInstructions %q; the (d) pin rides the turn prompt", got["developerInstructions"])
	}
}

// TestCodexNativeFormat_IsRecognized ties the pinned format to the
// recognizers mechanically: a body written exactly in the format the native
// request asks for synthesizes with the stated verdict and every finding
// structured on the native path.
func TestCodexNativeFormat_IsRecognized(t *testing.T) {
	fill := func(sev, msg, path, line string) string {
		return strings.NewReplacer("P1", sev, "<message>", msg, "<path>", path, "<line>", line).
			Replace(codexAdversarialFindingFormat)
	}
	body := strings.Replace(codexAdversarialVerdictFormat, "<pass|fail|inconclusive>", "fail", 1) + "\n" +
		fill("P1", "unchecked error", "internal/x.go", "12") + "\n" +
		"  the error from Close is dropped\n"

	out := synthesizeReviewOutput(body, codexMethodReviewStart)
	if out.Verdict != "fail" {
		t.Errorf("verdict = %q, want fail\nbody:\n%s", out.Verdict, body)
	}
	if len(out.Findings) != 1 {
		t.Fatalf("want exactly 1 finding, got %d (%+v)\nbody:\n%s", len(out.Findings), out.Findings, body)
	}
	if f := out.Findings[0]; f.Severity != "P1" || f.File != "internal/x.go" || f.Line != 12 {
		t.Errorf("finding[0] = %+v", f)
	}
	if out.Contradiction != "" {
		t.Errorf("a conforming body reported as contradictory: %q", out.Contradiction)
	}
}

// TestCodexNativeNoSignalBodyYieldsInconclusive is AC-CPS-004's second Then
// clause: a body carrying no recognized signal — no verdict statement in a
// form any recognizer reads, no recognizable findings — is reported
// inconclusive on the native path, not as a silent pass. The downgrade is
// keyed on the absence of a recognized signal and on nothing in the prose.
func TestCodexNativeNoSignalBodyYieldsInconclusive(t *testing.T) {
	for _, body := range []string{
		"I walked the diff and moved on.",
		"The change introduces no identifiable correctness or blocking issues.",
		"차단 사유 2건을 확인했습니다.",
	} {
		if got := synthesizeReviewOutput(body, codexMethodReviewStart).Verdict; got != VerdictInconclusive {
			t.Errorf("no-signal native body %q: Verdict = %q, want %q", body, got, VerdictInconclusive)
		}
	}
}

// TestCodexNativePinnedPassStaysPass is AC-CPS-004's first Then clause —
// REQ-CPS-009's protected class expressed through the pin: a body stating
// `Verdict: pass` in the pinned form, with no findings, yields pass. This is
// the test mutant M-A (downgrade every no-signal native body, the pinned pass
// included) must fail.
func TestCodexNativePinnedPassStaysPass(t *testing.T) {
	body := "Verdict: pass\n\nNo blocking issues found."
	out := synthesizeReviewOutput(body, codexMethodReviewStart)
	if out.Verdict != "pass" || len(out.Findings) != 0 || out.Contradiction != "" {
		t.Errorf("pinned pass body: got %s/%d contradiction=%q, want pass/0 with no contradiction",
			out.Verdict, len(out.Findings), out.Contradiction)
	}
}

package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M3 RED — the delta codex leg (AC-AEC-001a;
// REQ-AEC-003; card t2017).
//
// While a delta base is in effect for an audit_multi call, the codex leg keeps its
// adversarial turn/start and carries in that turn the delta_base...HEAD diff the
// SERVER collected, not an instruction to compute it; it does not switch to the
// native review target, and the result does not describe the codex scope as
// prose-only.
//
// The test drives the real leg: handleAuditMulti with the default backend caller,
// so performCodexAudit runs against the fake codex session and the serialized
// request lines are what is observed. Today the turn carries only the prompt plus
// an instruction (`git diff <sha>...HEAD`), and the note says "prose only".

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
)

// deltaDiffBytes returns `git diff <base>...HEAD` exactly as the server collects it.
func deltaDiffBytes(t *testing.T, root, base string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "diff", base+"...HEAD")
	cmd.Env = closuretest.ScrubbedEnv()
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff %s...HEAD: %v", base, err)
	}
	return string(out)
}

// sentMethods lists the method of every request line the fake session received.
func sentMethods(sess *fakeCodexSession) []string {
	var out []string
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil {
			continue
		}
		if method, _ := m["method"].(string); method != "" {
			out = append(out, method)
		}
	}
	return out
}

// sentTurnParams returns the params of the turn/start request, and whether one was sent.
func sentTurnParams(sess *fakeCodexSession) (map[string]any, bool) {
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil {
			continue
		}
		if method, _ := m["method"].(string); method == codexMethodTurnStart {
			p, _ := m["params"].(map[string]any)
			return p, true
		}
	}
	return nil, false
}

// AC-AEC-001a — the codex leg of a delta audit_multi call sends the adversarial
// turn carrying the server-collected diff bytes and no native review target, and
// the result states no prose-only scope.
func TestAEC001a_DeltaCodexLeg_CarriesTheServerCollectedDiffInItsAdversarialTurn(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())
	prev := backendCall
	backendCall = defaultBackendCaller // the REAL legs; only codex is gated on
	t.Cleanup(func() { backendCall = prev })

	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"target": codexTargetBaseBranch, "project_root": d.root, "delta_base": d.shaB,
		"gates": map[string]any{"claude": "off", "codex": "required", "glm": "off"},
	}}})
	if err != nil {
		t.Fatalf("handleAuditMulti: %v", err)
	}
	if res.IsError {
		t.Fatalf("a valid delta_base must not be a tool error: %v", res.Content)
	}
	var raw string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			raw += tc.Text
		}
	}
	var result ConvergenceResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("decode result: %v\n%s", err, raw)
	}

	// Control (passes today): the leg is the adversarial turn, never the native
	// review. turn/start has no target field, and no review/start was sent.
	methods := sentMethods(sess)
	if _, ok := sentReviewStart(t, sess); ok {
		t.Errorf("the delta codex leg must not switch to the native review target; methods sent: %v", methods)
	}
	params, ok := sentTurnParams(sess)
	if !ok {
		t.Fatalf("the delta codex leg must send its adversarial turn/start; methods sent: %v", methods)
	}
	if _, has := params["target"]; has {
		t.Errorf("the adversarial turn must carry no native review target: %v", params)
	}

	// The diff the server collected is IN the turn.
	prompt := sentTurnPrompt(t, sess)
	diff := deltaDiffBytes(t, d.root, d.shaB)
	if strings.TrimSpace(diff) == "" {
		t.Fatalf("setup: the delta %s...HEAD diff is empty", d.shaB)
	}
	if !strings.Contains(prompt, diff) {
		t.Errorf("the turn does not carry the server-collected %s...HEAD diff bytes.\nwant (contained):\n%s\nturn prompt:\n%s", d.shaB, diff, prompt)
	}
	if !strings.Contains(prompt, "canary_c_delta") {
		t.Errorf("the turn does not carry the delta content (canary_c_delta)")
	}
	if strings.Contains(prompt, "canary_b_preexisting") {
		t.Errorf("the turn carries content that predates the delta base (canary_b_preexisting)")
	}
	// An instruction to compute the diff is not the diff.
	if strings.Contains(prompt, "git diff "+d.shaB) {
		t.Errorf("the turn still asks codex to compute the diff itself: %q", prompt)
	}

	// The result states no prose-only scope, anywhere it describes the codex scope.
	var codexReviewBase string
	for _, v := range result.PerBackendVerdicts {
		if v.Backend == BackendCodex {
			codexReviewBase = v.ReviewBase
		}
	}
	for where, text := range map[string]string{"residual_risk_note": result.ResidualRiskNote, "the codex review_base": codexReviewBase} {
		for _, bad := range []string{"prose-only", "prose only"} {
			if strings.Contains(text, bad) {
				t.Errorf("%s still describes the codex scope as %q: %s", where, bad, text)
			}
		}
	}
	if !strings.Contains(codexReviewBase, d.shaB) || !strings.Contains(codexReviewBase, deltaBaseRef) {
		t.Errorf("the codex review_base must name the delta base %s; got %q", d.shaB, codexReviewBase)
	}
}

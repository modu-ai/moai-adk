package cli

import (
	"context"
	"strings"
	"testing"
)

// SPEC-V3R6-AUDIT-MODEL-PIN-001 M2 — the codex audit pin (workflow.audit.codex)
// precedence at the two transmission seams, and the REQ-AMP-008 task-path
// isolation. The canned-session + sent-line harness from
// SPEC-CODEX-PHASE2-001 is reused verbatim (AP-1): these tests exercise the
// RESOLUTION layer over the existing transport, never a second client.

// auditCodexPinYAML is a workflow.yaml body carrying only the codex audit pin.
func auditCodexPinYAML(model, effort string) string {
	return "workflow:\n  audit:\n    codex:\n      model: " + model + "\n      effort: " + effort + "\n"
}

// AC-AMP-002 positive arm — a populated audit.codex pin reaches BOTH
// transmission seams on the codex AUDIT path: thread/start carries the pinned
// model (the session-level destination), turn/start carries model + effort.
func TestCodexAuditPin_ReachesTransmittedParams(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got != "gpt-6.1-sol" {
		t.Errorf("thread/start model = %q, want the pinned %q", got, "gpt-6.1-sol")
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "gpt-6.1-sol" {
		t.Errorf("turn/start model = %q, want the pinned %q", got, "gpt-6.1-sol")
	}
	if got, _ := turn["effort"].(string); got != "high" {
		t.Errorf("turn/start effort = %q, want the pinned %q", got, "high")
	}
}

func TestCodexAuditPin_UnservableFallsBackToBackendDefault(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "medium")
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("opus", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("thread/start model = %q, want the fallback pin %q", got, codexAuditDefaultModel)
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("turn/start model = %q, want the fallback pin %q", got, codexAuditDefaultModel)
	}
	if got, _ := turn["effort"].(string); got != codexAuditDefaultEffort {
		t.Errorf("turn/start effort = %q, want the fallback pin %q", got, codexAuditDefaultEffort)
	}
}

// Edge case (§D.2) — effort set with an EMPTY model is no pin (the model is
// the gate): resolution falls through to the audit terminal fallback
// ({gpt-6.1-terra, xhigh} since the card t1980 scope expansion — previously
// "no model, no effort").
func TestCodexAuditPin_EffortAlonePinsNothing(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "medium")
	writeCodexWorkflowYAML(t, root, "workflow:\n  audit:\n    codex:\n      model: \"\"\n      effort: high\n")
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("thread/start model = %q, want the fallback pin %q (empty pin model = no pin)", got, codexAuditDefaultModel)
	}
	if got, _ := sentParams(t, sess.sent, 2)["effort"].(string); got != codexAuditDefaultEffort {
		t.Errorf("turn/start effort = %q, want the fallback pin %q (the pin effort is inert without a model)", got, codexAuditDefaultEffort)
	}
}

// M2 — an explicit caller `model` argument outranks the pin, mirroring the
// existing resolveCodexModelEffort precedence rule.
func TestCodexAuditPin_ExplicitModelOverridesPin(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"model":  "o4-mini",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != "o4-mini" {
		t.Errorf("thread/start model = %q, want the explicit override %q", got, "o4-mini")
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "o4-mini" {
		t.Errorf("turn/start model = %q, want the explicit override %q", got, "o4-mini")
	}
	if got, _ := turn["effort"].(string); got != "high" {
		t.Errorf("turn/start effort = %q, want the pinned %q (the explicit model overrides only the model)", got, "high")
	}
}

// Card t1980 scope expansion — an explicit caller `effort` argument outranks
// the pinned effort (the per-call override that makes `max` selectable per
// invocation), exactly as the explicit `model` argument does above. The pinned
// model is untouched by an effort-only override.
func TestCodexAuditPin_ExplicitEffortOverridesPin(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"effort": "max",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["effort"].(string); got != "max" {
		t.Errorf("turn/start effort = %q, want the explicit override %q", got, "max")
	}
	if got, _ := turn["model"].(string); got != "gpt-6.1-sol" {
		t.Errorf("turn/start model = %q, want the pinned %q (an effort-only override must not move the model)", got, "gpt-6.1-sol")
	}
}

// Card t1980 scope expansion — the effort override also applies on the
// terminal-fallback arm (absent pin): the caller's effort lands on the
// fallback pin's model instead of being dropped.
func TestCodexAuditPin_ExplicitEffortOverridesFallback(t *testing.T) {
	root := t.TempDir() // no workflow.yaml → the terminal fallback arm
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"effort": "max",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["effort"].(string); got != "max" {
		t.Errorf("turn/start effort = %q, want the explicit override %q on the fallback arm", got, "max")
	}
	if got, _ := turn["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("turn/start model = %q, want the fallback pin %q", got, codexAuditDefaultModel)
	}
}

// Card t1980 audit repair P2 (a) — an explicit model AND an explicit effort
// together are BOTH honored when the stored pin is ABSENT: the explicit-model
// branch (resolveCodexModelEffort) must not discard the effort override.
// (With a servable pin present, the pin arm already applies both overrides.)
func TestCodexAuditPin_ExplicitModelAndEffortTogether(t *testing.T) {
	root := t.TempDir() // no workflow.yaml → the explicit-model branch
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"model":  "o4-mini",
		"effort": "max",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "o4-mini" {
		t.Errorf("turn/start model = %q, want the explicit override %q", got, "o4-mini")
	}
	if got, _ := turn["effort"].(string); got != "max" {
		t.Errorf("turn/start effort = %q, want the explicit override %q (the explicit-model branch must not drop it)", got, "max")
	}
}

// Card t1980 audit repair P2 (b) — 'none' is a member of terra's
// reasoning.effort vocabulary and must survive the explicit-model branch
// verbatim (an empty turn/start effort field would mean "not sent", not none).
func TestCodexAuditPin_ExplicitModelWithEffortNone(t *testing.T) {
	root := t.TempDir() // no workflow.yaml → the explicit-model branch
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"model":  "o4-mini",
		"effort": "none",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	if got, _ := sentParams(t, sess.sent, 2)["effort"].(string); got != "none" {
		t.Errorf("turn/start effort = %q, want %q", got, "none")
	}
}

// Card t1980 audit repair P2 (c) — the effort closed-set validation fires in
// EVERY resolution branch: the pin path (override AND stored pin effort), the
// explicit-model path, and the terminal-fallback path all reject a non-member
// instead of transmitting it.
func TestCodexAuditPin_EffortClosedSetValidationEveryBranch(t *testing.T) {
	cases := []struct {
		name   string
		pin    string // auditCodexPinYAML body, "" = no workflow.yaml
		model  string
		effort string
	}{
		{"pin path rejects a bad override", auditCodexPinYAML("gpt-6.1-sol", "high"), "", "ultra"},
		{"pin path rejects a bad stored pin effort", auditCodexPinYAML("gpt-6.1-sol", "ultra"), "", ""},
		{"explicit-model path rejects a bad override", auditCodexPinYAML("gpt-6.1-sol", "high"), "o4-mini", "ultra"},
		{"fallback path rejects a bad override", "", "", "ultra"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.pin != "" {
				writeCodexWorkflowYAML(t, root, tc.pin)
			}
			params := map[string]any{"prompt": "review this", "cwd": root}
			if tc.model != "" {
				params["model"] = tc.model
			}
			if tc.effort != "" {
				params["effort"] = tc.effort
			}
			_, err := resolveCodexAuditModelEffort(params)
			if err == nil {
				t.Fatalf("resolveCodexAuditModelEffort(%v) = no error, want a closed-set rejection", params)
			}
			if !strings.Contains(err.Error(), "effort") || !strings.Contains(err.Error(), "closed set") {
				t.Errorf("error %q does not name the effort closed set", err)
			}
		})
	}
}

// Card t1980 audit repair P2 (c), valid arm — every vocabulary member passes
// the pin path, none included.
func TestCodexAuditPin_EffortClosedSetAcceptsEveryMember(t *testing.T) {
	for _, effort := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
		root := t.TempDir()
		writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", "high"))
		me, err := resolveCodexAuditModelEffort(map[string]any{"cwd": root, "effort": effort})
		if err != nil {
			t.Fatalf("effort %q: %v", effort, err)
		}
		if me.Effort != effort {
			t.Errorf("effort %q: resolved %+v, want the override kept", effort, me)
		}
	}
}

// AC-AMP-003 states (a)/(b), retargeted by SPEC-MODEL-MATRIX-UPDATE-001
// REQ-MMU-001 — no pin sub-keys / absent workflow.yaml: the audit path lands
// on the terminal fallback pin ({gpt-6.1-terra, xhigh} since the card t1980
// scope expansion; previously the byte-free legacy shape; the task path still
// resolves pin-free — REQ-AMP-008).
func TestCodexAuditPin_AbsentPinLandsOnDefaultPin(t *testing.T) {
	root := t.TempDir() // no workflow.yaml, no llm.yaml → the terminal fallback applies
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("thread/start model = %q, want the fallback pin %q", got, codexAuditDefaultModel)
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != codexAuditDefaultModel {
		t.Errorf("turn/start model = %q, want the fallback pin %q", got, codexAuditDefaultModel)
	}
	if got, _ := turn["effort"].(string); got != codexAuditDefaultEffort {
		t.Errorf("turn/start effort = %q, want the fallback pin %q", got, codexAuditDefaultEffort)
	}
}

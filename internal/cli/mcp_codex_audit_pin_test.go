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

// Card t1980 CI repair (leader ruling §15.125): the per-call effort MCP
// parameter was REMOVED — the audit-tool input schemas are snapshot-frozen
// (TestSelfReview_AuditToolSchemasUnchanged) and the reasoning effort is a
// CONFIG-KEY surface only. These tests exercise that config-key path: the
// workflow.audit.codex.effort file pin reaches the resolver and is
// transmitted verbatim, and a hand-edited non-member pin effort is rejected.

// TestCodexAuditPin_ConfigKeyEffortTransmitted pins the config-key path over
// the terra vocabulary: a file pin carrying each member of the closed set is
// transmitted verbatim on the turn/start effort field (none included — an
// absent field would mean not sent, not none).
func TestCodexAuditPin_ConfigKeyEffortTransmitted(t *testing.T) {
	for _, effort := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
		root := t.TempDir()
		writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", effort))
		sess := withCodexSession(t, codexSessionScript("clean"))

		if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
			"prompt": "review this",
			"cwd":    root,
		}); err != nil {
			t.Fatalf("effort %s: audit rpc: %v", effort, err)
		}

		turn := sentParams(t, sess.sent, 2)
		if got, _ := turn["effort"].(string); got != effort {
			t.Errorf("config-key effort %s: turn/start effort = %q, want the pinned value verbatim", effort, got)
		}
	}
}

// TestCodexAuditPin_ConfigKeyEffortClosedSetValidation pins the closed-set
// rejection on the config-key path: a hand-edited workflow.yaml pin effort
// outside the closed set is an ERROR at the resolver — the audit never
// transmits an assumed effort (the unparseable-pin contract). A padded
// member is rejected too: the resolver trims model inputs, not effort
// tokens.
func TestCodexAuditPin_ConfigKeyEffortClosedSetValidation(t *testing.T) {
	for _, bad := range []string{"ultra", "Maximum", "maximal"} {
		root := t.TempDir()
		writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-6.1-sol", bad))
		_, err := resolveCodexAuditModelEffort(map[string]any{"cwd": root})
		if err == nil {
			t.Errorf("pin effort %q: no error, want a closed-set rejection", bad)
			continue
		}
		if !strings.Contains(err.Error(), "closed set") {
			t.Errorf("pin effort %q: error %q does not name the closed set", bad, err)
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

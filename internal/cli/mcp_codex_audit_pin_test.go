package cli

import (
	"context"
	"os"
	"path/filepath"
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

// writeLLMYAML writes an llm.yaml agent-overrides SSOT cell for the codex audit
// agent key INTO the given root (unlike writeCodexLLMFixture, which creates its
// own temp dir — here the llm cell and the workflow pin must share one tree).
func writeLLMYAML(t *testing.T, root, model, effort string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "llm:\n  agent_overrides:\n    " + codexAuditAgentKey + ":\n      model: " + model + "\n      effort: " + effort + "\n"
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// AC-AMP-002 positive arm — a populated audit.codex pin reaches BOTH
// transmission seams on the codex AUDIT path: thread/start carries the pinned
// model (the session-level destination), turn/start carries model + effort.
func TestCodexAuditPin_ReachesTransmittedParams(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-5.6-sol", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got != "gpt-5.6-sol" {
		t.Errorf("thread/start model = %q, want the pinned %q", got, "gpt-5.6-sol")
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "gpt-5.6-sol" {
		t.Errorf("turn/start model = %q, want the pinned %q", got, "gpt-5.6-sol")
	}
	if got, _ := turn["effort"].(string); got != "high" {
		t.Errorf("turn/start effort = %q, want the pinned %q", got, "high")
	}
}

// AC-AMP-002 MF2 negative arm (REQ-AMP-008) — under the SAME populated pin, a
// codex_task turn carries NO pinned model/effort: the task path keeps the
// legacy SSOT-only resolution. This is the regression test for the shared-seam
// leak path codex_task.go → openCodexSessionOn → resolveCodexModelEffort.
func TestCodexAuditPin_TaskTurnCarriesNoPin(t *testing.T) {
	root := t.TempDir()
	// The pin IS resolvable in this tree (workflow.yaml present) AND an llm
	// SSOT cell is present too — so the task path's legacy resolution would
	// transmit the SSOT pair, proving it read llm.yaml and NOT the pin.
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-5.6-sol", "high"))
	writeLLMYAML(t, root, "gpt-5-codex", "medium")
	withCodexProjectDir(t, root)
	prev := codexLookPath
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	t.Cleanup(func() { codexLookPath = prev })
	sess := withCodexSession(t, codexTaskScript("trn-1", "task output"))

	callCodexTask(t, map[string]any{"prompt": "do the task"})

	thread := sentParams(t, sess.sent, 1)
	if got, _ := thread["model"].(string); got == "gpt-5.6-sol" {
		t.Error("codex_task thread/start must NOT carry the audit pin — the pin is audit-entry-only (REQ-AMP-008)")
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "gpt-5-codex" {
		t.Errorf("codex_task turn/start model = %q, want the legacy SSOT %q (not the pin)", got, "gpt-5-codex")
	}
	if got, _ := turn["effort"].(string); got != "medium" {
		t.Errorf("codex_task turn/start effort = %q, want the legacy SSOT %q (not the pin)", got, "medium")
	}
}

// AC-AMP-003 state (c) — an unservable pinned model (a Claude id) falls back
// through the SSOT path; the servability filter holds for the pin exactly as
// it does for the SSOT cell.
func TestCodexAuditPin_UnservableFallsBackToSSOT(t *testing.T) {
	root := writeCodexLLMFixture(t, "gpt-5-codex", "medium")
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("opus", "high"))
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != "gpt-5-codex" {
		t.Errorf("thread/start model = %q, want the SSOT fallback %q (unservable pin must not break the gate)", got, "gpt-5-codex")
	}
	turn := sentParams(t, sess.sent, 2)
	if got, _ := turn["model"].(string); got != "gpt-5-codex" {
		t.Errorf("turn/start model = %q, want the SSOT fallback %q", got, "gpt-5-codex")
	}
	if got, _ := turn["effort"].(string); got != "medium" {
		t.Errorf("turn/start effort = %q, want the SSOT fallback %q (the unservable pin's paired effort drops with it)", got, "medium")
	}
}

// Edge case (§D.2) — effort set with an EMPTY model is no pin (the model is
// the gate): resolution falls through to the legacy SSOT path unchanged.
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

	if got, _ := sentParams(t, sess.sent, 1)["model"].(string); got != "gpt-5-codex" {
		t.Errorf("thread/start model = %q, want the SSOT %q (empty model = no pin)", got, "gpt-5-codex")
	}
	if got, _ := sentParams(t, sess.sent, 2)["effort"].(string); got != "medium" {
		t.Errorf("turn/start effort = %q, want the SSOT %q (pin effort is inert without a model)", got, "medium")
	}
}

// M2 — an explicit caller `model` argument outranks the pin, mirroring the
// existing resolveCodexModelEffort precedence rule.
func TestCodexAuditPin_ExplicitModelOverridesPin(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditCodexPinYAML("gpt-5.6-sol", "high"))
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

// AC-AMP-003 states (a)/(b) — no pin sub-keys / absent workflow.yaml: the
// audit path resolves byte-identically to the pre-SPEC legacy resolution
// (C7 anchor; the deep legacy behavior is pinned by the unmodified
// TestCodexSession_* suite, this asserts the audit-scoped entry equivalence).
func TestCodexAuditPin_AbsentPinEqualsLegacyResolution(t *testing.T) {
	root := t.TempDir() // no workflow.yaml, no llm.yaml → both resolvers drop the pair
	sess := withCodexSession(t, codexSessionScript("clean"))

	if _, err := runCodexAuditReviewRPC(context.Background(), "/fake/codex", codexMethodTurnStart, map[string]any{
		"prompt": "review this",
		"cwd":    root,
	}); err != nil {
		t.Fatalf("audit rpc: %v", err)
	}

	thread := sentParams(t, sess.sent, 1)
	if _, ok := thread["model"]; ok {
		t.Error("thread/start must omit model when no pin and no SSOT cell resolve (byte-identical legacy shape)")
	}
	turn := sentParams(t, sess.sent, 2)
	if _, ok := turn["model"]; ok {
		t.Error("turn/start must omit model when no pin and no SSOT cell resolve")
	}
	if _, ok := turn["effort"]; ok {
		t.Error("turn/start must omit effort when no pin and no SSOT cell resolve")
	}
}

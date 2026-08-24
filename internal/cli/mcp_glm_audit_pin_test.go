package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/template"
)

// SPEC-V3R6-AUDIT-MODEL-PIN-001 M3 — the GLM audit pin (workflow.audit.glm):
// model + effort resolution and WIRE delivery on the glmMessagesRequest body,
// under the REQ-AMP-006 single-reading effort rule. The stubGLMDoer body
// capture + newGLMReviewTree harness from SPEC-MOAI-MCP-SERVER-001 M3 is
// reused verbatim (AP-1).

// auditGLMPinYAML is a workflow.yaml body carrying only the GLM audit pin.
func auditGLMPinYAML(model, effort string) string {
	return "workflow:\n  audit:\n    glm:\n      model: " + model + "\n      effort: " + effort + "\n"
}

// glmGLMSessionLLMYAML is an llm.yaml marking the session GLM-backed and
// pinning the task-family agent (super-advisor) to a NON-pin model, so the
// task-path isolation test can distinguish pin reads from SSOT reads.
func glmGLMSessionLLMYAML() string {
	return "llm:\n  team_mode: glm\n  agent_overrides:\n    " + glmTaskAgentKey + ":\n      model: glm-4.6\n      effort: low\n"
}

// writeRawSectionYAML writes a raw body as a section file under root.
func writeRawSectionYAML(t *testing.T, root, name, body string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// decodeGLMRequestBody unmarshals the captured outbound request body.
func decodeGLMRequestBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("request body not JSON: %v\nbody=%s", err, body)
	}
	return m
}

// thinkingDirective returns the request's thinking object, if present.
func thinkingDirective(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	raw, ok := body["thinking"]
	if !ok || raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("thinking field is %T, want an object", raw)
	}
	return m
}

// AC-AMP-004 positive arm — a populated audit.glm pin reaches the outbound
// request body: model carries the pinned id and the reasoning directive is set
// to max on the delivery field (hypothesis A: the Anthropic-style thinking
// object, the field the M5 live gate arbitrates).
func TestGLMAuditPin_ReachesRequestBody(t *testing.T) {
	root := newGLMReviewTree(t, true)
	writeCodexWorkflowYAML(t, root, auditGLMPinYAML("glm-5.3", template.GLMStateMax))
	withCodexProjectDir(t, root)
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "test-key", stub)

	if _, err := handleGLMAudit(context.Background(), glmAuditReq(root)); err != nil {
		t.Fatalf("handleGLMAudit: %v", err)
	}

	body := decodeGLMRequestBody(t, stub.gotBody)
	if got, _ := body["model"].(string); got != "glm-5.3" {
		t.Errorf("request model = %q, want the pinned %q", got, "glm-5.3")
	}
	th := thinkingDirective(t, body)
	if th == nil {
		t.Fatal("request carries no thinking directive for a valid pinned effort (max)")
	}
	if got, _ := th["type"].(string); got != template.GLMThinkingEnabledVal {
		t.Errorf("thinking.type = %q, want %q", got, template.GLMThinkingEnabledVal)
	}
	if budget, _ := th["budget_tokens"].(float64); int(budget) != glmAuditThinkingBudgetMax {
		t.Errorf("thinking.budget_tokens = %v, want %d (the max-state budget)", budget, glmAuditThinkingBudgetMax)
	}
}

// AC-AMP-004 second arm (REQ-AMP-006 single-reading rule) — a Claude-only
// vocabulary value (`medium`) is INVALID on this path: the reasoning directive
// is omitted while the model pin still applies.
func TestGLMAuditPin_InvalidEffortOmitsReasoningDirective(t *testing.T) {
	root := newGLMReviewTree(t, true)
	writeCodexWorkflowYAML(t, root, auditGLMPinYAML("glm-5.3", "medium"))
	withCodexProjectDir(t, root)
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "test-key", stub)

	if _, err := handleGLMAudit(context.Background(), glmAuditReq(root)); err != nil {
		t.Fatalf("handleGLMAudit: %v", err)
	}

	body := decodeGLMRequestBody(t, stub.gotBody)
	if got, _ := body["model"].(string); got != "glm-5.3" {
		t.Errorf("request model = %q, want the pinned %q (the model pin survives an invalid effort)", got, "glm-5.3")
	}
	if th := thinkingDirective(t, body); th != nil {
		t.Errorf("thinking directive must be omitted for invalid effort medium, got %v", th)
	}
}

// AC-AMP-004 third arm — no pin sub-keys: the body contains no reasoning field
// and the model resolves exactly as before (legacy SSOT → GLM default when no
// llm.yaml exists).
func TestGLMAuditPin_AbsentPinBodyUnchanged(t *testing.T) {
	root := newGLMReviewTree(t, true) // no workflow.yaml → no pin
	withCodexProjectDir(t, root)
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "test-key", stub)

	if _, err := handleGLMAudit(context.Background(), glmAuditReq(root)); err != nil {
		t.Fatalf("handleGLMAudit: %v", err)
	}

	body := decodeGLMRequestBody(t, stub.gotBody)
	if th := thinkingDirective(t, body); th != nil {
		t.Errorf("no thinking directive may appear without a pin, got %v", th)
	}
	if got, _ := body["model"].(string); got != glmAuditDefaultModel {
		t.Errorf("request model = %q, want the legacy default %q (byte-identical legacy shape)", got, glmAuditDefaultModel)
	}
	if strings.Contains(stub.gotBody, "reasoning") {
		t.Errorf("absent pin must leave the body free of reasoning fields:\n%s", stub.gotBody)
	}
}

// REQ-AMP-008 — glm_task resolution ignores a populated audit.glm pin: the
// task family resolves through its llm SSOT cell (glm-4.6 on super-advisor),
// not the pin (glm-5.3), under the SAME config.
func TestGLMAuditPin_TaskResolutionUnaffected(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditGLMPinYAML("glm-5.3", template.GLMStateMax))
	writeRawSectionYAML(t, root, "llm.yaml", glmGLMSessionLLMYAML())
	withCodexProjectDir(t, root)

	if got := resolveGLMTaskModel(); got != "glm-4.6" {
		t.Errorf("resolveGLMTaskModel = %q, want the SSOT cell glm-4.6 (the audit pin must not leak into glm_task)", got)
	}

	// The pin DOES apply on the audit resolver under the same config.
	me := resolveGLMAuditModelEffort()
	if me.Model != "glm-5.3" || me.Effort != template.GLMStateMax {
		t.Errorf("resolveGLMAuditModelEffort = %+v, want {glm-5.3 max} (the pin outranks the SSOT cell on the audit path)", me)
	}
}

// M3 — the pin bypasses the IsGLMBackend session check: a pin is
// by-construction a GLM id, and a wrong id degrades via the existing z.ai-4xx
// fail-open (design decision D3). Here: NO llm.yaml, no GLM session marker —
// the pin still resolves (the legacy fallback would have returned the default).
func TestGLMAuditPin_BypassesSessionBackendCheck(t *testing.T) {
	root := t.TempDir()
	writeCodexWorkflowYAML(t, root, auditGLMPinYAML("glm-4.6", template.GLMStateLow))
	withCodexProjectDir(t, root)

	me := resolveGLMAuditModelEffort()
	if me.Model != "glm-4.6" || me.Effort != template.GLMStateLow {
		t.Errorf("resolveGLMAuditModelEffort = %+v, want {glm-4.6 low} (a pin resolves without a GLM session marker)", me)
	}
}

// M3 — an explicit caller `model` argument outranks the pinned model, mirroring
// the codex resolver's precedence; the pinned effort stays.
func TestGLMAuditPin_ExplicitModelParamOverridesPin(t *testing.T) {
	root := newGLMReviewTree(t, true)
	writeCodexWorkflowYAML(t, root, auditGLMPinYAML("glm-5.3", template.GLMStateMax))
	withCodexProjectDir(t, root)
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "test-key", stub)

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		projectRootArg: root,
		"model":        "glm-4.6",
	}}}
	if _, err := handleGLMAudit(context.Background(), req); err != nil {
		t.Fatalf("handleGLMAudit: %v", err)
	}

	body := decodeGLMRequestBody(t, stub.gotBody)
	if got, _ := body["model"].(string); got != "glm-4.6" {
		t.Errorf("request model = %q, want the explicit override %q", got, "glm-4.6")
	}
	if th := thinkingDirective(t, body); th == nil {
		t.Error("the pinned effort (max) must still deliver the reasoning directive under an explicit model override")
	}
}

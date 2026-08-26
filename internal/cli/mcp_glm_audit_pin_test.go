package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

// parseGLMReview regression (live-gate finding, AC-AMP-006 first run): when
// the reasoning directive is delivered, z.ai returns a LEADING thinking
// content block before the text block. The pre-SPEC parser read Content[0]
// blindly — for a thinking block (whose payload lives in `thinking`, not
// `text`) it saw an empty Text and failed open to inconclusive even though a
// full review sat in the next block. The parser must select the first TEXT
// block, not the first block.
func TestGLMAuditParse_SkipsLeadingThinkingBlock(t *testing.T) {
	review := ReviewOutput{Verdict: "pass", Summary: "clean"}
	text, err := json.Marshal(review)
	if err != nil {
		t.Fatal(err)
	}
	envelope := map[string]any{
		"content": []map[string]any{
			{"type": "thinking", "thinking": "Let me analyze this diff carefully..."},
			{"type": "text", "text": string(text)},
		},
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}

	out := parseGLMReview(raw)
	if out.Verdict != "pass" {
		t.Errorf("parseGLMReview verdict = %q (summary: %q) — a leading thinking block must not eclipse the text payload", out.Verdict, out.Summary)
	}
}

// TestGLMAuditPin_RequestBody is the AC-AMP-004 table: every pin state's
// outbound request body, seam-captured through stubGLMDoer. The delivery field
// is the one SELECTED BY THE LIVE EVIDENCE (AC-AMP-006: hypothesis A's
// thinking budget measured a true null, 1.02; hypothesis B — the top-level
// z.ai reasoning_effort, carried verbatim — is the honored field).
func TestGLMAuditPin_RequestBody(t *testing.T) {
	cases := []struct {
		name string
		// pinYAML: the workflow.yaml audit.glm body ("" = no file at all).
		pinYAML string
		// wantModel: the model the body must carry.
		wantModel string
		// wantEffort: the reasoning_effort the body must carry ("" = the
		// field must be ABSENT).
		wantEffort string
	}{
		{
			name:       "valid max pin reaches both fields",
			pinYAML:    auditGLMPinYAML("glm-5.3", template.GLMStateMax),
			wantModel:  "glm-5.3",
			wantEffort: template.GLMStateMax,
		},
		{
			name:       "valid high pin transmitted verbatim",
			pinYAML:    auditGLMPinYAML("glm-5.3", template.GLMStateHigh),
			wantModel:  "glm-5.3",
			wantEffort: template.GLMStateHigh,
		},
		{
			name:       "invalid effort (Claude-only medium) omits directive, model pin survives",
			pinYAML:    auditGLMPinYAML("glm-5.3", "medium"),
			wantModel:  "glm-5.3",
			wantEffort: "",
		},
		{
			name:       "populated model + empty effort omits directive, model still pinned",
			pinYAML:    auditGLMPinYAML("glm-5.3", ""),
			wantModel:  "glm-5.3",
			wantEffort: "",
		},
		{
			name:       "absent pin leaves body unchanged (legacy default model, no reasoning)",
			pinYAML:    "",
			wantModel:  glmAuditDefaultModel,
			wantEffort: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newGLMReviewTree(t, true)
			if tc.pinYAML != "" {
				writeCodexWorkflowYAML(t, root, tc.pinYAML)
			}
			withCodexProjectDir(t, root)
			stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
			withGLMSeams(t, "test-key", stub)

			if _, err := handleGLMAudit(context.Background(), glmAuditReq(root)); err != nil {
				t.Fatalf("handleGLMAudit: %v", err)
			}

			body := decodeGLMRequestBody(t, stub.gotBody)
			if got, _ := body["model"].(string); got != tc.wantModel {
				t.Errorf("request model = %q, want %q", got, tc.wantModel)
			}
			if tc.wantEffort == "" {
				if _, ok := body["reasoning_effort"]; ok {
					t.Errorf("reasoning_effort must be ABSENT in this state; body: %s", stub.gotBody)
				}
			} else if got, _ := body["reasoning_effort"].(string); got != tc.wantEffort {
				t.Errorf("reasoning_effort = %q, want the pinned state %q transmitted verbatim", got, tc.wantEffort)
			}
			// Hypothesis A's field is retired: the live gate measured
			// budget_tokens ignored, so a stale thinking object must not ride
			// the request in ANY state.
			if _, ok := body["thinking"]; ok {
				t.Error("request still carries a thinking object — hypothesis A was rejected by the live differential (budget_tokens ignored)")
			}
		})
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

	// The pin DOES apply on the audit resolver under the same config. CR #8:
	// the caller names the reviewed tree explicitly; the same root resolves
	// the pin the audit reads.
	me := resolveGLMAuditModelEffort(root)
	if me.Model != "glm-5.3" || me.Effort != template.GLMStateMax {
		t.Errorf("resolveGLMAuditModelEffort(root) = %+v, want {glm-5.3 max} (the pin outranks the SSOT cell on the audit path)", me)
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

	me := resolveGLMAuditModelEffort(root)
	if me.Model != "glm-4.6" || me.Effort != template.GLMStateLow {
		t.Errorf("resolveGLMAuditModelEffort(root) = %+v, want {glm-4.6 low} (a pin resolves without a GLM session marker)", me)
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
	if got, _ := body["reasoning_effort"].(string); got != template.GLMStateMax {
		t.Errorf("reasoning_effort = %q, want the pinned %q (the pinned effort survives an explicit model override)", got, template.GLMStateMax)
	}
}

// codex_glm_auditor_model_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-020): codex and GLM audit verdicts record the auditor model, exactly
// as the Claude verdict already does on its provenance.
package cli

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// callToolGLMAudit invokes the glm_audit handler directly.
func callToolGLMAudit(t *testing.T, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "glm_audit"
	req.Params.Arguments = args
	res, err := handleGLMAudit(context.Background(), req)
	if err != nil {
		t.Fatalf("handleGLMAudit returned a hard error: %v", err)
	}
	return res
}

// TestCodexAudit_RecordsAuditorModel pins the model record: a codex_audit
// result carries provenance naming the resolved auditor model.
func TestCodexAudit_RecordsAuditorModel(t *testing.T) {
	root := newProbeProject(t, "SPEC-AUDMODEL-921")
	withCodexSession(t, codexSessionScript(realCleanReview))

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	prov, _ := got["provenance"].(map[string]any)
	if prov == nil {
		t.Fatal("codex_audit carried no provenance — the auditor model is unrecorded (REQ-MN-020)")
	}
	if rm, _ := prov["resolved_model"].(string); rm == "" {
		t.Errorf("provenance.resolved_model is empty — the auditor model must be recorded (REQ-MN-020)")
	}
	if b, _ := prov["backend"].(string); b != "codex" {
		t.Errorf("provenance.backend = %q, want codex", b)
	}
}

// TestGLMAudit_RecordsAuditorModel pins the same record on the GLM leg.
func TestGLMAudit_RecordsAuditorModel(t *testing.T) {
	root := newGLMReviewTree(t, true)
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Findings: []Finding{}, NextSteps: []string{}})}
	withGLMSeams(t, "stub-key", stub)

	got := structuredMap(t, callToolGLMAudit(t, map[string]any{"project_root": root}))
	prov, _ := got["provenance"].(map[string]any)
	if prov == nil {
		t.Fatal("glm_audit carried no provenance — the auditor model is unrecorded (REQ-MN-020)")
	}
	if rm, _ := prov["resolved_model"].(string); rm == "" {
		t.Errorf("provenance.resolved_model is empty — the auditor model must be recorded (REQ-MN-020)")
	}
	if b, _ := prov["backend"].(string); b != "glm" {
		t.Errorf("provenance.backend = %q, want glm", b)
	}
}

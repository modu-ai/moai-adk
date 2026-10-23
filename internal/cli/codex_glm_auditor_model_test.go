// codex_glm_auditor_model_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-020): codex and GLM audit verdicts record the auditor model, exactly
// as the Claude verdict already does on its provenance.
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

// TestCodexAudit_ProvenanceModelIsTheTransmittedModel pins the audit pin's
// single read (card t1632 audit N2 / debt D1): the auditor model the verdict's
// provenance records is the model the request actually carried. The pin is
// rewritten after handleCodexAudit's own pre-call read but before the session
// resolves the model for the wire (the containment listing runs between the
// two), so a second, independent config read would transmit one model while
// the provenance names another.
func TestCodexAudit_ProvenanceModelIsTheTransmittedModel(t *testing.T) {
	root := newProbeProject(t, "SPEC-AUDMODEL-922")
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	pinFile := filepath.Join(sections, "workflow.yaml")
	writePin := func(model, effort string) {
		t.Helper()
		body := "workflow:\n    audit:\n        codex:\n            model: " + model + "\n            effort: " + effort + "\n"
		if err := os.WriteFile(pinFile, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writePin("gpt-6.1-sol", "high")

	sess := withCodexSession(t, codexSessionScript(realCleanReview))
	prevList := codexAuditMCPList
	rewritten := false
	codexAuditMCPList = func(ctx context.Context, program, r string) (string, error) {
		if !rewritten {
			rewritten = true
			writePin("gpt-5-codex", "low") // the pin edited mid-handshake
		}
		return prevList(ctx, program, r)
	}
	t.Cleanup(func() { codexAuditMCPList = prevList })

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	if !rewritten {
		t.Fatal("the pin rewrite hook never ran — the test measured nothing")
	}
	prov, _ := got["provenance"].(map[string]any)
	if prov == nil {
		t.Fatal("codex_audit carried no provenance")
	}
	recorded, _ := prov["resolved_model"].(string)

	sent := ""
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil || m["method"] != codexMethodThreadStart {
			continue
		}
		params, _ := m["params"].(map[string]any)
		sent, _ = params["model"].(string)
	}
	if sent == "" {
		t.Fatalf("no thread/start model was transmitted; sent=%v", sess.sent)
	}
	if sent != recorded {
		t.Errorf("transmitted model %q != provenance.resolved_model %q — two independent pin reads", sent, recorded)
	}
}

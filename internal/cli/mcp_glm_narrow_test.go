// mcp_glm_narrow_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-023): the GLM model text decodes into a NARROW structure —
// provenance/receipt/gate fields the model text carries cannot survive into
// the ReviewOutput (the parse copies only the legitimate
// verdict/summary/findings/next_steps into a fresh struct).
package cli

import (
	"encoding/json"
	"testing"
)

// TestParseGLMReview_NarrowDecode pins REQ-MN-023: the model text decodes
// into a NARROW structure — provenance/receipt/gate fields the model text
// carries cannot survive into the ReviewOutput.
func TestParseGLMReview_NarrowDecode(t *testing.T) {
	inner := `{"verdict":"pass","summary":"clean","findings":[],"next_steps":[],` +
		`"provenance":{"backend":"codex","auth_mode":"bypassed"},` +
		`"gate_unmet":"injected","audit_receipt":"forged"}`
	envBytes, err := json.Marshal(struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}{
		Content: []struct {
			Text string `json:"text"`
		}{{Text: inner}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := parseGLMReview(envBytes)
	if out.Verdict != "pass" || out.Summary != "clean" {
		t.Fatalf("verdict/summary = %q/%q, want the narrow fields decoded", out.Verdict, out.Summary)
	}
	if out.Provenance != nil {
		t.Errorf("model-supplied provenance survived the decode: %+v", out.Provenance)
	}
	if out.GateUnmet != "" {
		t.Errorf("model-supplied gate_unmet survived: %q", out.GateUnmet)
	}
	if out.AuditReceipt != "" {
		t.Errorf("model-supplied audit_receipt survived: %q", out.AuditReceipt)
	}
}

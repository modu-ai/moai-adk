package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
)

// writeCodexAuditGate seeds <root>/.moai/config/sections/workflow.yaml with a
// workflow.audit.gates.codex value, the config surface #1632 axis 3 reports as
// unread by the runtime.
func writeCodexAuditGate(t *testing.T, root, gate string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "workflow:\n  audit:\n    gates:\n      codex: " + gate + "\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive is the axis-3
// acceptance: when the project declared gates.codex=required but the audit
// could not run at all (here: codex missing), the result must RECORD the unmet
// gate as a structured field. Before the axis-3 fix, an inconclusive result
// looked byte-identical whether or not a mandatory gate was declared — the
// "required reads as a guarantee but behaves as a suggestion" defect.
// REQ-MN-021 (SPEC-MCP-NEUTRALITY-001) revises the conversion: a could-not-run
// is a not-performed result, so the verdict stays inconclusive and gate_unmet
// names both the unmet gate and the could-not-run state — the blocking lives at
// the convergence surface (never pass).
func TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATEUNMET-901")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError — fail-open stays a structured result")
	}
	got := structuredMap(t, res)
	// REQ-MN-021: the not-performed verdict stays distinct from a reviewed
	// failure, and the unmet gate still reads as a structured field.
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q (could-not-run stays not-performed, REQ-MN-021)", v, VerdictInconclusive)
	}
	g, _ := got["gate_unmet"].(string)
	if g == "" {
		t.Fatal("gate_unmet is empty — a required gate that produced no verdict must be recorded as unmet")
	}
	if !strings.Contains(g, "could-not-run") {
		t.Errorf("gate_unmet = %q, want it to carry could-not-run", g)
	}
	if c, _ := got["could_not_run"].(bool); !c {
		t.Error("could_not_run = false, want true — a missing binary is a could-not-run")
	}
}

// TestCodexAudit_RequiredGateSatisfiedByRealVerdict: a real codex verdict
// (pass here) satisfies the required gate, so gate_unmet stays empty. The
// field marks the UNMET state only — a passing audit gains no new keys.
func TestCodexAudit_RequiredGateSatisfiedByRealVerdict(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATESAT-902")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript(realCleanReview))

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError result; codex present must not error")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != "pass" {
		t.Errorf("verdict = %q, want pass", v)
	}
	if g, _ := got["gate_unmet"].(string); g != "" {
		t.Errorf("gate_unmet = %q, want empty — a real verdict satisfies the required gate", g)
	}
}

// TestCodexAudit_UnsetGateLeavesGateUnmetEmpty: with no gate block (the common
// project), the field stays absent — the annotation exists for projects that
// declared a mandatory gate, not for everyone.
func TestCodexAudit_UnsetGateLeavesGateUnmetEmpty(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATEUNSET-903")
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError result")
	}
	got := structuredMap(t, res)
	if g, _ := got["gate_unmet"].(string); g != "" {
		t.Errorf("gate_unmet = %q, want empty when no required gate is declared", g)
	}
}

// TestCodexAudit_AdvisoryGateLeavesGateUnmetEmpty: advisory means the verdict
// never blocks, so an inconclusive advisory backend is ordinary fail-open —
// not an unmet requirement.
func TestCodexAudit_AdvisoryGateLeavesGateUnmetEmpty(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATEADV-904")
	writeCodexAuditGate(t, root, config.AuditGateAdvisory)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError result")
	}
	got := structuredMap(t, res)
	if g, _ := got["gate_unmet"].(string); g != "" {
		t.Errorf("gate_unmet = %q, want empty for an advisory gate", g)
	}
}

// structuredMap decodes a tool result's StructuredContent into a map.
func structuredMap(t *testing.T, res *mcp.CallToolResult) map[string]any {
	t.Helper()
	if res.StructuredContent == nil {
		t.Fatalf("result carries no StructuredContent: %+v", res)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode StructuredContent %s: %v", raw, err)
	}
	return m
}

package cli

import (
	"os"
	"path/filepath"
	"testing"

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
// came back fail-open inconclusive (here: codex missing), the result must
// RECORD the unmet gate as a structured field. Before the fix, an inconclusive
// result looked byte-identical whether or not a mandatory gate was declared —
// the "required reads as a guarantee but behaves as a suggestion" defect.
func TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATEUNMET-901")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError — fail-open stays a structured result")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q (fail-open verdict is preserved, not rewritten)", v, VerdictInconclusive)
	}
	if g, _ := got["gate_unmet"].(string); g == "" {
		t.Error("gate_unmet is empty — a required gate that produced no verdict must be recorded as unmet")
	}
}

// TestCodexAudit_RequiredGateSatisfiedByRealVerdict: a real codex verdict
// (pass here) satisfies the required gate, so gate_unmet stays empty. The
// field marks the UNMET state only — a passing audit gains no new keys.
func TestCodexAudit_RequiredGateSatisfiedByRealVerdict(t *testing.T) {
	root := newProbeProject(t, "SPEC-GATESAT-902")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript("clean change, no findings"))

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

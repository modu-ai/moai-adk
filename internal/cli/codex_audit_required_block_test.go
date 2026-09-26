package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// assertRequiredGateBlock checks the blocking result an explicitly required
// codex gate yields when the audit produced no verdict: a structured result
// (isError false) whose verdict is fail, whose gate_unmet is set, and whose
// summary names both the unmet gate and the original no-verdict cause.
func assertRequiredGateBlock(t *testing.T, root, wantCause string) {
	t.Helper()
	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("isError = true — the block is a structured verdict, not a tool error")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != "fail" {
		t.Errorf("verdict = %q, want fail — an explicitly required gate left without a verdict must block", v)
	}
	if g, _ := got["gate_unmet"].(string); g == "" {
		t.Error("gate_unmet is empty — the block must say it is an unmet gate, not a reviewed failure")
	}
	summary, _ := got["summary"].(string)
	if !strings.Contains(summary, "required") {
		t.Errorf("summary = %q, want it to name the unmet required gate", summary)
	}
	if !strings.Contains(summary, wantCause) {
		t.Errorf("summary = %q, want it to keep the original no-verdict cause %q", summary, wantCause)
	}
}

// AC-CAG-001: required + codex binary absent blocks.
func TestCodexAudit_RequiredGateBlocksWhenBinaryAbsent(t *testing.T) {
	root := newProbeProject(t, "SPEC-CAGBLOCK-001")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	assertRequiredGateBlock(t, root, "codex binary not found")
}

// AC-CAG-002 (i): required + RPC failure (fail-open inconclusive) blocks.
func TestCodexAudit_RequiredGateBlocksOnRPCFailure(t *testing.T) {
	root := newProbeProject(t, "SPEC-CAGBLOCK-002")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	prevLook, prevRPC := codexLookPath, codexReviewRPC
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexReviewRPC = func(context.Context, string, string, map[string]any) (ReviewOutput, error) {
		return inconclusiveReview("rpc transport closed"), nil
	}
	t.Cleanup(func() { codexLookPath, codexReviewRPC = prevLook, prevRPC })

	assertRequiredGateBlock(t, root, "rpc transport closed")
}

// AC-CAG-002 (ii): required + blank review output blocks.
func TestCodexAudit_RequiredGateBlocksOnBlankOutput(t *testing.T) {
	root := newProbeProject(t, "SPEC-CAGBLOCK-003")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript("   \n  "))

	assertRequiredGateBlock(t, root, "blank")
}

// AC-CAG-004: the distributed engine default (codex required when the key is
// absent) is not an explicit opt-in.
func TestCodexAudit_EngineDefaultRequiredIsNotOptIn(t *testing.T) {
	root := newProbeProject(t, "SPEC-CAGDEFAULT-004")
	cfg := workflowAuditPins(root)
	if got := gateOr(cfg.Gates.Codex, config.AuditGateRequired); got != config.AuditGateRequired {
		t.Fatalf("premise: engine-defaulted codex gate = %q, want required", got)
	}
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want inconclusive — the default is not an opt-in", v)
	}
	if _, present := got["gate_unmet"]; present {
		t.Errorf("gate_unmet present for a tree with no audit block: %v", got["gate_unmet"])
	}
}

// AC-CAG-005: a real pass or fail verdict is left untouched under required.
func TestCodexAudit_RequiredGateLeavesRealVerdicts(t *testing.T) {
	for name, tc := range map[string]struct{ body, want string }{
		"pass": {realCleanReview, "pass"},
		"fail": {issue1632ReviewBody, "fail"},
	} {
		t.Run(name, func(t *testing.T) {
			root := newProbeProject(t, "SPEC-CAGREAL-005")
			writeCodexAuditGate(t, root, config.AuditGateRequired)
			withCodexSession(t, codexSessionScript(tc.body))

			got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
			if v, _ := got["verdict"].(string); v != tc.want {
				t.Errorf("verdict = %q, want %q", v, tc.want)
			}
			if g, _ := got["gate_unmet"].(string); g != "" {
				t.Errorf("gate_unmet = %q, want empty for a real verdict", g)
			}
		})
	}
}

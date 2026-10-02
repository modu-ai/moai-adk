// Package cli — the enforcement follows the plan resolved at call start
// (SPEC-AUDIT-MODEL-CONVERGE-001 M8, sync-audit findings F1 and F2; REQ-ACV-006
// and REQ-ACV-008).
//
// One plan, resolved once with the gates the call supplied, drives both the
// fan-out and the unmet-gate enforcement: a supplied gate that is not
// `required` overrides a configured `required` at enforcement as well as at the
// fan-out, and an edit to workflow.yaml made while the backends run changes
// neither.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// perBackendGate returns the gate the result's entry for backend carries.
func perBackendGate(t *testing.T, m map[string]any, backend string) string {
	t.Helper()
	list, _ := m["per_backend_verdicts"].([]any)
	for _, e := range list {
		entry, _ := e.(map[string]any)
		if entry["backend"] == backend {
			g, _ := entry["gate"].(string)
			return g
		}
	}
	t.Fatalf("no per_backend_verdicts entry for %s in %v", backend, m["per_backend_verdicts"])
	return ""
}

// TestAuditMulti_SuppliedGateOverridesConfiguredRequiredAtEnforcement (F1,
// REQ-ACV-006): with `audit.model: multi` a call that supplies codex `advisory`
// is a call whose codex gate IS advisory — at the entry and at enforcement. An
// unanswered codex is then not an unmet required gate.
func TestAuditMulti_SuppliedGateOverridesConfiguredRequiredAtEnforcement(t *testing.T) {
	codexDown := map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")}

	t.Run("advisory over a configured required", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		_, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "advisory"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		if g := perBackendGate(t, m, BackendCodex); g != config.AuditGateAdvisory {
			t.Errorf("codex entry gate = %q, want %q", g, config.AuditGateAdvisory)
		}
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent: the entry says advisory, so enforcement must agree", m["gate_unmet"])
		}
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass (an advisory gate with no verdict is fail-open)", m["overall_verdict"])
		}
	})

	t.Run("off over a configured required", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		rc, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "off"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		requireBackends(t, "codex off", invoked(rc), []string{BackendClaude, BackendGLM})
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
		}
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass", m["overall_verdict"])
		}
	})

	t.Run("advisory over an audit.gates required", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML("", map[string]string{"codex": config.AuditGateRequired}))
		_, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "advisory"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
		}
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass", m["overall_verdict"])
		}
	})

	// The neighbours stay as they are: a supplied `required` is not itself an
	// opt-in, but it does not cancel a configured one either.
	t.Run("required over a configured required still fails closed", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		_, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "required"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		if m["overall_verdict"] != "fail" || m["gate_unmet"] != BackendCodex {
			t.Errorf("overall_verdict = %v, gate_unmet = %v, want fail / codex", m["overall_verdict"], m["gate_unmet"])
		}
	})
	t.Run("required over an unconfigured tree stays fail-open", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, distributedAuditPinsYAML)
		_, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "required"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass (an argument-sourced required is not explicit)", m["overall_verdict"])
		}
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
		}
	})
	t.Run("a different backend's supplied gate leaves the configured codex gate enforced", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		_, res, m := planCall(t, root, map[string]any{"gates": map[string]any{"glm": "off"}}, codexDown)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		if m["overall_verdict"] != "fail" || m["gate_unmet"] != BackendCodex {
			t.Errorf("overall_verdict = %v, gate_unmet = %v, want fail / codex", m["overall_verdict"], m["gate_unmet"])
		}
	})
}

// callAuditMultiRewritingConfig runs one audit_multi call whose codex leg
// rewrites the tree's workflow.yaml to rewritten BEFORE it answers — an edit
// that lands after the plan was resolved and before the enforcement runs — and
// returns the decoded result.
func callAuditMultiRewritingConfig(t *testing.T, root, rewritten string, codexAnswer ReviewOutput) map[string]any {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	path := filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")
	orig := backendCall
	backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		if backend != BackendCodex {
			return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
		}
		if err := os.WriteFile(path, []byte(rewritten), 0o644); err != nil {
			t.Errorf("rewrite workflow.yaml mid-call: %v", err)
		}
		return codexAnswer
	}
	t.Cleanup(func() { backendCall = orig })

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"project_root": root}
	res, err := handleAuditMulti(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAuditMulti returned a Go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(toolResultText(res)), &m); err != nil {
		t.Fatalf("decode audit_multi result: %v\n%s", err, toolResultText(res))
	}
	return m
}

// TestAuditMulti_EnforcementFollowsTheCallStartPlan (F2): the plan is resolved
// once at call start, and an edit of workflow.yaml made while the backends run
// moves neither the fan-out nor the enforcement.
func TestAuditMulti_EnforcementFollowsTheCallStartPlan(t *testing.T) {
	codexDown := inconclusiveVerdict("codex unavailable")

	t.Run("a required gate cannot be edited away mid-call", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		m := callAuditMultiRewritingConfig(t, root, planWorkflowYAML(config.AuditModelClaude, nil), codexDown)
		if m["overall_verdict"] != "fail" || m["gate_unmet"] != BackendCodex {
			t.Errorf("overall_verdict = %v, gate_unmet = %v, want fail / codex (the call-start plan required codex)", m["overall_verdict"], m["gate_unmet"])
		}
	})

	t.Run("a gate cannot be edited in mid-call", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, distributedAuditPinsYAML)
		m := callAuditMultiRewritingConfig(t, root, planWorkflowYAML(config.AuditModelMulti, nil), codexDown)
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass (nothing was configured at call start)", m["overall_verdict"])
		}
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
		}
	})
}

// TestCallStartEnforcement pins the gates the enforcement keys on, per input:
// what the operator wrote, with a supplied off or advisory replacing it, a
// supplied required adding nothing, and the orphaned-root assumption kept.
func TestCallStartEnforcement(t *testing.T) {
	multi := config.AuditConfig{Model: config.AuditModelMulti}
	cases := []struct {
		name         string
		audit        config.AuditConfig
		unidentified bool
		supplied     config.AuditGates
		want         config.AuditGates
		wantNote     bool
	}{
		{"nothing configured", config.AuditConfig{}, false, config.AuditGates{}, config.AuditGates{}, false},
		{"multi token", multi, false, config.AuditGates{},
			config.AuditGates{Claude: "required", Codex: "required", GLM: "advisory"}, false},
		{"supplied advisory replaces required", multi, false, config.AuditGates{Codex: "advisory"},
			config.AuditGates{Claude: "required", Codex: "advisory", GLM: "advisory"}, false},
		{"supplied off replaces required", multi, false, config.AuditGates{Claude: "off"},
			config.AuditGates{Claude: "off", Codex: "required", GLM: "advisory"}, false},
		{"supplied required over a configured required keeps it", multi, false, config.AuditGates{Codex: "required"},
			config.AuditGates{Claude: "required", Codex: "required", GLM: "advisory"}, false},
		{"supplied required over nothing adds nothing", config.AuditConfig{}, false, config.AuditGates{Codex: "required"},
			config.AuditGates{}, false},
		{"unidentified primary assumes codex required", config.AuditConfig{}, true, config.AuditGates{},
			config.AuditGates{Codex: "required"}, true},
		{"supplied gate still wins on an unidentified primary", config.AuditConfig{}, true, config.AuditGates{Codex: "advisory"},
			config.AuditGates{Codex: "advisory"}, true},
		{"a rejected configuration reads as not configured", config.AuditConfig{Model: "grok"}, false, config.AuditGates{},
			config.AuditGates{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, note := callStartEnforcement(tc.audit, tc.unidentified, tc.supplied)
			if got != tc.want {
				t.Errorf("gates = %+v, want %+v", got, tc.want)
			}
			if (note != "") != tc.wantNote {
				t.Errorf("note = %q, want non-empty = %v", note, tc.wantNote)
			}
		})
	}
}

// TestRunMultiAudit_CarrylessCallerKeepsTheReReadFallback: a caller that hands
// runMultiAudit no enforcement gates still gets the post-fan-out re-read of the
// tree's configuration — the behaviour every direct caller had before M8.
func TestRunMultiAudit_CarrylessCallerKeepsTheReReadFallback(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
	rc := &recordingCallerMulti{verdictBy: map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")}}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	cfg := MultiAuditConfig{
		Gates:       config.AuditGates{Claude: "required", Codex: "required", GLM: "advisory"},
		ProjectRoot: root,
	}
	if cfg.EnforcementGates != nil {
		t.Fatal("precondition: a carry-less config has no enforcement gates")
	}
	res := runMultiAudit(context.Background(), ReviewOutput{}, "", "", cfg, nil)
	if res.OverallVerdict != "fail" || res.GateUnmet != BackendCodex {
		t.Errorf("overall = %q, gate_unmet = %q, want fail / codex from the re-read", res.OverallVerdict, res.GateUnmet)
	}
}

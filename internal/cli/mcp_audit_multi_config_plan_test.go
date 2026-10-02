// Package cli — `audit_multi` follows the configured audit plan
// (SPEC-AUDIT-MODEL-CONVERGE-001 M3a, AC-ACV-006/-007/-009/-010 and the
// consumer half of AC-ACV-003).
//
// Every test drives handleAuditMulti over a temp project root with a stubbed
// backendCall seam, so what is asserted is what the handler hands the fan-out
// and what the convergence result reports — never a live backend.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// planWorkflowYAML renders a workflow.yaml whose audit block carries the given
// model token and gate values, each double-quoted so surrounding whitespace in a
// value survives the YAML read. An empty model or an absent gate key is omitted.
func planWorkflowYAML(model string, gates map[string]string) string {
	var b strings.Builder
	b.WriteString("workflow:\n  audit:\n")
	if model != "" {
		fmt.Fprintf(&b, "    model: %q\n", model)
	}
	if len(gates) > 0 {
		b.WriteString("    gates:\n")
		for _, k := range []string{"claude", "codex", "glm"} {
			if v, ok := gates[k]; ok {
				fmt.Fprintf(&b, "      %s: %q\n", k, v)
			}
		}
	}
	return b.String()
}

// inconclusiveVerdict is a backend answer the convergence engine reads as "no
// verdict": binary missing, key missing, quota refusal, timeout,
// unauthenticated and malformed answers all arrive at the seam in this shape.
func inconclusiveVerdict(summary string) ReviewOutput {
	return ReviewOutput{Verdict: VerdictInconclusive, Summary: summary, Findings: []Finding{}, NextSteps: []string{}}
}

// planCall is one audit_multi call over a root, returning the recording stub,
// the raw tool result and the decoded JSON result.
func planCall(t *testing.T, root string, extra map[string]any, verdictBy map[string]ReviewOutput) (*recordingCallerMulti, *mcp.CallToolResult, map[string]any) {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	rc := &recordingCallerMulti{verdictBy: verdictBy}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	args := map[string]any{"project_root": root}
	for k, v := range extra {
		args[k] = v
	}
	res, err := callToolAuditMulti(t, nil, args)
	if err != nil {
		t.Fatalf("handleAuditMulti returned a Go error: %v", err)
	}
	var m map[string]any
	if !res.IsError {
		if err := json.Unmarshal([]byte(toolResultText(res)), &m); err != nil {
			t.Fatalf("decode audit_multi result: %v\n%s", err, toolResultText(res))
		}
	}
	return rc, res, m
}

// invoked returns the backends the stub recorded, sorted.
func invoked(rc *recordingCallerMulti) []string {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	out := make([]string, 0, len(rc.calls))
	for _, c := range rc.calls {
		out = append(out, c.backend)
	}
	sort.Strings(out)
	return out
}

// perBackendNames lists the backends the result carries a verdict entry for.
func perBackendNames(t *testing.T, m map[string]any) []string {
	t.Helper()
	list, _ := m["per_backend_verdicts"].([]any)
	out := make([]string, 0, len(list))
	for _, e := range list {
		entry, _ := e.(map[string]any)
		name, _ := entry["backend"].(string)
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func requireBackends(t *testing.T, label string, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s: backends invoked = %v, want %v", label, got, want)
	}
}

// TestAuditMulti_ConfigPlanFallback (AC-ACV-006): with no `gates` argument the
// handler applies the plan the tree's audit.model token assigns, and one row
// omits a key to show the omitted backend takes its gate from the plan too.
func TestAuditMulti_ConfigPlanFallback(t *testing.T) {
	cases := []struct {
		name    string
		model   string
		gates   map[string]any
		backend []string
	}{
		{"codex alone", config.AuditModelCodex, nil, []string{BackendCodex}},
		{"glm alone", config.AuditModelGLM, nil, []string{BackendGLM}},
		// D7': Claude explicit, codex and glm on the default profile, so the
		// fan-out is the one the engine always had.
		{"claude token keeps the default fan-out", config.AuditModelClaude, nil, []string{BackendClaude, BackendCodex, BackendGLM}},
		{"multi", config.AuditModelMulti, nil, []string{BackendClaude, BackendCodex, BackendGLM}},
		// The call names claude only; codex is omitted and the glm-alone plan
		// turns it off, where the engine default would have run it.
		{"omitted key takes the plan", config.AuditModelGLM, map[string]any{"claude": "off"}, []string{BackendGLM}},
		// The call names glm only; the codex plan keeps codex and turns claude off.
		{"omitted claude key takes the plan", config.AuditModelCodex, map[string]any{"glm": "advisory"}, []string{BackendCodex, BackendGLM}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newAuditMultiBaselineRoot(t, planWorkflowYAML(tc.model, nil))
			extra := map[string]any{}
			if tc.gates != nil {
				extra["gates"] = tc.gates
			}
			rc, res, m := planCall(t, root, extra, nil)
			if res.IsError {
				t.Fatalf("unexpected error result: %s", toolResultText(res))
			}
			requireBackends(t, tc.name, invoked(rc), tc.backend)
			if got := m["plan_source"]; got != "config" {
				t.Errorf("plan_source = %v, want \"config\"", got)
			}
		})
	}
}

// TestAuditMulti_ArgsWinOverConfig (AC-ACV-007): a gate the call supplies wins
// over the configuration for that backend.
func TestAuditMulti_ArgsWinOverConfig(t *testing.T) {
	t.Run("arguments turn configured backends off", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
		rc, res, m := planCall(t, root, map[string]any{
			"gates": map[string]any{"codex": "off", "glm": "off", "claude": "required"},
		}, nil)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		requireBackends(t, "args off", invoked(rc), []string{BackendClaude})
		requireBackends(t, "args off per_backend_verdicts", perBackendNames(t, m), []string{BackendClaude})
	})
	t.Run("argument required wins over a configured off", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, map[string]string{"codex": "off"}))
		rc, res, _ := planCall(t, root, map[string]any{"gates": map[string]any{"codex": "required"}}, nil)
		if res.IsError {
			t.Fatalf("unexpected error result: %s", toolResultText(res))
		}
		requireBackends(t, "arg required over config off", invoked(rc), []string{BackendClaude, BackendCodex, BackendGLM})
	})
}

// TestAuditMulti_ArgumentRequiredIsNotExplicit (design.md §D.2): a `required`
// the caller wrote in the call enforces nothing on a tree that configures
// nothing — today's behaviour, kept so existing callers do not change.
func TestAuditMulti_ArgumentRequiredIsNotExplicit(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, distributedAuditPinsYAML)
	_, res, m := planCall(t, root, map[string]any{
		"gates": map[string]any{"codex": "required"},
	}, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")})
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	if m["overall_verdict"] != "pass" {
		t.Errorf("overall_verdict = %v, want pass (an argument-sourced required is fail-open)", m["overall_verdict"])
	}
	if _, present := m["gate_unmet"]; present {
		t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
	}
	if _, present := m["plan_source"]; present {
		t.Errorf("plan_source = %v, want absent (nothing came from configuration)", m["plan_source"])
	}
}

// TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed (AC-ACV-009): a
// tree that writes only `model: multi` makes codex an explicit required gate, so
// a codex leg that cannot answer fails the verdict by name.
func TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed(t *testing.T) {
	causes := []string{
		"codex binary not found", "codex api key missing", "codex quota refused",
		"codex timed out", "codex unauthenticated", "codex answer malformed",
	}
	for _, cause := range causes {
		t.Run(cause, func(t *testing.T) {
			root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
			_, res, m := planCall(t, root, nil, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict(cause)})
			if res.IsError {
				t.Fatalf("unexpected error result: %s", toolResultText(res))
			}
			if m["overall_verdict"] != "fail" {
				t.Errorf("overall_verdict = %v, want fail", m["overall_verdict"])
			}
			if m["gate_unmet"] != BackendCodex {
				t.Errorf("gate_unmet = %v, want %q", m["gate_unmet"], BackendCodex)
			}
			if note, _ := m["residual_risk_note"].(string); !strings.Contains(note, BackendCodex) {
				t.Errorf("residual_risk_note = %q, want it to name codex", note)
			}
			if m["plan_source"] != "config" {
				t.Errorf("plan_source = %v, want \"config\"", m["plan_source"])
			}
			var codexVerdict string
			for _, e := range m["per_backend_verdicts"].([]any) {
				entry := e.(map[string]any)
				if entry["backend"] == BackendCodex {
					codexVerdict, _ = entry["verdict"].(string)
				}
			}
			if codexVerdict != VerdictInconclusive {
				t.Errorf("codex per_backend verdict = %q, want it to stay %q", codexVerdict, VerdictInconclusive)
			}
			failOpen, _ := m["fail_open_backends"].([]any)
			if len(failOpen) != 1 || failOpen[0] != BackendCodex {
				t.Errorf("fail_open_backends = %v, want [codex]", failOpen)
			}
		})
	}

	// The same stubs on a tree that configures nothing stay a pass (fail-open):
	// the unmet gate follows the configuration, not the backend's failure.
	t.Run("unconfigured sibling root still passes", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, "")
		_, _, m := planCall(t, root, nil, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")})
		if m["overall_verdict"] != "pass" {
			t.Errorf("overall_verdict = %v, want pass", m["overall_verdict"])
		}
		if _, present := m["gate_unmet"]; present {
			t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
		}
	})
}

// TestAuditMulti_ExplicitClaudeToken (decision D7', AC-ACV-006): `model: claude`
// written in the tree leaves codex and glm on the default profile, so the fan-out
// is the engine's own and a codex that cannot answer stays fail-open — while
// the result still says the plan came from configuration.
func TestAuditMulti_ExplicitClaudeToken(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelClaude, nil))
	rc, res, m := planCall(t, root, nil, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")})
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	requireBackends(t, "claude token", invoked(rc), []string{BackendClaude, BackendCodex, BackendGLM})
	if m["plan_source"] != "config" {
		t.Errorf("plan_source = %v, want \"config\"", m["plan_source"])
	}
	if m["overall_verdict"] != "pass" {
		t.Errorf("overall_verdict = %v, want pass (codex is the non-explicit default under the claude token)", m["overall_verdict"])
	}
	if _, present := m["gate_unmet"]; present {
		t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
	}
}

// TestAuditMulti_UnknownConfiguredTokenIsToolError (AC-ACV-003, handler part):
// an audit.model or audit.gates value outside the closed set is a tool error
// that names the key and value, and no backend is invoked.
func TestAuditMulti_UnknownConfiguredTokenIsToolError(t *testing.T) {
	cases := []struct {
		name, yaml, want string
	}{
		{"unknown token", planWorkflowYAML("grok", nil), `audit_model "grok" unknown (want one of claude|codex|glm|multi)`},
		{"case-sensitive token", planWorkflowYAML("Multi", nil), `audit_model "Multi" unknown`},
		{"unknown gate", planWorkflowYAML("", map[string]string{"codex": "requird"}), `audit.gates.codex "requird" unknown (want one of off|advisory|required)`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newAuditMultiBaselineRoot(t, tc.yaml)
			rc, res, _ := planCall(t, root, nil, nil)
			if !res.IsError {
				t.Fatalf("IsError = false, want a tool error; result: %s", toolResultText(res))
			}
			if text := toolResultText(res); !strings.Contains(text, tc.want) {
				t.Errorf("error text = %q, want it to contain %q", text, tc.want)
			}
			if got := invoked(rc); len(got) != 0 {
				t.Errorf("backends invoked = %v, want none after a configuration error", got)
			}
		})
	}

	// Surrounding whitespace is trimmed, not rejected.
	t.Run("padded token is accepted", func(t *testing.T) {
		root := newAuditMultiBaselineRoot(t, planWorkflowYAML(" multi ", nil))
		_, res, m := planCall(t, root, nil, nil)
		if res.IsError {
			t.Fatalf("padded token rejected: %s", toolResultText(res))
		}
		if m["plan_source"] != "config" {
			t.Errorf("plan_source = %v, want \"config\"", m["plan_source"])
		}
	})
}

// TestAuditMulti_PaddedConfiguredGateEnforcesLikeTheTrimmedValue pins the one
// reading the resolver gives a gate written with surrounding whitespace: before
// the resolver, ` required ` reached the explicit-gate check untrimmed and never
// matched `required`, so the unmet gate was silently not enforced.
func TestAuditMulti_PaddedConfiguredGateEnforcesLikeTheTrimmedValue(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, planWorkflowYAML("", map[string]string{"codex": " required "}))
	_, res, m := planCall(t, root, nil, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")})
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	if m["overall_verdict"] != "fail" || m["gate_unmet"] != BackendCodex {
		t.Errorf("overall_verdict = %v, gate_unmet = %v, want fail / codex", m["overall_verdict"], m["gate_unmet"])
	}
}

// TestAuditMulti_InvalidSuppliedGateIsNotAToolError pins the one deliberate
// asymmetry of REQ-ACV-003: a value the CALL supplies outside off|advisory|
// required never became an error (TestAuditMulti_NoHardErrorPath_AC_AMM_024
// constrains the hard-error surface to an unusable project_root and, now, an
// invalid CONFIGURATION). The handler reads it as "not supplied" for that
// backend, so the tree's plan — or the default — decides the gate.
func TestAuditMulti_InvalidSuppliedGateIsNotAToolError(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, "")
	rc, res, m := planCall(t, root, map[string]any{
		"gates": map[string]any{"codex": "requird", "glm": 7},
	}, nil)
	if res.IsError {
		t.Fatalf("invalid supplied gate produced a tool error: %s", toolResultText(res))
	}
	requireBackends(t, "invalid supplied gates", invoked(rc), []string{BackendClaude, BackendCodex, BackendGLM})
	gate := map[string]string{}
	for _, e := range m["per_backend_verdicts"].([]any) {
		entry := e.(map[string]any)
		gate[entry["backend"].(string)], _ = entry["gate"].(string)
	}
	want := map[string]string{BackendClaude: "required", BackendCodex: "required", BackendGLM: "advisory"}
	for backend, g := range want {
		if gate[backend] != g {
			t.Errorf("%s gate = %q, want the default %q (an invalid supplied value reads as not supplied)", backend, gate[backend], g)
		}
	}
	if _, present := m["plan_source"]; present {
		t.Errorf("plan_source = %v, want absent", m["plan_source"])
	}
}

// TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate (AC-ACV-010): under
// `model: multi` the codex receipt predicate fires, so the result carries
// audit_receipt even when the gate was unmet.
func TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate(t *testing.T) {
	root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))
	_, res, m := planCall(t, root, nil, map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")})
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	if m["gate_unmet"] != BackendCodex {
		t.Fatalf("gate_unmet = %v, want codex (precondition)", m["gate_unmet"])
	}
	if id, _ := m["audit_receipt"].(string); id == "" {
		t.Errorf("audit_receipt is empty; a model: multi tree must record a receipt on an unmet gate")
	}
}

// TestCodexAudit_PaddedRequiredGateFailsClosed: the single-backend surface reads
// a padded `required` like the trimmed value, as audit_multi does. The reading
// used to be exact (a padded value was a non-required state with a golden of its
// own); the resolver trims gate values on every surface now (EC-1).
func TestCodexAudit_PaddedRequiredGateFailsClosed(t *testing.T) {
	root := newProbeProject(t, "SPEC-ACV-PADDED-GATE")
	writeRawCodexGateWorkflow(t, root, "workflow:\n  audit:\n    gates:\n      codex: \"required \"\n")
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	if v, _ := got["verdict"].(string); v != "fail" {
		t.Errorf("verdict = %q, want fail", v)
	}
	if g, _ := got["gate_unmet"].(string); g == "" {
		t.Error("gate_unmet is empty — a padded required gate with no verdict must be recorded as unmet")
	}
}

// TestCodexAudit_ModelCodexUnmetGateFails (AC-ACV-010): the single-backend
// codex_audit reads the same plan, so `model: codex` makes a codex that cannot
// answer an unmet gate.
func TestCodexAudit_ModelCodexUnmetGateFails(t *testing.T) {
	root := newProbeProject(t, "SPEC-ACV-MODEL-CODEX")
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(planWorkflowYAML(config.AuditModelCodex, nil)), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError — fail-open stays a structured result")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != "fail" {
		t.Errorf("verdict = %q, want fail (model: codex makes codex an explicit required gate)", v)
	}
	if g, _ := got["gate_unmet"].(string); g == "" {
		t.Error("gate_unmet is empty — a required codex gate that produced no verdict must be recorded as unmet")
	}
}

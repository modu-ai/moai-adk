// Package cli — baseline characterization of `audit_multi` with no `gates`
// argument (SPEC-AUDIT-MODEL-CONVERGE-001 M1, AC-ACV-008).
//
// WHY THIS FILE EXISTS. The SPEC claims that a call carrying no configuration
// and no arguments, or a configuration carrying only the three backend pins,
// returns the same bytes after the change as before it. A claim of that shape is
// only witnessable when the "before" artifact lands in a commit that precedes the
// change, so this file and testdata/audit_multi_default.golden.json are the
// baseline-first commit: they were generated from the pre-change handler and must
// pass on that tree without any production-code edit.
//
// The result is compared as raw JSON members, not as a decoded
// ConvergenceResult: a member added by a later milestone must show up here as a
// diff even when the struct that decodes it did not exist when this was written.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// updateAuditMultiDefaultGolden regenerates the golden. Set
// UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN=1 to merge the cases a run covers into the
// golden file; it must only ever be run against the pre-change handler.
var updateAuditMultiDefaultGolden = os.Getenv("UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN") == "1"

const auditMultiDefaultGoldenPath = "testdata/audit_multi_default.golden.json"

// distributedAuditPinsYAML has the shape of the `audit` block of the distributed
// workflow.yaml (internal/template/templates/.moai/config/sections/workflow.yaml,
// the three backend pins and nothing else): no `model` axis and no `gates` key.
const distributedAuditPinsYAML = `workflow:
    audit:
        claude:
            model: claude-opus-5-5
            effort: high
        codex:
            model: gpt-6.1-sol
            effort: high
        glm:
            model: glm-5.3
            effort: max
`

// newAuditMultiBaselineRoot returns a canonical project root that carries a
// .moai directory (validateProjectRoot requires one) and, when workflowYAML is
// non-empty, that body as .moai/config/sections/workflow.yaml.
func newAuditMultiBaselineRoot(t *testing.T, workflowYAML string) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if workflowYAML != "" {
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(workflowYAML), 0o644); err != nil {
			t.Fatalf("WriteFile workflow.yaml: %v", err)
		}
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", root, err)
	}
	return canonical
}

// auditMultiBaselineCases are the two backend situations covered per root: every
// backend passes, and the codex backend cannot answer. The second must stay a
// pass (fail-open) when no `gates` argument and no explicit gate is present.
func auditMultiBaselineCases() map[string]map[string]ReviewOutput {
	codexInconclusive := ReviewOutput{
		Verdict: VerdictInconclusive, Summary: "codex unavailable",
		Findings: []Finding{}, NextSteps: []string{},
	}
	return map[string]map[string]ReviewOutput{
		"all-pass":           nil,
		"codex-inconclusive": {BackendCodex: codexInconclusive},
	}
}

// runAuditMultiBaselineCase drives handleAuditMulti over root with stubbed
// backends and NO `gates` argument, and returns the normalized result bytes.
// Only fields that depend on the machine or the temp path are normalized, here
// and not in the golden: build_commit/build_lag (the binary's identity) and
// tree_root (asserted equal to root first, then dropped).
func runAuditMultiBaselineCase(t *testing.T, root string, verdictBy map[string]ReviewOutput) []byte {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	rc := &recordingCallerMulti{verdictBy: verdictBy}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	res, err := callToolAuditMulti(t, nil, map[string]any{"project_root": root})
	if err != nil {
		t.Fatalf("handleAuditMulti returned Go error: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("handleAuditMulti returned an error result: %+v", res)
	}
	return normalizeAuditMultiBaseline(t, res, root)
}

func normalizeAuditMultiBaseline(t *testing.T, res *mcp.CallToolResult, root string) []byte {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(toolResultText(res)), &m); err != nil {
		t.Fatalf("decode audit_multi result: %v\n%s", err, toolResultText(res))
	}
	if got := m["tree_root"]; got != root {
		t.Fatalf("tree_root = %v, want %q", got, root)
	}
	for _, k := range []string{"build_commit", "build_lag", "tree_root"} {
		delete(m, k)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal normalized result: %v", err)
	}
	return append(out, '\n')
}

// compareAuditMultiGolden compares got with the golden entry for name, or, under
// the update toggle, merges got into the golden file.
func compareAuditMultiGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	golden := map[string]json.RawMessage{}
	raw, err := os.ReadFile(auditMultiDefaultGoldenPath)
	switch {
	case err == nil:
		if err := json.Unmarshal(raw, &golden); err != nil {
			t.Fatalf("decode %s: %v", auditMultiDefaultGoldenPath, err)
		}
	case !os.IsNotExist(err) || !updateAuditMultiDefaultGolden:
		t.Fatalf("read %s: %v (UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN=1 regenerates, pre-change tree only)", auditMultiDefaultGoldenPath, err)
	}

	if updateAuditMultiDefaultGolden {
		golden[name] = json.RawMessage(bytes.TrimSpace(got))
		out, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatalf("marshal golden: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(auditMultiDefaultGoldenPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(auditMultiDefaultGoldenPath, append(out, '\n'), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	entry, ok := golden[name]
	if !ok {
		t.Fatalf("golden has no entry %q", name)
	}
	var want bytes.Buffer
	if err := json.Indent(&want, entry, "", "  "); err != nil {
		t.Fatalf("indent golden entry %q: %v", name, err)
	}
	want.WriteByte('\n')
	if !bytes.Equal(got, want.Bytes()) {
		t.Errorf("audit_multi result drifted for %q\ngot:\n%s\nwant:\n%s", name, got, want.Bytes())
	}
}

// assertNoPlanMembers checks the two members a later milestone adds only when a
// plan is in force: neither may appear for a call with no configuration, or with
// pins only, and no gate may be reported unmet.
func assertNoPlanMembers(t *testing.T, name string, got []byte) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	for _, k := range []string{"plan_source", "gate_unmet"} {
		if _, present := m[k]; present {
			t.Errorf("%s: result carries %q, want it absent", name, k)
		}
	}
	if m["overall_verdict"] != "pass" {
		t.Errorf("%s: overall_verdict = %v, want pass (fail-open)", name, m["overall_verdict"])
	}
}

// assertDefaultAuditPlan checks the last clause of AC-ACV-008: the plan resolved
// for the root, from the audit values its workflow.yaml carries and no caller
// gates, is the distributed default plan with every entry not explicit. The input
// is the RAW section (loadWorkflowAuditSection), never a default-merged
// configuration — a merged read would give the pins-only root `model: claude`.
func assertDefaultAuditPlan(t *testing.T, rootName, workflowYAML string) {
	t.Helper()
	root := newAuditMultiBaselineRoot(t, workflowYAML)
	audit, err := loadWorkflowAuditSection(root)
	if err != nil {
		t.Fatalf("%s: loadWorkflowAuditSection: %v", rootName, err)
	}
	plan, err := config.ResolveAuditPlan(audit, config.AuditGates{})
	if err != nil {
		t.Fatalf("%s: ResolveAuditPlan: %v", rootName, err)
	}
	wantGates := map[string]string{
		"claude": config.AuditGateRequired,
		"codex":  config.AuditGateRequired,
		"glm":    config.AuditGateAdvisory,
	}
	// Positive control: three entries, so the loop below cannot pass vacuously.
	if len(plan.Backends) != len(wantGates) {
		t.Fatalf("%s: plan has %d entries, want %d: %+v", rootName, len(plan.Backends), len(wantGates), plan.Backends)
	}
	for _, e := range plan.Backends {
		if e.Gate != wantGates[e.Backend] || e.Source != config.AuditPlanSourceDefault || e.Explicit {
			t.Errorf("%s: %s = %+v, want gate %q, source %q, explicit false",
				rootName, e.Backend, e, wantGates[e.Backend], config.AuditPlanSourceDefault)
		}
	}
	if plan.Model != "" || plan.ModelSource() != config.AuditPlanSourceDefault || plan.FromConfig() {
		t.Errorf("%s: model %q (%s), FromConfig %v, want no model token and nothing from config",
			rootName, plan.Model, plan.ModelSource(), plan.FromConfig())
	}
	if got := plan.ExplicitGates(); got != (config.AuditGates{}) {
		t.Errorf("%s: ExplicitGates() = %+v, want none (no explicit gate means fail-open stays)", rootName, got)
	}
}

func runAuditMultiBaselineRoot(t *testing.T, rootName, workflowYAML string) {
	assertDefaultAuditPlan(t, rootName, workflowYAML)
	for caseName, verdictBy := range auditMultiBaselineCases() {
		t.Run(caseName, func(t *testing.T) {
			root := newAuditMultiBaselineRoot(t, workflowYAML)
			got := runAuditMultiBaselineCase(t, root, verdictBy)
			name := rootName + "/" + caseName
			assertNoPlanMembers(t, name, got)
			compareAuditMultiGolden(t, name, got)
		})
	}
}

// TestAuditMulti_NoConfigNoArgs_ByteIdentical pins the result for a project root
// with no workflow.yaml and a call with no `gates` (AC-ACV-008).
func TestAuditMulti_NoConfigNoArgs_ByteIdentical(t *testing.T) {
	runAuditMultiBaselineRoot(t, "no-config", "")
}

// TestAuditMulti_PinsOnlyConfig_ByteIdentical pins the result for the distributed
// shape: a workflow.yaml whose audit block carries only the three backend pins
// (AC-ACV-008). A resolver built on a default-merged configuration would give
// this root `model: claude` and move these bytes.
func TestAuditMulti_PinsOnlyConfig_ByteIdentical(t *testing.T) {
	runAuditMultiBaselineRoot(t, "pins-only", distributedAuditPinsYAML)
}

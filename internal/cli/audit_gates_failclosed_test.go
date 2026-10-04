// Package cli — SPEC-AUDIT-CEILING-002 M3 (REQ-ACR-006 / AC-ACR-008/011, R4):
// the audit-gates resolution path keeps its error distinct from an empty
// not-configured result end to end, and every inventoried caller site
// surfaces the error instead of assuming an absent configuration.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeTreeWorkflow seeds a fresh temp tree whose workflow.yaml holds body and
// returns the root. A tree carrying .moai/config/sections is never
// config-orphaned, so resolveAuditGates reads its own file.
func writeTreeWorkflow(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// breakTreeWorkflow makes an existing tree's workflow.yaml unparseable.
func breakTreeWorkflow(t *testing.T, root string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte("workflow: [not: a: mapping\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

const unparseableWorkflow = "workflow: [not: a: mapping\n"

// TestWorkflowAuditPinsErrorNotFolded — AC-ACR-008's pins half: an unreadable
// and an unparseable workflow.yaml each return an error (never a folded zero
// configuration), while an absent file still reads as absent — the legitimate
// not-configured case keeps its meaning.
func TestWorkflowAuditPinsErrorNotFolded(t *testing.T) {
	t.Parallel()

	t.Run("unparseable workflow.yaml errors", func(t *testing.T) {
		t.Parallel()
		audit, err := workflowAuditPins(writeTreeWorkflow(t, unparseableWorkflow))
		if err == nil {
			t.Fatal("an unparseable workflow.yaml must return an error, not a folded zero configuration")
		}
		if audit.Model != "" || audit.Codex.Model != "" || audit.GLM.Model != "" {
			t.Errorf("error case folded pins into %+v", audit)
		}
	})

	t.Run("unreadable workflow.yaml errors", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// workflow.yaml as a directory: stat succeeds, the read fails.
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := workflowAuditPins(dir); err == nil {
			t.Fatal("an unreadable workflow.yaml must return an error")
		}
	})

	t.Run("absent file still reads as absent", func(t *testing.T) {
		t.Parallel()
		audit, err := workflowAuditPins(t.TempDir())
		if err != nil {
			t.Fatalf("an absent workflow.yaml is not-configured, not an error: %v", err)
		}
		if audit.Model != "" || audit.Codex.Model != "" || audit.GLM.Model != "" {
			t.Errorf("absent file yields %+v, want zero", audit)
		}
	})
}

// TestResolveAuditGatesConfigErrorDistinct — AC-ACR-008's resolver half: both
// error classes (a pins-loader read/parse error, an audit-plan resolver
// rejection) stay distinct from the empty not-configured result.
func TestResolveAuditGatesConfigErrorDistinct(t *testing.T) {
	t.Parallel()

	t.Run("pins-loader error class stays an error", func(t *testing.T) {
		t.Parallel()
		gates, _, err := resolveAuditGates(writeTreeWorkflow(t, unparseableWorkflow))
		if err == nil {
			t.Fatalf("resolver folded a pins-loader error into gates %+v", gates)
		}
	})

	t.Run("audit-plan resolver error class stays an error", func(t *testing.T) {
		t.Parallel()
		gates, _, err := resolveAuditGates(writeTreeWorkflow(t, "workflow:\n  audit:\n    model: bogus-backend\n"))
		if err == nil {
			t.Fatalf("resolver folded an audit-plan rejection into gates %+v", gates)
		}
	})

	t.Run("absent file stays the legitimate not-configured result", func(t *testing.T) {
		t.Parallel()
		root := newProbeProject(t, "SPEC-GATEABS-011")
		gates, _, err := resolveAuditGates(root)
		if err != nil {
			t.Fatalf("an absent workflow.yaml is not-configured, not an error: %v", err)
		}
		if gates.Codex != "" || gates.Claude != "" || gates.GLM != "" {
			t.Errorf("absent file yields %+v, want empty gates", gates)
		}
	})
}

// TestWorktreeRootSurfacesGateError — AC-ACR-011's tool surface: the
// codex_audit tool (the root-accepting audit surface whose result carries the
// gate read) reports the configuration error as a fail verdict naming it,
// never as an absent configuration.
func TestWorktreeRootSurfacesGateError(t *testing.T) {
	root := newProbeProject(t, "SPEC-WTRGATE-011")
	breakTreeWorkflow(t, root)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatal("the surfacing is a structured verdict, not a tool error")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != "fail" {
		t.Errorf("verdict = %q, want fail — an unreadable audit configuration must not read as an absent one", v)
	}
	if g, _ := got["gate_unmet"].(string); !strings.Contains(g, "workflow.yaml") {
		t.Errorf("gate_unmet = %q, want it to name the unreadable configuration", g)
	}
}

// TestGateErrorPropagatesToCallerSites — AC-ACR-011's R4 arm: the propagation
// claim is verified PER SITE, one subtest per inventoried caller of plan file
// 18 (LEDGER-ACR-N's six sites outside the two repaired surfaces).
func TestGateErrorPropagatesToCallerSites(t *testing.T) {
	t.Run("mcp_claude.go resolveClaudeAuditModelEffort", func(t *testing.T) {
		_, err := resolveClaudeAuditModelEffort(writeTreeWorkflow(t, unparseableWorkflow), "", "")
		if err == nil || !strings.Contains(err.Error(), "workflow.yaml") {
			t.Errorf("err = %v, want the pins read error surfaced", err)
		}
	})

	t.Run("codex_audit_launch.go prepareCodexAudit", func(t *testing.T) {
		repo := newAuditRepo(t)
		installFakeCodex(t)
		breakTreeWorkflow(t, repo.a1)

		r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, Root: repo.a1})
		if r.res.ExitCode == 0 {
			t.Fatal("a launch whose audit configuration cannot be read must refuse")
		}
		if !strings.Contains(r.stderr, "workflow.audit pins unreadable") || !strings.Contains(r.stderr, "workflow.yaml") {
			t.Errorf("stderr = %q, want the pins error surfaced", r.stderr)
		}
	})

	t.Run("mcp_glm.go resolveGLMAuditModelEffort", func(t *testing.T) {
		_, err := resolveGLMAuditModelEffort(writeTreeWorkflow(t, unparseableWorkflow))
		if err == nil || !strings.Contains(err.Error(), "workflow.yaml") {
			t.Errorf("err = %v, want the pins read error surfaced", err)
		}
	})

	t.Run("mcp_codex.go resolveCodexAuditModelEffort", func(t *testing.T) {
		_, err := resolveCodexAuditModelEffort(map[string]any{"cwd": writeTreeWorkflow(t, unparseableWorkflow)})
		if err == nil || !strings.Contains(err.Error(), "workflow.yaml") {
			t.Errorf("err = %v, want the pins read error surfaced", err)
		}
	})

	t.Run("mcp_convergence.go workflowAuditGates via runMultiAudit", func(t *testing.T) {
		prev := backendCall
		backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
			return ReviewOutput{Verdict: "pass", Summary: backend, Findings: []Finding{}, NextSteps: []string{}}
		}
		t.Cleanup(func() { backendCall = prev })

		r := runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", MultiAuditConfig{
			Gates: config.AuditGates{
				Claude: config.AuditGateRequired,
				Codex:  config.AuditGateRequired,
				GLM:    config.AuditGateOff,
			},
			ProjectRoot: writeTreeWorkflow(t, unparseableWorkflow),
		}, nil)
		if !strings.Contains(r.GateUnmet, "workflow.yaml") {
			t.Errorf("gate_unmet = %q, want the unreadable configuration surfaced", r.GateUnmet)
		}
		// CR2-P2-2 (card-review r2): the gate lookup error must promote the
		// OVERALL verdict to fail — every backend passed, but a gate posture
		// that cannot be read is not an absent one, and pass must not survive
		// on a note alone.
		if r.OverallVerdict != overallVerdictFail {
			t.Errorf("overall_verdict = %q with a gate lookup error and every backend passing, want %q", r.OverallVerdict, overallVerdictFail)
		}
	})

	t.Run("mcp_codex.go applyGateUnmet", func(t *testing.T) {
		out := applyGateUnmet(inconclusiveReview("probe cause"), writeTreeWorkflow(t, unparseableWorkflow))
		if out.Verdict != "fail" {
			t.Errorf("verdict = %q, want fail — a gate read that errors must not read as unconfigured", out.Verdict)
		}
		if !strings.Contains(out.GateUnmet, "workflow.yaml") {
			t.Errorf("gate_unmet = %q, want it to name the unreadable configuration", out.GateUnmet)
		}
	})
}

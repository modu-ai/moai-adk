package web

// mcp_audit_surface_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 /
// AC-MCP-021). Verifies the audit selection surfaces in the web console
// schema, AND that the web console does NOT fork the audit interpreter — it
// reuses the M3 typed config (config.AuditConfig). It carries no per-agent
// model/effort resolver at all: subagents inherit the main session's model and
// effort (SPEC-AGENT-MODEL-INHERIT-001).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestSchemaSurfaces_AuditSelection verifies AC-MCP-021: the audit_model +
// per-auditor audit_gate selection surface as schema fields in the Workflow
// section (so the web console renders them via the same schema-driven form +
// yamlpatch seam worktree/branch_guard use). The fields reuse the M3 typed
// config yaml paths (workflow.audit.model / .gates.*) — the identical
// interpreter.
func TestSchemaSurfaces_AuditSelection(t *testing.T) {
	wantNames := map[string]bool{
		"workflow.audit.model":        true,
		"workflow.audit.gates.claude": true,
		"workflow.audit.gates.codex":  true,
		"workflow.audit.gates.glm":    true,
	}
	got := map[string]bool{}
	for _, f := range settings.AllFields() {
		if f.Section == settings.SectionWorkflow {
			got[f.Name] = true
		}
	}
	for name := range wantNames {
		if !got[name] {
			t.Errorf("workflow schema field %q missing (AC-MCP-021 — audit selection must surface in the web console)", name)
		}
	}
}

// TestWebConsole_AuditNoForkedInterpreter verifies the "identical interpreter"
// clause of AC-MCP-021: the web console MUST NOT define its own audit-backend
// resolver. The audit selection surfaces via the schema-driven form (which
// reads/writes the M3 config.AuditConfig yaml paths); the model/effort
// resolution for the audit backends stays in the MCP handlers (which call the
// shared SSOT). A second ResolveAuditPlan / audit-model resolver in
// internal/web would be the fork this test forbids.
//
// The sentinel is only meaningful while it names a live symbol:
// TestWebConsole_AuditSentinelExists is the positive control that fails when
// config.ResolveAuditPlan is renamed or removed, so this guard cannot go
// vacuous.
func TestWebConsole_AuditNoForkedInterpreter(t *testing.T) {
	// Collect every non-test .go file in internal/web.
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if strings.Contains(string(data), sentinel) {
			t.Errorf("internal/web/%s defines/references %q — the web console must NOT fork the audit interpreter (AC-MCP-021); it reads the M3 config.AuditConfig via the schema seam", f, sentinel)
		}
	}
}

// sentinel is the live audit-model resolver symbol the web console must not
// fork: config.ResolveAuditPlan, the single consumer of the workflow.audit.model
// token. Both the guard and its positive control read this one constant.
const sentinel = "ResolveAuditPlan"

// declaresFunc reports how many non-test .go files in dir declare
// `func <name>(`. It returns an error when dir holds no non-test source at
// all, so an empty sweep is never read as a verdict.
func declaresFunc(dir, name string) (int, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return 0, err
	}
	var scanned, hits int
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return 0, err
		}
		scanned++
		if strings.Contains(string(data), "func "+name+"(") {
			hits++
		}
	}
	if scanned == 0 {
		return 0, os.ErrNotExist
	}
	return hits, nil
}

// TestWebConsole_AuditSentinelExists is the positive control for
// TestWebConsole_AuditNoForkedInterpreter: a guard whose sentinel names a
// deleted symbol passes vacuously, so the sentinel must be declared in the
// non-test source of internal/config. The second subtest proves the probe can
// report absence.
func TestWebConsole_AuditSentinelExists(t *testing.T) {
	t.Run("sentinel is declared in internal/config non-test source", func(t *testing.T) {
		hits, err := declaresFunc(filepath.Join("..", "config"), sentinel)
		if err != nil {
			t.Fatalf("scan internal/config: %v", err)
		}
		if hits == 0 {
			t.Errorf("func %s( is not declared in internal/config non-test source — the no-forked-interpreter guard names no live symbol and passes vacuously", sentinel)
		}
	})
	t.Run("probe reports an absent symbol as absent", func(t *testing.T) {
		hits, err := declaresFunc(filepath.Join("..", "config"), "NoSuchAuditResolverSymbol")
		if err != nil {
			t.Fatalf("scan internal/config: %v", err)
		}
		if hits != 0 {
			t.Errorf("probe found %d declarations of a symbol that does not exist — the control cannot fail", hits)
		}
	})
}

// TestWebConsole_NoPerAgentModelResolver verifies the web console neither
// defines nor calls a per-agent model/effort resolver: the agent-settings tab
// that consumed one is gone (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011), and no
// other surface may reintroduce per-agent assignment.
func TestWebConsole_NoPerAgentModelResolver(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	var scanned int
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		scanned++
		for _, sym := range []string{"ResolveAgentModelEffort", "ProfileMatrixAgents", "EffectiveProfile", "AgentOverrides"} {
			if strings.Contains(string(data), sym) {
				t.Errorf("internal/web/%s references %s — the web console assigns no per-agent model or effort", f, sym)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no internal/web source file — the guard read nothing")
	}
}

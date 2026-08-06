package web

// mcp_audit_surface_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 /
// AC-MCP-021). Verifies the audit selection surfaces in the web console
// schema, AND that the web console does NOT fork the audit interpreter — it
// reuses the M3 typed config (config.AuditConfig) + the shared model/effort
// SSOT (template.ResolveAgentModelEffort). No second definition, no reresolver.

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
// shared SSOT). A second activeAuditBackend / audit-model resolver in
// internal/web would be the fork this test forbids.
func TestWebConsole_AuditNoForkedInterpreter(t *testing.T) {
	// Collect every non-test .go file in internal/web.
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	const sentinel = "activeAuditBackend" // the M3 resolver symbol in internal/cli
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

// TestWebConsole_ResolveAgentModelEffortSSOTShared verifies the ResolveAgentModelEffort
// SSOT clause of AC-MCP-021: the model/effort interpreter is defined ONCE (in
// internal/template) and the web console IMPORTS it — it does NOT redefine a
// second resolver. This is the "ResolveAgentModelEffort SSOT respected" guard.
func TestWebConsole_ResolveAgentModelEffortSSOTShared(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	const def = "func ResolveAgentModelEffort"
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if strings.Contains(string(data), def) {
			t.Errorf("internal/web/%s redefines %q — the SSOT lives in internal/template; the web console must import, not redefine (AC-MCP-021)", f, def)
		}
	}
}
